package kiro

import (
	"encoding/json"
	"os/exec"
	"strings"
)

// AgentEmail runs `kiro-cli whoami --format json` (or plain) and returns email, or nil.
func AgentEmail() *string {
	cmd := exec.Command("kiro-cli", "whoami", "--format", "json")
	out, err := cmd.Output()
	if err != nil {
		// Fallback plain text: "Email: …"
		cmd = exec.Command("kiro-cli", "whoami")
		out, err = cmd.Output()
		if err != nil {
			return nil
		}
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			lower := strings.ToLower(line)
			if !strings.HasPrefix(lower, "email:") {
				continue
			}
			email := strings.TrimSpace(line[strings.Index(line, ":")+1:])
			if email == "" {
				return nil
			}
			return &email
		}
		return nil
	}
	var doc map[string]any
	if err := json.Unmarshal(out, &doc); err != nil {
		return nil
	}
	email := findEmail(doc)
	if email == "" {
		return nil
	}
	return &email
}

func findEmail(v any) string {
	switch t := v.(type) {
	case map[string]any:
		for k, child := range t {
			if strings.EqualFold(k, "email") {
				if s, ok := child.(string); ok {
					return strings.TrimSpace(s)
				}
			}
			if s := findEmail(child); s != "" {
				return s
			}
		}
	case []any:
		for _, child := range t {
			if s := findEmail(child); s != "" {
				return s
			}
		}
	}
	return ""
}
