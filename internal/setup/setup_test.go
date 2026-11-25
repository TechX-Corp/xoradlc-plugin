package setup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TechX-Corp/xoradlc-plugin/internal/fsutil"
)

func claudeHookManifest() map[string]any {
	return map[string]any{"hooks": []any{
		map[string]any{"source": "hooks/session-start.json", "kind": "session_start_print"},
		map[string]any{"source": "hooks/usage.json", "kind": "usage"},
	}}
}

func hookCommands(t *testing.T, settingsPath, ev string) []string {
	t.Helper()
	b, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Hooks map[string][]struct {
			Hooks []struct {
				Command string `json:"command"`
			} `json:"hooks"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, g := range doc.Hooks[ev] {
		for _, h := range g.Hooks {
			out = append(out, h.Command)
		}
	}
	return out
}

func TestRemoveClaudeKeepsTheUsersSettingsAndServers(t *testing.T) {
	bundle, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	base := t.TempDir()
	settingsRel := ".claude/settings.local.json"
	settingsPath := filepath.Join(base, filepath.FromSlash(settingsRel))
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatal(err)
	}
	user := `{"permissions":{"allow":["Bash(ls)"]},"hooks":{"Stop":[{"hooks":[{"type":"command","command":"notify-send done"}]}]}}`
	if err := os.WriteFile(settingsPath, []byte(user), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := MergeClaudeHooks(claudeHookManifest(), bundle, settingsPath, "/a/CLAUDE.md", "/a/MCP_ROUTING.md"); err != nil {
		t.Fatal(err)
	}
	mcpPath := filepath.Join(base, ".mcp.json")
	if err := os.WriteFile(mcpPath, []byte(`{"mcpServers":{"postgres":{"command":"pg"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteMcpConfig(mcpPath, McpEntry()); err != nil {
		t.Fatal(err)
	}
	sd, _ := fsutil.SHA256File(settingsPath)
	md, _ := fsutil.SHA256File(mcpPath)
	// Digests match what setup recorded: the old remove deleted both files outright.
	rec := Record{Agent: "claude", Scope: "local", Path: base, SettingsRel: settingsRel, McpJSON: ".mcp.json",
		Files: map[string]string{settingsRel: sd, ".mcp.json": md}}
	if err := RemoveClaude(rec, false); err != nil {
		t.Fatal(err)
	}

	if got := hookCommands(t, settingsPath, "Stop"); len(got) != 1 || got[0] != "notify-send done" {
		t.Fatalf("Stop = %q", got)
	}
	for _, ev := range []string{"SessionStart", "SubagentStop", "SessionEnd"} {
		if got := hookCommands(t, settingsPath, ev); len(got) != 0 {
			t.Fatalf("%s left: %q", ev, got)
		}
	}
	b, _ := os.ReadFile(settingsPath)
	if !strings.Contains(string(b), "Bash(ls)") {
		t.Fatalf("permissions lost: %s", b)
	}
	m, err := os.ReadFile(mcpPath)
	if err != nil || !strings.Contains(string(m), "postgres") || strings.Contains(string(m), "xora-dlc") {
		t.Fatalf(".mcp.json = %s, %v", m, err)
	}
}

func TestStripClaudeHooksDeletesASettingsFileOnlyXdlcFilled(t *testing.T) {
	bundle, _ := filepath.Abs(filepath.Join("..", ".."))
	settingsPath := filepath.Join(t.TempDir(), "settings.local.json")
	if err := MergeClaudeHooks(claudeHookManifest(), bundle, settingsPath, "/a/CLAUDE.md", "/a/MCP_ROUTING.md"); err != nil {
		t.Fatal(err)
	}
	if err := StripClaudeHooks(settingsPath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(settingsPath); !os.IsNotExist(err) {
		t.Fatalf("settings file should be gone, stat err = %v", err)
	}
}

func TestMergeClaudeHooksKeepsUserHooksAndIsIdempotent(t *testing.T) {
	bundle, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	settingsPath := filepath.Join(t.TempDir(), "settings.json")
	user := `{"model":"opus","hooks":{
		"Stop":[{"hooks":[{"type":"command","command":"notify-send done"}]}],
		"SessionStart":[{"matcher":"startup","hooks":[{"type":"command","command":"echo hi"}]}]
	}}`
	if err := os.WriteFile(settingsPath, []byte(user), 0o644); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 2; i++ {
		if err := MergeClaudeHooks(claudeHookManifest(), bundle, settingsPath, "/a/CLAUDE.md", "/a/MCP_ROUTING.md"); err != nil {
			t.Fatal(err)
		}
	}

	stop := hookCommands(t, settingsPath, "Stop")
	if len(stop) != 2 || stop[0] != "notify-send done" || !strings.HasSuffix(stop[1], " usage --hook claude") {
		t.Fatalf("Stop = %q", stop)
	}
	start := hookCommands(t, settingsPath, "SessionStart")
	if len(start) != 3 || start[0] != "echo hi" || !strings.Contains(start[1], " _print ") ||
		!strings.HasSuffix(start[2], " usage --hook claude --flush") {
		t.Fatalf("SessionStart = %q", start)
	}
	for _, ev := range []string{"SubagentStop", "SessionEnd"} {
		if got := hookCommands(t, settingsPath, ev); len(got) != 1 {
			t.Fatalf("%s = %q", ev, got)
		}
	}
	b, _ := os.ReadFile(settingsPath)
	if !strings.Contains(string(b), `"model": "opus"`) {
		t.Fatalf("other settings lost: %s", b)
	}
}
