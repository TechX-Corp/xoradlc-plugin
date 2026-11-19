package claude

import (
	"encoding/json"
	"os/exec"
	"strings"
)

// AgentEmail runs `claude auth status` and returns email, or nil.
func AgentEmail() *string {
	cmd := exec.Command("claude", "auth", "status")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	var doc map[string]any
	if err := json.Unmarshal(out, &doc); err != nil {
		return nil
	}
	if logged, _ := doc["loggedIn"].(bool); !logged {
		return nil
	}
	email, _ := doc["email"].(string)
	email = strings.TrimSpace(email)
	if email == "" {
		return nil
	}
	return &email
}
