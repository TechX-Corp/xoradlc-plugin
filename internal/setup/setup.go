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

// InstallKiroHooks copies host-agents/kiro hook JSON files into the workspace (.kiro/hooks).
func InstallKiroHooks(man map[string]any, bundle, base string, files map[string]string) error {
	for _, h := range manifestHookEntries(man) {
		src := resolveHostHookFile(bundle, "kiro", h.source)
		dest := h.dest
		if dest == "" || dest == "<nil>" {
			dest = filepath.ToSlash(filepath.Join(".kiro", "hooks", filepath.Base(h.source)))
		}
		dest = filepath.ToSlash(dest)
		b, err := os.ReadFile(src)
		if err != nil {
			return err
		}
		var doc map[string]any
		if err := json.Unmarshal(b, &doc); err != nil {
			return err
		}
		// Rewrite command paths to the installed xdlc binary.
		usageCmd := ShellCmd(append(XdlcArgv(), "usage", "--hook", "kiro"))
		flushCmd := ShellCmd(append(XdlcArgv(), "usage", "--hook", "kiro", "--flush"))
		if hooks, ok := doc["hooks"].([]any); ok {
			for _, item := range hooks {
				hm, _ := item.(map[string]any)
				if hm == nil {
					continue
				}
				action, _ := hm["action"].(map[string]any)
				if action == nil {
					continue
				}
				c, _ := action["command"].(string)
				if strings.Contains(c, "--flush") {
					action["command"] = flushCmd
				} else if strings.Contains(c, "usage") {
					action["command"] = usageCmd
				}
			}
		}
		out, err := json.MarshalIndent(doc, "", "  ")
		if err != nil {
			return err
		}
		dst := filepath.Join(base, filepath.FromSlash(dest))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(dst, append(out, '\n'), 0o644); err != nil {
			return err
		}
		d, _ := fsutil.SHA256File(dst)
		files[dest] = d
	}
	return nil
}

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
		dest = filepath.ToSlash(dest)
		if st, err := os.Stat(src); err == nil && !st.IsDir() {
			digest, err := fsutil.CopyFile(src, filepath.Join(base, filepath.FromSlash(dest)))
			if err != nil {
				return err
			}
			files[dest] = digest
		}
	}
	return nil
}

// MergeClaudeHooks merges host-agents/claude hook templates into Claude settings JSON.
// kind session_start_print → rewrite command to xdlc _print; kind usage → xdlc usage --hook claude.
func MergeClaudeHooks(man map[string]any, bundle, settingsPath, agentsMD, guidanceMD string) error {
	var settings map[string]any
	_ = fsutil.LoadJSON(settingsPath, &settings)
	if settings == nil {
		settings = map[string]any{}
	}
	hooksOut, _ := settings["hooks"].(map[string]any)
	if hooksOut == nil {
		hooksOut = map[string]any{}
		settings["hooks"] = hooksOut
	}
	printCmd := ShellCmd(append(XdlcArgv(), "_print", agentsMD, guidanceMD))
	usageCmd := ShellCmd(append(XdlcArgv(), "usage", "--hook", "claude"))
	usageFlushCmd := ShellCmd(append(XdlcArgv(), "usage", "--hook", "claude", "--flush"))

	for _, item := range manifestHookEntries(man) {
		path := resolveHostHookFile(bundle, "claude", item.source)
		doc := loadHookDoc(path)
		srcHooks, _ := doc["hooks"].(map[string]any)
		if srcHooks == nil {
			continue
		}
		switch item.kind {
		case "session_start_print", "":
			rewriteHookCommands(srcHooks, printCmd)
			if sess, ok := srcHooks["SessionStart"]; ok {
				hooksOut["SessionStart"] = sess
			}
		case "usage":
			for ev, groups := range srcHooks {
				rewriteUsageCommands(groups, usageCmd, usageFlushCmd)
				if ev == "SessionStart" {
					// Append usage flush after AGENTS print.
					existing, _ := hooksOut["SessionStart"].([]any)
					add, _ := groups.([]any)
					hooksOut["SessionStart"] = append(existing, add...)
					continue
				}
				hooksOut[ev] = groups
			}
		}
	}
	return fsutil.SaveJSON(settingsPath, settings, false)
}

