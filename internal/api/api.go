package api

import (
	"fmt"
	"strings"
	"time"

	"github.com/TechX-Corp/xoradlc-plugin/internal/auth"
	"github.com/TechX-Corp/xoradlc-plugin/internal/cli"
	"github.com/TechX-Corp/xoradlc-plugin/internal/config"
)

// Step is one in-progress work step from My Work.
type Step struct {
	WorkOrderID   string `json:"work_order_id"`
	StepID        string `json:"step_id"`
	WorkOrderName string `json:"work_order_name,omitempty"`
	StepName      string `json:"step_name,omitempty"`
}

const conventionAskPrompt = "Return the full coding convention that Developers must follow for this organization / " +
	"current work, from the knowledge base. Prefer the official coding convention document. " +
	"Include any automated rule IDs (e.g. SEC-01) if present. Plain text only."

// ConventionAskPrompt is the default AAP convention question.
var ConventionAskPrompt = conventionAskPrompt

// API calls the backend with the login token (auto-refresh once on 401).
// Returns (status, json) or (0, nil) when not logged in — status -1 means not logged in for callers.
func API(method, path string, body any, timeout time.Duration) (int, any) {
	token, err := auth.AccessToken(false)
	if err != nil || token == "" {
		return -1, nil
	}
	url := config.BaseURL() + path
	status, out := Request(method, url, RequestOpts{Body: body, Token: token, Timeout: timeout})
	if status == 401 {
		token, err = auth.AccessToken(true)
		if err == nil && token != "" {
			status, out = Request(method, url, RequestOpts{Body: body, Token: token, Timeout: timeout})
		}
	}
	return status, out
}

// AskAAP asks via POST /mcp/user tools/call xora_ask.
func AskAAP(question string, timeout time.Duration) (string, error) {
	if timeout == 0 {
		timeout = 120 * time.Second
	}
	token, err := auth.AccessToken(false)
	if err != nil {
		return "", err
	}
	if token == "" {
		return "", cli.Usage("Not logged in to Xora DLC: run xdlc login.")
	}
	if len(question) > 3800 {
		question = question[:3800]
	}
	reqBody := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "xora_ask",
			"arguments": map[string]any{
				"text": question,
			},
		},
	}
	url := config.BaseURL() + "/mcp/user"
	status, out := Request("POST", url, RequestOpts{Body: reqBody, Token: token, Timeout: timeout})
	if status == 401 {
		token, err = auth.AccessToken(true)
		if err != nil {
			return "", err
		}
		if token != "" {
			status, out = Request("POST", url, RequestOpts{Body: reqBody, Token: token, Timeout: timeout})
		}
	}
	if status == 0 {
		return "", cli.Fail(fmt.Sprintf("Could not connect to %s (%v).", config.BaseURL(), out))
	}
	m, ok := out.(map[string]any)
	if status != 200 || !ok {
		return "", cli.Fail(fmt.Sprintf("AAP ask failed (HTTP %d): %v", status, out))
	}
	result, _ := m["result"].(map[string]any)
	if result != nil {
		if isErr, _ := result["isError"].(bool); isErr {
			content, _ := result["content"].([]any)
			msg := "AAP error"
			if len(content) > 0 {
				if c0, ok := content[0].(map[string]any); ok {
					if t, ok := c0["text"].(string); ok && t != "" {
						msg = t
					}
				}
			}
			return "", cli.Fail(msg)
		}
		var texts []string
		if content, ok := result["content"].([]any); ok {
			for _, item := range content {
				c, ok := item.(map[string]any)
				if !ok {
					continue
				}
				if c["type"] == "text" {
					if t, ok := c["text"].(string); ok && t != "" {
						texts = append(texts, t)
					}
				}
			}
		}
		if len(texts) == 0 {
			return "", cli.Fail("AAP returned no convention text.")
		}
		return texts[0], nil
	}
	return "", cli.Fail(fmt.Sprintf("AAP ask failed (HTTP %d): %v", status, out))
}

// MySteps returns steps you are working on (My Work: steps_in_progress).
func MySteps() ([]Step, error) {
	status, out := API("GET", "/my-work", nil, 15*time.Second)
	if status == -1 {
		return nil, cli.Usage("Not logged in to Xora DLC: run xdlc login.")
	}
	if status == 0 {
		return nil, cli.Fail(fmt.Sprintf("Could not connect to %s (%v).", config.BaseURL(), out))
	}
	m, ok := out.(map[string]any)
	if status != 200 || !ok {
		return nil, cli.Fail(fmt.Sprintf("Could not load My Work (HTTP %d): %v", status, out))
	}
	raw, _ := m["steps_in_progress"].([]any)
	var steps []Step
	for _, item := range raw {
		sm, ok := item.(map[string]any)
		if !ok {
			continue
		}
		wo := fmt.Sprint(sm["work_order_id"])
		sid := fmt.Sprint(sm["step_id"])
		if wo == "" || sid == "" || wo == "<nil>" || sid == "<nil>" {
			continue
		}
		steps = append(steps, Step{
			WorkOrderID:   wo,
			StepID:        sid,
			WorkOrderName: fmt.Sprint(sm["work_order_name"]),
			StepName:      fmt.Sprint(sm["step_name"]),
		})
	}
	return steps, nil
}

// StepLabel formats a step for menus.
func StepLabel(s Step) string {
	name := s.WorkOrderName
	if name == "<nil>" {
		name = ""
	}
	stepName := s.StepName
	if stepName == "" || stepName == "<nil>" {
		stepName = s.StepID
	}
	return fmt.Sprintf("%s · %s · %s", s.WorkOrderID, name, stepName)
}

// LatestRelease returns the latest vX.Y.Z tag on GitHub, or "" when unavailable.
func LatestRelease() string {
	url := fmt.Sprintf("https://api.github.com/repos/%s/tags?per_page=100", config.Repo)
	status, out := Request("GET", url, RequestOpts{Timeout: 8 * time.Second})
	if status != 200 {
		return ""
	}
	list, ok := out.([]any)
	if !ok {
		return ""
	}
	type pair struct {
		ver  [3]int
		name string
	}
	var tags []pair
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		name := fmt.Sprint(m["name"])
		v := semver(name)
		if v == nil {
			continue
		}
		tags = append(tags, pair{ver: *v, name: name})
	}
	if len(tags) == 0 {
		return ""
	}
	best := tags[0]
	for _, t := range tags[1:] {
		if semverGreater(t.ver, best.ver) {
			best = t
		}
	}
	return best.name
}

func semver(tag string) *[3]int {
	tag = strings.TrimPrefix(tag, "v")
	parts := strings.Split(tag, ".")
	if len(parts) != 3 {
		return nil
	}
	var v [3]int
	for i, p := range parts {
		var n int
		if _, err := fmt.Sscanf(p, "%d", &n); err != nil {
			return nil
		}
		v[i] = n
	}
	return &v
}

func semverGreater(a, b [3]int) bool {
	for i := 0; i < 3; i++ {
		if a[i] != b[i] {
			return a[i] > b[i]
		}
	}
	return false
}