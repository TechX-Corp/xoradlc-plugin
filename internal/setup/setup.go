package setup

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/TechX-Corp/xoradlc-plugin/internal/cli"
	"github.com/TechX-Corp/xoradlc-plugin/internal/config"
	"github.com/TechX-Corp/xoradlc-plugin/internal/fsutil"
	"github.com/TechX-Corp/xoradlc-plugin/internal/gitx"
	"github.com/TechX-Corp/xoradlc-plugin/internal/pack"
	"github.com/TechX-Corp/xoradlc-plugin/internal/tui"
)

var safeSkill = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

// InstallStandards copies manifest standards entries into base.
func InstallStandards(man map[string]any, bundle, base string, files map[string]string) error {
	stdSrc := filepath.Join(bundle, "standards")
	standards, _ := man["standards"].([]any)
	for _, item := range standards {
		m, _ := item.(map[string]any)
		if m == nil {
			continue
		}
		srcName := fmt.Sprint(m["source"])
		dest := fmt.Sprint(m["dest"])
		if srcName == "" || dest == "" || srcName == "<nil>" || dest == "<nil>" {
			continue
		}
		src := filepath.Join(stdSrc, srcName)
		if st, err := os.Stat(src); err == nil && !st.IsDir() {
			digest, err := fsutil.CopyFile(src, filepath.Join(base, dest))
			if err != nil {
				return err
			}
			files[dest] = digest
		}
	}
	return nil
}

// MergeClaudeHooks merges SessionStart hook from a host template into settings JSON.
// tmplPath is the absolute path to the hook JSON (from host-agents/claude/manifest.yaml).
func MergeClaudeHooks(tmplPath, settingsPath, agentsMD, guidanceMD string) error {
	var hookDoc map[string]any
	if tmplPath != "" {
		if b, err := os.ReadFile(tmplPath); err == nil {
			_ = json.Unmarshal(b, &hookDoc)
		}
	}
	if hookDoc == nil {
		hookDoc = map[string]any{"hooks": map[string]any{}}
	}
	cmd := ShellCmd(append(XdlcArgv(), "_print", agentsMD, guidanceMD))
	hooksRoot, _ := hookDoc["hooks"].(map[string]any)
	for _, groups := range hooksRoot {
		glist, _ := groups.([]any)
		for _, g := range glist {
			gm, _ := g.(map[string]any)
			inner, _ := gm["hooks"].([]any)
			for _, h := range inner {
				hm, _ := h.(map[string]any)
				if _, ok := hm["command"]; ok {
					hm["command"] = cmd
				}
			}
		}
	}
	var settings map[string]any
	_ = fsutil.LoadJSON(settingsPath, &settings)
	if settings == nil {
		settings = map[string]any{}
	}
	hooks := settings
	if h, ok := settings["hooks"].(map[string]any); ok {
		hooks = h
	} else {
		hooks = map[string]any{}
		settings["hooks"] = hooks
	}
	srcHooks, _ := hookDoc["hooks"].(map[string]any)
	if sess, ok := srcHooks["SessionStart"]; ok {
		hooks["SessionStart"] = sess
	} else {
		hooks["SessionStart"] = []any{}
	}
	return fsutil.SaveJSON(settingsPath, settings, false)
}

// claudeHookTemplate resolves the first hooks[].source under host-agents/claude/.
func claudeHookTemplate(man map[string]any, bundle string) string {
	hooks, _ := man["hooks"].([]any)
	for _, item := range hooks {
		m, _ := item.(map[string]any)
		if m == nil {
			continue
		}
		src := strings.TrimSpace(fmt.Sprint(m["source"]))
		if src == "" || src == "<nil>" {
			continue
		}
		rel := filepath.FromSlash(src)
		if filepath.IsAbs(rel) {
			return rel
		}
		// Prefer paths relative to this host pack; allow legacy repo-root paths too.
		hostRel := filepath.Join(bundle, "host-agents", "claude", rel)
		if st, err := os.Stat(hostRel); err == nil && !st.IsDir() {
			return hostRel
		}
		rootRel := filepath.Join(bundle, rel)
		if st, err := os.Stat(rootRel); err == nil && !st.IsDir() {
			return rootRel
		}
		return hostRel
	}
	return ""
}

// Record is persisted via installs package (same JSON shape).
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

