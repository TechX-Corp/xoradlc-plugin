package gitx

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetExcludeRootPatterns(t *testing.T) {
	dir := t.TempDir()
	if out, err := exec.Command("git", "-C", dir, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v (%s)", err, out)
	}
	rels := []string{
		".kiro/steering/xora-dlc.md",
		filepath.FromSlash(".kiro/settings/mcp.json"), // may be backslash on Windows
		".kiro/skills/implement/SKILL.md",
	}
	if err := SetExclude(dir, rels); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, ".git", "info", "exclude"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)
	for _, want := range []string{
		"/.kiro/steering/xora-dlc.md",
		"/.kiro/settings/mcp.json",
		"/.kiro/skills/implement/SKILL.md",
	} {
		found := false
		for _, ln := range strings.Split(text, "\n") {
			if ln == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("exclude missing line %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, `\`) {
		t.Fatalf("exclude has backslashes:\n%s", text)
	}
}

func TestSameDirFold(t *testing.T) {
	dir := t.TempDir()
	a, _ := filepath.Abs(dir)
	b := filepath.Join(a, ".", ".")
	if !sameDir(a, b) {
		t.Fatalf("sameDir(%q,%q) = false", a, b)
	}
}
