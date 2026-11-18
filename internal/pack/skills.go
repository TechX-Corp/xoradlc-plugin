package pack

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/TechX-Corp/xoradlc-plugin/internal/api"
	"github.com/TechX-Corp/xoradlc-plugin/internal/tui"
)

const skillAllowlistPrompt = "Return ONLY a JSON array of skills allowed for my Xora DLC role. " +
	`Each element must be {"name":"<skill-id>","description":"<one-line description>"}. ` +
	"No markdown fences, no commentary — JSON array only."

// Skill describes one hydrated SKILL.md entry.
type Skill struct {
	Name        string
	Description string
}

// SkillMD returns SKILL.md body for a skill name.
func SkillMD(name, description string) string {
	desc := strings.ReplaceAll(strings.TrimSpace(description), "\n", " ")
	if desc == "" {
		desc = name
	}
	return fmt.Sprintf("---\nname: %s\ndescription: %s\n---\n\n# %s\n\n"+
		"When the user invokes this skill, call MCP tool `xora_ask` with a clear request to run "+
		"or fulfill skill `%s` (include the skill name and the user's intent). "+
		"Let AAP decide how to execute.\n", name, desc, name, name)
}

// ParseSkillAllowlist parses AAP natural-language JSON array output.
func ParseSkillAllowlist(text string) ([]Skill, error) {
	s := strings.TrimSpace(text)
	if strings.Contains(s, "```") {
		parts := strings.Split(s, "```")
		for i, block := range parts {
			if i%2 == 1 {
				body := strings.TrimSpace(block)
				if strings.HasPrefix(strings.ToLower(body), "json") {
					body = strings.TrimLeft(body[4:], " \t")
				}
				s = body
				break
			}
		}
	}
	start := strings.Index(s, "[")
	end := strings.LastIndex(s, "]")
	if start < 0 || end <= start {
		return nil, fmt.Errorf("no JSON array")
	}
	var data []map[string]any
	if err := json.Unmarshal([]byte(s[start:end+1]), &data); err != nil {
		return nil, err
	}
	var out []Skill
	for _, item := range data {
		name := strings.TrimSpace(fmt.Sprint(item["name"]))
		if name == "" || name == "<nil>" {
			continue
		}
		desc := strings.TrimSpace(fmt.Sprint(item["description"]))
		if desc == "" || desc == "<nil>" {
			desc = name
		}
		out = append(out, Skill{Name: name, Description: desc})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("empty skill list")
	}
	return out, nil
}

// FetchSkillAllowlist loads skills from AAP; soft-fails to empty slice.
func FetchSkillAllowlist() []Skill {
	text, err := api.AskAAP(skillAllowlistPrompt, 90*time.Second)
	if err != nil {
		fmt.Println(tui.Warn("Skill allowlist:") + tui.Dim(fmt.Sprintf(" could not load from AAP (%v); continuing without SKILL.md files.", err)))
		return nil
	}
	skills, err := ParseSkillAllowlist(text)
	if err != nil {
		fmt.Println(tui.Warn("Skill allowlist:") + tui.Dim(fmt.Sprintf(" could not parse AAP response (%v); continuing without SKILL.md files.", err)))
		return nil
	}
	return skills
}