// SetupKiro installs the Kiro pack.
func SetupKiro(bundle, scope, path string, old *Record, interactive bool, roleID, agentsSrc string, skills []pack.Skill) (Record, error) {
	man, err := pack.HostManifest("kiro", bundle)
	if err != nil {
		return Record{}, err
	}
	base := path
	mcpRel := nestedString(man, "mcp", "config_path")
	agentsRel := nestedString(man, "artifacts", "agents_md")
	skillsRel := nestedString(man, "artifacts", "skills_dir")
	plan := []string{agentsRel, mcpRel}
	for _, it := range standardsDests(man) {
		plan = append(plan, it)
	}
	var oldMap map[string]any
	if old != nil {
		oldMap = recordToMap(old)
	}
	gitChoice, err := gitx.LocalGit(base, plan, oldMap, scope == "local" && interactive)
	if err != nil {
		return Record{}, err
	}
	if scope != "local" {
		gitChoice = ""
	}
	files := map[string]string{}
	digest, err := fsutil.CopyFile(agentsSrc, filepath.Join(base, agentsRel))
	if err != nil {
		return Record{}, err
	}
	files[agentsRel] = digest
	if err := InstallStandards(man, bundle, base, files); err != nil {
		return Record{}, err
	}
	for _, sk := range skills {
		safe := safeSkill.ReplaceAllString(sk.Name, "-")
		safe = strings.Trim(safe, "-")
		if safe == "" {
			safe = "skill"
		}
		rel := filepath.ToSlash(filepath.Join(skillsRel, safe, "SKILL.md"))
		dst := filepath.Join(base, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return Record{}, err
		}
		if err := os.WriteFile(dst, []byte(pack.SkillMD(sk.Name, sk.Description)), 0o644); err != nil {
			return Record{}, err
		}
		d, _ := fsutil.SHA256File(dst)
		files[rel] = d
	}
	if old != nil {
		for rel, digest := range old.Files {
			if _, ok := files[rel]; !ok {
				removeFile(base, rel, digest, interactive)
			}
		}
	}
	mcpFile := filepath.Join(base, mcpRel)
	created := false
	if old != nil {
		created = old.CreatedMcpJSON
	} else {
		created = !pathExists(mcpFile)
	}
	if err := WriteMcpConfig(mcpFile, McpEntry()); err != nil {
		return Record{}, err
	}
	rec := Record{
		Agent: "kiro", Scope: scope, Path: base, Version: config.Version, Role: roleID,
		Files: files, McpJSON: mcpRel, CreatedMcpJSON: created, GitTracked: gitChoice,
	}
	if scope == "local" {
		excluded := []string{}
		for rel := range files {
			if gitChoice != "keep" || len(gitx.GitTracked(base, []string{rel})) == 0 {
				excluded = append(excluded, rel)
			}
		}
		if created || gitChoice == "untrack" {
			excluded = append(excluded, mcpRel)
		}
		_ = gitx.SetExclude(base, excluded)
	}
	return rec, nil
}

// SetupClaude installs the Claude Code pack.
func SetupClaude(bundle, scope, path string, old *Record, interactive bool, roleID, agentsSrc string, skills []pack.Skill) (Record, error) {
	if _, err := exec.LookPath("claude"); err != nil {
		return Record{}, cli.Fail("Command 'claude' not found. Install Claude Code first: https://code.claude.com")
	}
	man, err := pack.HostManifest("claude", bundle)
	if err != nil {
		return Record{}, err
	}
	base := config.Home()
	if scope == "local" {
		base = path
	}
	mcpRel := nestedString(man, "mcp", "config_path_global")
	if scope == "local" {
		mcpRel = nestedString(man, "mcp", "config_path_local")
	}
	mcpPath := filepath.Join(base, filepath.FromSlash(strings.TrimPrefix(mcpRel, "./")))
	if scope == "global" && mcpRel == ".claude.json" {
		mcpPath = filepath.Join(config.Home(), ".claude.json")
	}
	agentsRel := nestedString(man, "artifacts", "agents_md")
	skillsRel := nestedString(man, "artifacts", "skills_dir")
	settingsRel := nestedString(man, "artifacts", "settings_global")
	if scope == "local" {
		settingsRel = nestedString(man, "artifacts", "settings_local")
	}
	plan := []string{settingsRel, mcpRel, agentsRel}
	var oldMap map[string]any
	if old != nil {
		oldMap = recordToMap(old)
	}
	gitChoice, err := gitx.LocalGit(base, plan, oldMap, scope == "local" && interactive)
	if err != nil {
		return Record{}, err
	}
	if scope != "local" {
		gitChoice = ""
	}
	files := map[string]string{}
	agentsDst := filepath.Join(base, agentsRel)
	digest, err := fsutil.CopyFile(agentsSrc, agentsDst)
	if err != nil {
		return Record{}, err
	}
	files[agentsRel] = digest
	if err := InstallStandards(man, bundle, base, files); err != nil {
		return Record{}, err
	}
	for _, sk := range skills {
		safe := safeSkill.ReplaceAllString(sk.Name, "-")
		safe = strings.Trim(safe, "-")
		if safe == "" {
			safe = "skill"
		}
		rel := filepath.ToSlash(filepath.Join(skillsRel, safe, "SKILL.md"))
		dst := filepath.Join(base, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return Record{}, err
		}
		if err := os.WriteFile(dst, []byte(pack.SkillMD(sk.Name, sk.Description)), 0o644); err != nil {
			return Record{}, err
		}
		d, _ := fsutil.SHA256File(dst)
		files[rel] = d
	}
	guidanceRel := ""
	for _, d := range standardsDests(man) {
		if strings.Contains(d, "GUIDANCE") {
			guidanceRel = d
			break
		}
	}
	guidancePath := agentsDst
	if guidanceRel != "" {
		gp := filepath.Join(base, guidanceRel)
		if st, err := os.Stat(gp); err == nil && !st.IsDir() {
			guidancePath = gp
		}
	}
	settingsPath := filepath.Join(base, settingsRel)
	tmpl := claudeHookTemplate(man, bundle)
	if err := MergeClaudeHooks(tmpl, settingsPath, agentsDst, guidancePath); err != nil {
		return Record{}, err
	}
	d, _ := fsutil.SHA256File(settingsPath)
	files[settingsRel] = d
	created := false
	if old != nil {
		created = old.CreatedMcpJSON
	} else {
		created = !pathExists(mcpPath)
	}
	if err := WriteMcpConfig(mcpPath, McpEntry()); err != nil {
		return Record{}, err
	}
	mcpTrack := relOrAbs(base, mcpPath)
	dm, _ := fsutil.SHA256File(mcpPath)
	files[mcpTrack] = dm
	if old != nil {
		for rel, digest := range old.Files {
			if _, ok := files[rel]; !ok {
				removeFile(base, rel, digest, interactive)
			}
		}
	}
	if scope == "local" && gitChoice != "keep" {
		_ = gitx.SetExclude(base, []string{settingsRel, mcpRel})
	}
	return Record{
		Agent: "claude", Scope: scope, Path: base, Version: config.Version, Role: roleID,
		Files: files, McpJSON: mcpTrack, CreatedMcpJSON: created, GitTracked: gitChoice, SettingsRel: settingsRel,
	}, nil
}

