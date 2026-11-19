package pack

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TechX-Corp/xoradlc-plugin/internal/auth"
	"github.com/TechX-Corp/xoradlc-plugin/internal/cli"
	"github.com/TechX-Corp/xoradlc-plugin/internal/tui"
	"gopkg.in/yaml.v3"
)

type catalogFile struct {
	Roles []roleEntry `yaml:"roles"`
}

type roleEntry struct {
	ID          string   `yaml:"id"`
	Enabled     bool     `yaml:"enabled"`
	WhoamiRoles []string `yaml:"whoami_roles"`
	Path        string   `yaml:"path"`
}

// HostManifest loads host-agents/<agent>/manifest.yaml.
func HostManifest(agent, bundle string) (map[string]any, error) {
	if bundle == "" {
		var err error
		bundle, err = EnsureBundle()
		if err != nil {
			return nil, err
		}
	}
	path := filepath.Join(bundle, "host-agents", agent, "manifest.yaml")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, cli.Fail(fmt.Sprintf("Missing host manifest: %s", path))
	}
	var doc map[string]any
	if err := yaml.Unmarshal(b, &doc); err != nil {
		return nil, err
	}
	return doc, nil
}

// CatalogRoles returns role entries from roles/catalog.yaml.
func CatalogRoles(bundle string) ([]roleEntry, error) {
	if bundle == "" {
		var err error
		bundle, err = EnsureBundle()
		if err != nil {
			return nil, err
		}
	}
	path := filepath.Join(bundle, "roles", "catalog.yaml")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, cli.Fail(fmt.Sprintf("Missing roles/catalog.yaml at %s", path))
	}
	var cat catalogFile
	if err := yaml.Unmarshal(b, &cat); err != nil {
		return nil, err
	}
	return cat.Roles, nil
}

// ResolveRolePack returns role id and AGENTS.md path for the logged-in account.
func ResolveRolePack(bundle string) (roleID string, agentsPath string, err error) {
	if bundle == "" {
		bundle, err = EnsureBundle()
		if err != nil {
			return "", "", err
		}
	}
	u, why := auth.Whoami()
	if u == nil {
		return "", "", cli.Usage(fmt.Sprintf("Setup requires login: %s. Run %s first.", why, tui.Em("xdlc login")))
	}
	raw := strings.TrimSpace(fmt.Sprint(u["role"]))
	if raw == "" {
		return "", "", cli.Fail(fmt.Sprintf(
			"Account %s has no role. Ask a Xora DLC admin to assign a role (e.g. DEV), then run setup again.",
			auth.AccountLine(u)))
	}
	roles, err := CatalogRoles(bundle)
	if err != nil {
		return "", "", err
	}
	key := strings.ToLower(raw)
	for _, r := range roles {
		if !r.Enabled {
			continue
		}
		aliases := append([]string{strings.ToLower(r.ID)}, lowerAll(r.WhoamiRoles)...)
		for _, a := range aliases {
			if a == key {
				rid := r.ID
				rel := r.Path
				if rel == "" {
					rel = "roles/" + rid
				}
				agents := filepath.Join(bundle, filepath.FromSlash(rel), "AGENTS.md")
				if st, e := os.Stat(agents); e != nil || st.IsDir() {
					return "", "", cli.Fail(fmt.Sprintf("Role '%s' is enabled but %s is missing (catalog path: %s).", rid, agents, rel))
				}
				return rid, agents, nil
			}
		}
	}
	var enabled []string
	for _, r := range roles {
		if r.Enabled {
			enabled = append(enabled, r.ID)
		}
	}
	return "", "", cli.Fail(fmt.Sprintf(
		"Account role %s is not enabled in this plugin.\n  Enabled role ids: %s\n  Edit roles/catalog.yaml (whoami_roles) or use an account with an enabled role.",
		tui.Em(raw), strings.Join(enabled, ", ")))
}

func lowerAll(in []string) []string {
	out := make([]string, len(in))
	for i, s := range in {
		out[i] = strings.ToLower(s)
	}
	return out
}