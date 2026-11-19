package installs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TechX-Corp/xoradlc-plugin/internal/cli"
	"github.com/TechX-Corp/xoradlc-plugin/internal/config"
	"github.com/TechX-Corp/xoradlc-plugin/internal/fsutil"
	"github.com/TechX-Corp/xoradlc-plugin/internal/pack"
	"github.com/TechX-Corp/xoradlc-plugin/internal/setup"
	"github.com/TechX-Corp/xoradlc-plugin/internal/tui"
)

// Record is one agent setup persisted in installs.json.
type Record struct {
	Agent          string            `json:"agent"`
	Scope          string            `json:"scope"`
	Path           string            `json:"path"`
	Version        string            `json:"version"`
	Role           string            `json:"role"`
	Files          map[string]string `json:"files"`
	McpJSON        string            `json:"mcp_json"`
	CreatedMcpJSON bool              `json:"created_mcp_json"`
	GitTracked     string            `json:"git_tracked,omitempty"`
	SettingsRel    string            `json:"settings_rel,omitempty"`
}

type file struct {
	Version  string   `json:"version"`
	Installs []Record `json:"installs"`
}

func path() string {
	return filepath.Join(config.XDLC(), "installs.json")
}

// Load returns saved install records.
func Load() []Record {
	var f file
	_ = fsutil.LoadJSON(path(), &f)
	return f.Installs
}

// Save writes install records.
func Save(items []Record) error {
	return fsutil.SaveJSON(path(), file{Version: config.Version, Installs: items}, true)
}

// Label formats a record for menus (matches Python).
func Label(rec Record) string {
	agent := config.Agents[rec.Agent]
	if agent == "" {
		agent = rec.Agent
	}
	first := strings.Split(agent, " ")[0]
	scope := config.ScopeName[rec.Scope]
	if scope == "" {
		scope = rec.Scope
	}
	where := ""
	if rec.Scope != "global" {
		where = "  " + fsutil.Short(rec.Path)
	}
	return fmt.Sprintf("%-7s %-7s%s", first, scope, where)
}

// Same reports whether two records identify the same setup slot.
func Same(a, b Record) bool {
	return a.Agent == b.Agent && a.Scope == b.Scope && a.Path == b.Path
}

// Reapply reinstalls every saved setup (non-interactive).
func Reapply(items []Record, interactive bool) error {
	bundle, err := pack.EnsureBundle()
	if err != nil {
		return err
	}
	done := make([]Record, 0, len(items))
	for _, r := range items {
		if r.Scope != "global" {
			if st, err := os.Stat(r.Path); err != nil || !st.IsDir() {
				fmt.Println(tui.Dim(fmt.Sprintf("  Skipping %s: directory no longer exists.", strings.TrimSpace(Label(r)))))
				done = append(done, r)
				continue
			}
		}
		roleID := r.Role
		if roleID == "" {
			roleID = "dev"
		}
		agentsSrc := filepath.Join(bundle, "roles", roleID, "AGENTS.md")
		if _, err := os.Stat(agentsSrc); err != nil {
			agentsSrc = filepath.Join(bundle, "roles", "dev", "AGENTS.md")
		}
		var newRec Record
		var setupErr error
		old := toSetupRecord(r)
		switch r.Agent {
		case "kiro":
			sr, err := setup.SetupKiro(bundle, r.Scope, r.Path, &old, interactive, roleID, agentsSrc, nil)
			newRec, setupErr = fromSetupRecord(sr), err
		case "claude":
			sr, err := setup.SetupClaude(bundle, r.Scope, r.Path, &old, interactive, roleID, agentsSrc, nil)
			newRec, setupErr = fromSetupRecord(sr), err
		default:
			setupErr = cli.Fail("unknown agent " + r.Agent)
		}
		if setupErr != nil {
			fmt.Printf("  %s updating %s: %v\n", tui.Warn("Error"), strings.TrimSpace(Label(r)), setupErr)
			done = append(done, r)
			continue
		}
		done = append(done, newRec)
		fmt.Printf("  %s %s\n", tui.Ok("Updated"), strings.TrimSpace(Label(r)))
	}
	return Save(done)
}

// ToSetupRecord converts a persisted record for setup package.
func ToSetupRecord(r Record) setup.Record {
	return toSetupRecord(r)
}

func toSetupRecord(r Record) setup.Record {
	return setup.Record{
		Agent: r.Agent, Scope: r.Scope, Path: r.Path, Version: r.Version, Role: r.Role,
		Files: r.Files, McpJSON: r.McpJSON, CreatedMcpJSON: r.CreatedMcpJSON,
		GitTracked: r.GitTracked, SettingsRel: r.SettingsRel,
	}
}

func fromSetupRecord(r setup.Record) Record {
	return Record{
		Agent: r.Agent, Scope: r.Scope, Path: r.Path, Version: r.Version, Role: r.Role,
		Files: r.Files, McpJSON: r.McpJSON, CreatedMcpJSON: r.CreatedMcpJSON,
		GitTracked: r.GitTracked, SettingsRel: r.SettingsRel,
	}
}