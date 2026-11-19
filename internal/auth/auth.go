package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/TechX-Corp/xoradlc-plugin/internal/cli"
	"github.com/TechX-Corp/xoradlc-plugin/internal/config"
	"github.com/TechX-Corp/xoradlc-plugin/internal/fsutil"
	"github.com/TechX-Corp/xoradlc-plugin/internal/tui"
)

// AuthTokens is the on-disk OAuth token set.
type AuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresAt    int64  `json:"expires_at"`
}

var tokenMu sync.Mutex

// NotLoggedInMsg is the user-facing hint when OAuth session is missing.
const NotLoggedInMsg = "Not logged in. Run xdlc login first."

func authPath() string {
	return filepath.Join(config.XDLC(), "auth.json")
}

func loadRaw() (map[string]any, error) {
	var raw map[string]any
	if err := fsutil.LoadJSON(authPath(), &raw); err != nil {
		if os.IsNotExist(err) {
			return map[string]any{}, nil
		}
		return nil, err
	}
	if raw == nil {
		return map[string]any{}, nil
	}
	if config.Windows() {
		if enc, ok := raw["dpapi"].(string); ok && enc != "" {
			cipher, err := base64.StdEncoding.DecodeString(enc)
			if err != nil {
				return nil, err
			}
			plain, err := dpapi(cipher, false)
			if err != nil {
				return nil, err
			}
			if err := json.Unmarshal(plain, &raw); err != nil {
				return nil, err
			}
		}
	}
	return raw, nil
}

// Tokens returns the login token dict, or nil. Migrates legacy auth.json keyed by env (prod/local).
func Tokens() (*AuthTokens, error) {
	raw, err := loadRaw()
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, nil
	}
	if _, ok := raw["access_token"]; ok {
		b, _ := json.Marshal(raw)
		var t AuthTokens
		if err := json.Unmarshal(b, &t); err != nil {
			return nil, err
		}
		if t.AccessToken != "" {
			return &t, nil
		}
	}
	if prod, ok := raw["prod"].(map[string]any); ok {
		b, _ := json.Marshal(prod)
		var t AuthTokens
		if err := json.Unmarshal(b, &t); err == nil && t.AccessToken != "" {
			return &t, nil
		}
	}
	return nil, nil
}

// SaveTokens writes tokens (nil clears the file).
func SaveTokens(t *AuthTokens) error {
	var data any
	if t != nil {
		data = t
	} else {
		data = map[string]any{}
	}
	if config.Windows() {
		b, err := json.Marshal(data)
		if err != nil {
			return err
		}
		enc, err := dpapi(b, true)
		if err != nil {
			return err
		}
		data = map[string]any{"dpapi": base64.StdEncoding.EncodeToString(enc)}
	}
	return fsutil.SaveJSON(authPath(), data, true)
}

func store(granted map[string]any) (*AuthTokens, error) {
	expiresIn := int64(3600)
	if v, ok := granted["expires_in"]; ok {
		switch x := v.(type) {
		case float64:
			expiresIn = int64(x)
		case int:
			expiresIn = int64(x)
		}
	}
	t := &AuthTokens{
		AccessToken: fmt.Sprint(granted["access_token"]),
		ExpiresAt:   time.Now().Unix() + expiresIn,
	}
	if v, ok := granted["refresh_token"].(string); ok {
		t.RefreshToken = v
	}
	if err := SaveTokens(t); err != nil {
		return nil, err
	}
	return t, nil
}

// AccessToken returns a valid access token (auto-refresh), or "" when logged out / session expired.
func AccessToken(forceRefresh bool) (string, error) {
	tokenMu.Lock()
	defer tokenMu.Unlock()

	t, err := Tokens()
	if err != nil {
		return "", err
	}
	if t == nil {
		return "", nil
	}
	if !forceRefresh && t.ExpiresAt-60 > time.Now().Unix() {
		return t.AccessToken, nil
	}
	if t.RefreshToken == "" {
		return "", nil
	}
	status, out := oauthRequest("POST", config.BaseURL()+"/oauth/token", map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": t.RefreshToken,
		"client_id":     config.ClientID,
	}, "")
	if status == 200 {
		if m, ok := out.(map[string]any); ok && m["access_token"] != nil {
			st, err := store(m)
			if err != nil {
				return "", err
			}
			return st.AccessToken, nil
		}
	}
	if status == 400 || status == 401 {
		_ = SaveTokens(nil)
	}
	return "", nil
}

// UserInfo is the OAuth userinfo payload.
type UserInfo map[string]any

// Whoami returns account info and a reason string when unavailable.
func Whoami() (UserInfo, string) {
	token, err := AccessToken(false)
	if err != nil {
		return nil, NotLoggedInMsg
	}
	if token == "" {
		return nil, NotLoggedInMsg
	}
	status, out := oauthRequest("GET", config.BaseURL()+"/oauth/userinfo", nil, token)
	if status == 401 {
		token, err = AccessToken(true)
		if err != nil {
			return nil, NotLoggedInMsg
		}
		if token != "" {
			status, out = oauthRequest("GET", config.BaseURL()+"/oauth/userinfo", nil, token)
		}
	}
	if status == 200 {
		if m, ok := out.(map[string]any); ok {
			return m, ""
		}
	}
	if status == 0 {
		return nil, fmt.Sprintf("could not connect to %s (%v)", config.BaseURL(), out)
	}
	if status == 401 {
		return nil, "session expired: run xdlc login"
	}
	return nil, fmt.Sprintf("HTTP error %d", status)
}

// AccountLine formats a signed-in user for display.
func AccountLine(u UserInfo) string {
	role := ""
	if r, ok := u["role"].(string); ok {
		role = r
	}
	name := fmt.Sprint(u["name"])
	user := fmt.Sprint(u["username"])
	return fmt.Sprintf("%s (%s) · role %s", name, user, tui.RoleLabel(role))
}

// Logout deletes stored tokens.
func Logout() error {
	t, err := Tokens()
	if err != nil {
		return err
	}
	if t == nil {
		return nil
	}
	return SaveTokens(nil)
}

// StoreGranted saves tokens from an OAuth grant response (login / refresh).
func StoreGranted(granted map[string]any) (*AuthTokens, error) {
	return store(granted)
}

// RequireLoggedIn fails with a friendly message when there is no session.
func RequireLoggedIn() error {
	u, why := Whoami()
	if u != nil {
		return nil
	}
	if why == "" {
		why = NotLoggedInMsg
	}
	return cli.Fail(why)
}

// FinishLogin validates post-login output (optional helper for cmd layer).
func RequireLoginOption(argv []string) error {
	if len(argv) > 0 {
		return cli.Usage("login takes no options.")
	}
	return nil
}