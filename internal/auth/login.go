package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/TechX-Corp/xoradlc-plugin/internal/cli"
	"github.com/TechX-Corp/xoradlc-plugin/internal/config"
	"github.com/TechX-Corp/xoradlc-plugin/internal/tui"
)

// Login runs the browser PKCE OAuth flow.
func Login() error {
	if err := RequireLoginOption(nil); err != nil {
		return err
	}
	base := config.BaseURL()
	status, meta := oauthRequest("GET", base+"/.well-known/oauth-authorization-server", nil, "")
	if status != 200 {
		detail := meta
		if status != 0 {
			detail = "HTTP " + fmt.Sprint(status)
		}
		return cli.Fail(fmt.Sprintf("Could not load login metadata from %s (%v).", base, detail))
	}
	m, ok := meta.(map[string]any)
	if !ok {
		return cli.Fail(fmt.Sprintf("Could not load login metadata from %s (HTTP %d).", base, status))
	}
	verifier, err := tokenURLSafe(48)
	if err != nil {
		return err
	}
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	state, err := tokenURLSafe(16)
	if err != nil {
		return err
	}

	result := map[string]string{}
	mux := http.NewServeMux()
	srv := &http.Server{Handler: mux}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	port := ln.Addr().(*net.TCPAddr).Port
	redirect := fmt.Sprintf("http://127.0.0.1:%d/oauth/callback", port)

	mux.HandleFunc("/oauth/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		for k, vs := range q {
			if len(vs) > 0 {
				result[k] = vs[0]
			}
		}
		success := result["code"] != ""
		detail := result["error_description"]
		if detail == "" {
			detail = result["error"]
		}
		body := oauthCallbackPage(success, detail)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Length", fmt.Sprint(len(body)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	})

	go func() { _ = srv.Serve(ln) }()
	defer func() {
		_ = srv.Close()
	}()

	authURL, _ := url.Parse(fmt.Sprint(m["authorization_endpoint"]))
	q := authURL.Query()
	q.Set("response_type", "code")
	q.Set("client_id", config.ClientID)
	q.Set("redirect_uri", redirect)
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")
	q.Set("state", state)
	authURL.RawQuery = q.Encode()
	fullURL := authURL.String()

	fmt.Printf("Sign in to %s. Open this page if the browser does not open:\n%s\n", tui.Em("Xora DLC"), fullURL)
	if os.Getenv("XDLC_NO_BROWSER") == "" {
		openBrowser(fullURL)
	}

	deadline := time.Now().Add(5 * time.Minute)
	for len(result) == 0 && time.Now().Before(deadline) {
		time.Sleep(200 * time.Millisecond)
	}
	if len(result) == 0 {
		return cli.Fail("Timed out after 5 minutes without completing sign-in. Run xdlc login again.")
	}
	if result["state"] != state {
		return cli.Fail("Login response state mismatch. Run xdlc login again.")
	}
	code := result["code"]
	if code == "" {
		msg := result["error_description"]
		if msg == "" {
			msg = result["error"]
		}
		return cli.Fail("Sign-in failed: " + msg)
	}

	tokenEndpoint := fmt.Sprint(m["token_endpoint"])
	status, out := oauthRequest("POST", tokenEndpoint, map[string]string{
		"grant_type":    "authorization_code",
		"code":          code,
		"redirect_uri":  redirect,
		"client_id":     config.ClientID,
		"code_verifier": verifier,
	}, "")
	if status != 200 {
		return cli.Fail(fmt.Sprintf("Token exchange failed: %v", out))
	}
	granted, ok := out.(map[string]any)
	if !ok || granted["access_token"] == nil {
		return cli.Fail(fmt.Sprintf("Token exchange failed: %v", out))
	}
	if _, err := store(granted); err != nil {
		return err
	}
	u, why := Whoami()
	if u != nil {
		fmt.Printf("%s %s\n", tui.Ok("Signed in:"), AccountLine(u))
	} else {
		fmt.Printf("%s (could not read account: %s)\n", tui.Ok("Signed in:"), why)
	}
	if u != nil {
		if r, ok := u["role"].(string); !ok || r == "" {
			fmt.Println(tui.Warn("Account has no role:") + tui.Dim(fmt.Sprintf(" ask a %s admin to assign a role (e.g. %s) before asking the agent.", tui.Em("Xora DLC"), tui.Em("DEV"))))
		}
	}
	fmt.Println(tui.Dim(fmt.Sprintf("Wrong account? Sign out of %s in the browser, run %s, then %s.",
		tui.Em("Xora DLC"), tui.Em("xdlc logout"), tui.Em("xdlc login"))))
	return nil
}

func tokenURLSafe(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func openBrowser(u string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", u)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", u)
	default:
		cmd = exec.Command("xdg-open", u)
	}
	_ = cmd.Start()
}

func oauthCallbackPage(success bool, detail string) []byte {
	// Minimal port of Python _oauth_callback_page (dark theme).
	var status, title, body, hint, badge string
	if success {
		status = "ok"
		title = "You're signed in"
		body = "Return to the terminal — you can close this tab."
		hint = "This window can be closed."
		badge = "Signed in"
	} else {
		status = "bad"
		title = "Sign-in didn't complete"
		if detail == "" {
			detail = "Something went wrong during sign-in."
		}
		body = html.EscapeString(detail)
		hint = "Close this tab and run xdlc login again in the terminal."
		badge = "Failed"
	}
	page := fmt.Sprintf(`<!doctype html>
<html lang="en" data-theme="dark">
<head><meta charset="utf-8"><title>Xora DLC</title></head>
<body>
<main data-status="%s">
<p class="brand">Xora DLC</p>
<span class="badge">%s</span>
<h1>%s</h1>
<p class="body">%s</p>
<p class="hint">%s</p>
</main>
</body></html>`, status, html.EscapeString(badge), html.EscapeString(title), body, html.EscapeString(hint))
	return []byte(page)
}

// LogoutCommand mirrors cmd_logout messaging.
func LogoutCommand(argv []string) error {
	if len(argv) > 0 {
		return cli.Usage("logout takes no options.")
	}
	t, err := Tokens()
	if err != nil {
		return err
	}
	if t == nil {
		fmt.Println("Not logged in.")
		return nil
	}
	if err := SaveTokens(nil); err != nil {
		return err
	}
	fmt.Printf("Logged out of %s. %s\n", tui.Em("Xora DLC"),
		tui.Dim(fmt.Sprintf("The agent needs %s again.", tui.Em("xdlc login"))))
	return nil
}

// DecodeTokens is a helper for tests.
func DecodeTokens(b []byte) (*AuthTokens, error) {
	var t AuthTokens
	if err := json.Unmarshal(b, &t); err != nil {
		return nil, err
	}
	return &t, nil
}