type hookEntry struct {
	source string
	kind   string
	dest   string
}

func manifestHookEntries(man map[string]any) []hookEntry {
	raw, _ := man["hooks"].([]any)
	var out []hookEntry
	for _, item := range raw {
		m, _ := item.(map[string]any)
		if m == nil {
			continue
		}
		src := strings.TrimSpace(fmt.Sprint(m["source"]))
		if src == "" || src == "<nil>" {
			continue
		}
		out = append(out, hookEntry{
			source: src,
			kind:   strings.TrimSpace(fmt.Sprint(m["kind"])),
			dest:   strings.TrimSpace(fmt.Sprint(m["dest"])),
		})
	}
	return out
}

func resolveHostHookFile(bundle, agent, src string) string {
	rel := filepath.FromSlash(src)
	if filepath.IsAbs(rel) {
		return rel
	}
	hostRel := filepath.Join(bundle, "host-agents", agent, rel)
	if st, err := os.Stat(hostRel); err == nil && !st.IsDir() {
		return hostRel
	}
	return filepath.Join(bundle, rel)
}

func loadHookDoc(path string) map[string]any {
	var hookDoc map[string]any
	if path != "" {
		if b, err := os.ReadFile(path); err == nil {
			_ = json.Unmarshal(b, &hookDoc)
		}
	}
	if hookDoc == nil {
		hookDoc = map[string]any{"hooks": map[string]any{}}
	}
	return hookDoc
}

func rewriteHookCommands(hooksRoot map[string]any, cmd string) {
	for _, groups := range hooksRoot {
		glist, _ := groups.([]any)
		for _, g := range glist {
			gm, _ := g.(map[string]any)
			inner, _ := gm["hooks"].([]any)
			for _, h := range inner {
				hm, _ := h.(map[string]any)
				if hm == nil {
					continue
				}
				if _, ok := hm["command"]; ok {
					hm["command"] = cmd
				}
			}
		}
	}
}

func rewriteUsageCommands(groups any, usageCmd, flushCmd string) {
	glist, _ := groups.([]any)
	for _, g := range glist {
		gm, _ := g.(map[string]any)
		inner, _ := gm["hooks"].([]any)
		for _, h := range inner {
			hm, _ := h.(map[string]any)
			if hm == nil {
				continue
			}
			c, _ := hm["command"].(string)
			if strings.Contains(c, "--flush") {
				hm["command"] = flushCmd
			} else {
				hm["command"] = usageCmd
			}
		}
	}
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
	agentsRel = filepath.ToSlash(agentsRel)
	mcpRel = filepath.ToSlash(mcpRel)
	skillsRel = filepath.ToSlash(skillsRel)
	digest, err := fsutil.CopyFile(agentsSrc, filepath.Join(base, filepath.FromSlash(agentsRel)))
	if err != nil {
		return Record{}, err
	}
	files[agentsRel] = digest
	if err := InstallStandards(man, bundle, base, files); err != nil {
		return Record{}, err
	}
	if err := InstallKiroHooks(man, bundle, base, files); err != nil {
		return Record{}, err
	}
	for _, sk := range skills {
		safe := safeSkill.ReplaceAllString(sk.Name, "-")
		safe = strings.Trim(safe, "-")
		if safe == "" {
			safe = "skill"
		}
		rel := filepath.ToSlash(filepath.Join(filepath.FromSlash(skillsRel), safe, "SKILL.md"))
		dst := filepath.Join(base, filepath.FromSlash(rel))
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
		if strings.Contains(d, "MCP_ROUTING") {
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
	if err := MergeClaudeHooks(man, bundle, settingsPath, agentsDst, guidancePath); err != nil {
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