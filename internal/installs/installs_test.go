package installs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TechX-Corp/xoradlc-plugin/internal/fsutil"
)

func write(t *testing.T, path, body string) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	d, err := fsutil.SHA256File(path)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	return string(b)
}

func TestReapplyRemovesLegacyGlobalSetupsButKeepsTheUsersConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDLC_PLUGIN_DIR", filepath.Join("..", ".."))

	steering := filepath.Join(home, ".kiro", "steering", "xora-dlc.md")
	kiroFiles := map[string]string{".kiro/steering/xora-dlc.md": write(t, steering, "role")}
	kiroMcp := filepath.Join(home, ".kiro", "settings", "mcp.json")
	write(t, kiroMcp, `{"mcpServers":{"postgres":{"command":"pg"},"xora-dlc":{"command":"xdlc"}}}`)

	settings := filepath.Join(home, ".claude", "settings.json")
	sd := write(t, settings, `{"model":"opus","hooks":{"Stop":[`+
		`{"hooks":[{"type":"command","command":"notify-send done"}]},`+
		`{"hooks":[{"type":"command","command":"\"/u/.local/bin/xdlc\" usage --hook claude"}]}]}}`)
	claudeJSON := filepath.Join(home, ".claude.json")
	cd := write(t, claudeJSON, `{"numStartups":42,"mcpServers":{"xora-dlc":{"command":"xdlc"}}}`)
	claudeMd := filepath.Join(home, "CLAUDE.md")
	md := write(t, claudeMd, "role")

	items := []Record{
		{ID: "k1", Agent: "kiro", Scope: "global", Path: home, Files: kiroFiles, McpJSON: ".kiro/settings/mcp.json"},
		{ID: "c1", Agent: "claude", Scope: "global", Path: home, McpJSON: ".claude.json", SettingsRel: ".claude/settings.json",
			Files: map[string]string{".claude/settings.json": sd, ".claude.json": cd, "CLAUDE.md": md}},
	}
	if err := Reapply(items, false); err != nil {
		t.Fatal(err)
	}

	if left := Load(); len(left) != 0 {
		t.Fatalf("installs left: %+v", left)
	}
	if _, err := os.Stat(steering); !os.IsNotExist(err) {
		t.Fatalf("kiro steering should be removed: %v", err)
	}
	if _, err := os.Stat(claudeMd); !os.IsNotExist(err) {
		t.Fatalf("CLAUDE.md should be removed: %v", err)
	}
	if s := read(t, kiroMcp); !strings.Contains(s, "postgres") || strings.Contains(s, "xora-dlc") {
		t.Fatalf("kiro mcp.json = %s", s)
	}
	if s := read(t, settings); !strings.Contains(s, "opus") || !strings.Contains(s, "notify-send done") || strings.Contains(s, "usage --hook") {
		t.Fatalf("settings.json = %s", s)
	}
	if s := read(t, claudeJSON); !strings.Contains(s, "numStartups") || strings.Contains(s, "xora-dlc") {
		t.Fatalf(".claude.json = %s", s)
	}
}