// RemoveKiro removes a Kiro setup.
func RemoveKiro(rec Record, interactive bool) error {
	base := rec.Path
	for rel, digest := range rec.Files {
		removeFile(base, rel, digest, interactive)
	}
	mcpRel := rec.McpJSON
	if mcpRel == "" {
		mcpRel = ".kiro/settings/mcp.json"
	}
	_ = StripMcpConfig(filepath.Join(base, mcpRel), rec.CreatedMcpJSON)
	if rec.Scope == "local" {
		_ = gitx.SetExclude(base, nil)
	}
	kiro := filepath.Join(base, ".kiro")
	if st, err := os.Stat(kiro); err == nil && st.IsDir() && base != config.Home() {
		filepath.WalkDir(kiro, func(path string, d os.DirEntry, err error) error {
			if err != nil || !d.IsDir() {
				return nil
			}
			entries, _ := os.ReadDir(path)
			if len(entries) == 0 {
				_ = os.Remove(path)
			}
			return nil
		})
		entries, _ := os.ReadDir(kiro)
		if len(entries) == 0 {
			_ = os.Remove(kiro)
		}
	}
	return nil
}

// RemoveClaude removes a Claude setup.
func RemoveClaude(rec Record, interactive bool) error {
	base := rec.Path
	for rel, digest := range rec.Files {
		p := filepath.Join(base, rel)
		if filepath.IsAbs(rel) {
			p = rel
		}
		removeFilePath(p, digest, interactive)
	}
	mcp := rec.McpJSON
	if !filepath.IsAbs(mcp) {
		mcp = filepath.Join(base, mcp)
	}
	_ = StripMcpConfig(mcp, rec.CreatedMcpJSON)
	if rec.Scope == "local" {
		_ = gitx.SetExclude(base, nil)
	}
	return nil
}

func removeFile(base, rel, digest string, interactive bool) {
	removeFilePath(filepath.Join(base, rel), digest, interactive)
}

func removeFilePath(p, digest string, interactive bool) {
	st, err := os.Stat(p)
	if err != nil || st.IsDir() {
		return
	}
	cur, err := fsutil.SHA256File(p)
	if err == nil && cur != digest {
		if interactive {
			ok, err := tui.Confirm(fmt.Sprintf("  %s was modified after install. Delete anyway?", fsutil.Short(p)))
			if err != nil || !ok {
				fmt.Printf("  Keeping %s (modified).\n", fsutil.Short(p))
				return
			}
		} else {
			fmt.Printf("  Keeping %s (modified).\n", fsutil.Short(p))
			return
		}
	}
	_ = os.Remove(p)
}

func nestedString(m map[string]any, keys ...string) string {
	var cur any = m
	for _, k := range keys {
		cm, _ := cur.(map[string]any)
		if cm == nil {
			return ""
		}
		cur = cm[k]
	}
	return fmt.Sprint(cur)
}

func standardsDests(man map[string]any) []string {
	var out []string
	standards, _ := man["standards"].([]any)
	for _, item := range standards {
		im, _ := item.(map[string]any)
		if im == nil {
			continue
		}
		if d := fmt.Sprint(im["dest"]); d != "" && d != "<nil>" {
			out = append(out, d)
		}
	}
	return out
}

func recordToMap(r *Record) map[string]any {
	if r == nil {
		return nil
	}
	return map[string]any{"git_tracked": r.GitTracked}
}