package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/TechX-Corp/xoradlc-plugin/internal/auth"
	"github.com/TechX-Corp/xoradlc-plugin/internal/cli"
	"github.com/TechX-Corp/xoradlc-plugin/internal/config"
	"github.com/TechX-Corp/xoradlc-plugin/internal/fsutil"
	"github.com/TechX-Corp/xoradlc-plugin/internal/gitx"
	"github.com/TechX-Corp/xoradlc-plugin/internal/installs"
	"github.com/TechX-Corp/xoradlc-plugin/internal/link"
	"github.com/TechX-Corp/xoradlc-plugin/internal/mcp"
	"github.com/TechX-Corp/xoradlc-plugin/internal/pack"
	hostclaude "github.com/TechX-Corp/xoradlc-plugin/host-agents/claude"
	hostkiro "github.com/TechX-Corp/xoradlc-plugin/host-agents/kiro"
	"github.com/TechX-Corp/xoradlc-plugin/internal/sensor"
	"github.com/TechX-Corp/xoradlc-plugin/internal/setup"
	"github.com/TechX-Corp/xoradlc-plugin/internal/tui"
	"github.com/TechX-Corp/xoradlc-plugin/internal/update"
	"github.com/TechX-Corp/xoradlc-plugin/internal/usage"
)

// Run dispatches argv[0] command with remaining args.
func Run(argv []string) error {
	if len(argv) == 0 {
		tui.PrintHelp(MainHelp(), os.Stdout)
		return nil
	}
	if argv[0] == "-v" || argv[0] == "--version" {
		fmt.Printf("%s %s\n", tui.Bold("xdlc"), config.Version)
		return nil
	}
	cmd, rest := argv[0], argv[1:]
	if cmd == "-h" || cmd == "--help" || cmd == "help" {
		if len(rest) == 1 && (rest[0] == "-v" || rest[0] == "version") {
			if h, ok := CmdHelp("--version"); ok {
				tui.PrintHelp(h, os.Stdout)
			}
			return nil
		}
		if len(rest) == 1 {
			if h, ok := CmdHelp(rest[0]); ok {
				tui.PrintHelp(h, os.Stdout)
				return nil
			}
		}
		tui.PrintHelp(MainHelp(), os.Stdout)
		return nil
	}
	if len(rest) > 0 && (rest[0] == "-h" || rest[0] == "--help") {
		if h, ok := CmdHelp(cmd); ok {
			tui.PrintHelp(h, os.Stdout)
			return nil
		}
	}
	switch cmd {
	case "setup":
		return cmdSetup(rest)
	case "remove":
		return cmdRemove(rest)
	case "link":
		return link.CmdLink(rest)
	case "login":
		return auth.Login()
	case "logout":
		return auth.LogoutCommand(rest)
	case "status":
		return cmdStatus(rest)
	case "view":
		return cmdView(rest)
	case "update":
		return update.CmdUpdate(rest)
	case "uninstall":
		return update.CmdUninstall(rest)
	case "mcp":
		return mcp.Run()
	case "sensor":
		return sensor.CmdSensor(rest)
	case "usage":
		return usage.CmdUsage(rest, map[string]usage.Host{
			"claude": hostclaude.Host{},
			"kiro":   hostkiro.Host{},
		})
	case "_print":
		return update.CmdPrint(rest)
	case "_reapply":
		return update.CmdReapply(rest)
	case "version":
		fmt.Printf("%s xdlc %s\n", tui.Bold("xdlc"), config.Version)
		return nil
	default:
		fmt.Fprintf(os.Stderr, "xdlc: unknown command '%s'.\n", cmd)
		tui.PrintHelp(MainHelp(), os.Stderr)
		return cli.Usage("")
	}
}

func cmdSetup(argv []string) error {
	opt, err := parseOptions(argv, "agent")
	if err != nil {
		return err
	}
	if _, err := pack.EnsureBundle(); err != nil {
		return err
	}
	if err := auth.RequireLoggedIn(); err != nil {
		return err
	}
	agent := opt["agent"]
	if agent == "" {
		key, err := tui.Choose("Install for which agent?", []tui.Option{
			{Key: "kiro", Label: config.Agents["kiro"]},
			{Key: "claude", Label: config.Agents["claude"]},
		})
		if err != nil {
			return err
		}
		agent = key
	}
	if agent == "claude" {
		if _, err := exec.LookPath("claude"); err != nil {
			return cli.Fail("Command 'claude' not found. Install Claude Code first: https://code.claude.com")
		}
	}
	if agent == "kiro" {
		if _, err := exec.LookPath("kiro-cli"); err != nil {
			fmt.Println(tui.Warn("Note:") + tui.Dim(fmt.Sprintf(" command 'kiro-cli' not found on this machine; installing anyway — %s will use it once %s is installed.",
				tui.Em("Kiro"), tui.Em("Kiro CLI"))))
		}
	}
	path, _ := os.Getwd()
	if samePath(path, config.Home()) {
		// The agents read ~/.kiro and ~/.claude as machine-wide config: installing here would be global.
		return cli.Usage(fmt.Sprintf("Current directory is your home (%s), not a project. cd into a project directory and run xdlc setup again.", fsutil.Short(path)))
	}
	bundle, err := pack.EnsureBundle()
	if err != nil {
		return err
	}
	roleID, agentsSrc, err := pack.ResolveRolePack(bundle)
	if err != nil {
		return err
	}
	slot := installs.Record{Agent: agent, Scope: config.ScopeLocal, Path: path}
	items := installs.Load()
	var old *installs.Record
	for i := range items {
		if installs.Same(items[i], slot) {
			old = &items[i]
			break
		}
	}
	fmt.Println(tui.Dim("Loading skill allowlist from AAP (may take up to a minute)…"))
	skills := pack.FetchSkillAllowlist(roleID)
	if len(skills) > 0 {
		fmt.Println(tui.Dim(fmt.Sprintf("Installing %d skill(s) for role %s.", len(skills), tui.RoleLabel(roleID))))
	}
	var rec installs.Record
	var oldSetup *setup.Record
	if old != nil {
		s := installs.ToSetupRecord(*old)
		oldSetup = &s
	}
	switch agent {
	case "kiro":
		sr, err := setup.SetupKiro(bundle, path, oldSetup, true, roleID, agentsSrc, skills)
		if err != nil {
			return err
		}
		rec = fromSetup(sr)
	case "claude":
		sr, err := setup.SetupClaude(bundle, path, oldSetup, true, roleID, agentsSrc, skills)
		if err != nil {
			return err
		}
		rec = fromSetup(sr)
	default:
		return cli.Fail("unknown agent")
	}
	newItems := make([]installs.Record, 0, len(items)+1)
	for _, r := range items {
		if !installs.Same(r, rec) {
			newItems = append(newItems, r)
		}
	}
	rec.UpdatedAt = time.Now().Unix()
	if old != nil {
		rec.ID = old.ID
	} else {
		rec.ID = installs.NewID(items)
	}
	newItems = append(newItems, rec)
	if err := installs.Save(newItems); err != nil {
		return err
	}
	verb := tui.Ok("Installed")
	if old != nil {
		verb = tui.Ok("Reinstalled")
	}
	fmt.Println()
	fmt.Printf("%s %s for %s: %s · role %s\n", verb, tui.Em("xora-dlc"), tui.Em(config.Agents[agent]),
		strings.TrimSpace(installs.Label(rec)), tui.RoleLabel(roleID))
	fmt.Println(tui.Warn("Warning:") + tui.Dim(fmt.Sprintf(" setup adds agent instructions that prefer MCP %s / %s for Xora DLC "+
		"(shared context, work orders, KB). Other plugins may still win if their routing is stronger "+
		"or the user overrides. Open a new agent session after setup.", tui.Em("xora-dlc"), tui.Em("xora_ask"))))
	tip := "open a new Claude Code session (accept the xora-dlc MCP server if asked)"
	if agent == "kiro" {
		tip = "open a new Kiro session"
	}
	fmt.Println()
	fmt.Println(tui.Dim("Next: " + tip + "."))
	return nil
}

func cmdRemove(argv []string) error {
	opt, err := parseOptions(argv, "agent")
	if err != nil {
		return err
	}
	agentFilter := opt["agent"]
	items := installs.Load()
	var cands []installs.Record
	for _, r := range items {
		if agentFilter == "" || r.Agent == agentFilter {
			cands = append(cands, r)
		}
	}
	if len(cands) == 0 {
		msg := "xora-dlc"
		if agentFilter != "" {
			msg = config.Agents[agentFilter]
		}
		return cli.Cancel(fmt.Sprintf("No %s setup found.", msg))
	}
	cwd, _ := os.Getwd()
	var chosen []installs.Record
	if len(cands) == 1 {
		ok, err := tui.Confirm("Remove " + strings.TrimSpace(installs.Label(cands[0])) + "?")
		if err != nil {
			return err
		}
		if !ok {
			return cli.Cancel("Cancelled.")
		}
		chosen = cands
	} else {
		labels := make([]string, len(cands))
		for i, r := range cands {
			labels[i] = installs.Label(r)
			if r.Scope != config.ScopeGlobal && r.Path == cwd {
				labels[i] += "   ← current directory"
			}
		}
		var idx []int
		if tui.ArrowKeys() {
			idx, err = tui.Menu("Select setups to remove:", labels, true)
			if err != nil {
				return err
			}
		} else {
			fmt.Println("Select setups to remove (e.g. 1,3 | 1-2 | all):")
			for i, text := range labels {
				fmt.Printf("  %d) %s\n", i+1, text)
			}
			for {
				ans, err := tui.Ask("> ")
				if err != nil {
					return err
				}
				idx = tui.PickMany(len(cands), ans)
				if idx != nil {
					break
				}
				fmt.Println("Enter numbers, a range, or all.")
			}
		}
		for _, i := range idx {
			chosen = append(chosen, cands[i])
		}
		fmt.Println()
		fmt.Println(tui.Bold("Will remove:"))
		for _, r := range chosen {
			var what string
			if r.Agent == "kiro" {
				what = fmt.Sprintf("%d file(s) under .kiro/, %s block in %s", len(r.Files), tui.Em(config.Server), r.McpJSON)
			} else {
				what = fmt.Sprintf("MCP %s + .claude/ files", tui.Em(config.Server))
			}
			fmt.Printf("  - %s: %s\n", strings.TrimSpace(installs.Label(r)), what)
		}
		ok, err := tui.Confirm("Continue?")
		if err != nil {
			return err
		}
		if !ok {
			return cli.Cancel("Cancelled.")
		}
	}
	fmt.Println()
	remaining := items
	for _, r := range chosen {
		sr := installs.ToSetupRecord(r)
		switch r.Agent {
		case "kiro":
			_ = setup.RemoveKiro(sr, true)
		case "claude":
			_ = setup.RemoveClaude(sr, true)
		}
		var next []installs.Record
		for _, x := range remaining {
			if !installs.Same(x, r) {
				next = append(next, x)
			}
		}
		remaining = next
		_ = installs.Save(remaining)
		fmt.Printf("%s %s\n", tui.Ok("Removed"), strings.TrimSpace(installs.Label(r)))
	}
	if len(remaining) == 0 {
		_ = os.RemoveAll(config.Bundle())
	}
	return nil
}

func cmdStatus(argv []string) error {
	if len(argv) > 0 {
		return cli.Usage("status takes no options.")
	}
	fmt.Printf("%s xdlc %s\n", tui.Bold("Version:"), config.Version)
	fmt.Println()
	u, why := auth.Whoami()
	tokens, _ := auth.Tokens()
	exp := ""
	if tokens != nil {
		exp = time.Unix(tokens.ExpiresAt, 0).Format("15:04 02/01")
	}
	if u != nil {
		fmt.Printf("%s %s · token expires %s (auto-refresh)\n", tui.Bold("Account:"), auth.AccountLine(u), exp)
	} else {
		fmt.Printf("%s %s\n", tui.Bold("Account:"), why)
	}
	cwd, _ := os.Getwd()
	fmt.Println()
	items := installs.Load()
	if len(items) == 0 {
		fmt.Printf("%s%s\n", tui.Bold("Setups:"), tui.Dim(" none (run xdlc setup)"))
	} else {
		fmt.Println(tui.Bold("Setups:"))
		var rows [][]string
		for _, r := range items {
			agent := strings.Split(config.Agents[r.Agent], " ")[0]
			rows = append(rows, []string{tui.Em(r.ID), agent, setupProjectCell(r, cwd)})
		}
		tui.PrintTable([]string{"ID", "Agent", "Project"}, rows, "  ")
		fmt.Println(tui.Dim("Details: ") + tui.Em("xdlc view <id>"))
	}
	return nil
}

func fromSetup(r setup.Record) installs.Record {
	return installs.Record{
		Agent: r.Agent, Scope: r.Scope, Path: r.Path, Version: r.Version, Role: r.Role,
		Files: r.Files, McpJSON: r.McpJSON, CreatedMcpJSON: r.CreatedMcpJSON,
		GitTracked: r.GitTracked, SettingsRel: r.SettingsRel,
	}
}

var shortFlags = map[string]string{"-a": "agent"}

func parseOptions(argv []string, allowed ...string) (map[string]string, error) {
	argv = normalizeArgv(argv)
	allow := map[string]struct{}{}
	vals := map[string][]string{
		"agent": config.AgentKeys(),
	}
	for _, a := range allowed {
		allow[a] = struct{}{}
	}
	out := map[string]string{}
	for i := 0; i < len(argv); i++ {
		k := argv[i]
		if strings.HasPrefix(k, "-") && len(k) == 2 {
			if full, ok := shortFlags[k]; ok {
				k = "--" + full
			}
		}
		if !strings.HasPrefix(k, "--") {
			return nil, cli.Usage("unexpected argument: " + k)
		}
		name := k[2:]
		if _, ok := allow[name]; !ok {
			return nil, cli.Usage("Options: " + optionsHelp(allowed))
		}
		if i+1 >= len(argv) {
			return nil, cli.Usage("missing value for --" + name)
		}
		v := argv[i+1]
		if !contains(vals[name], v) {
			return nil, cli.Usage(fmt.Sprintf("--%s must be %s.", name, strings.Join(vals[name], " or ")))
		}
		out[name] = v
		i++
	}
	return out, nil
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func samePath(a, b string) bool {
	ra, errA := filepath.EvalSymlinks(a)
	rb, errB := filepath.EvalSymlinks(b)
	if errA != nil {
		ra = a
	}
	if errB != nil {
		rb = b
	}
	ra, _ = filepath.Abs(ra)
	rb, _ = filepath.Abs(rb)
	return filepath.Clean(ra) == filepath.Clean(rb)
}

// cwdUnderSetup is true when cwd is the install root or inside a local setup tree.
func setupProjectCell(r installs.Record, cwd string) string {
	if r.Scope == config.ScopeGlobal {
		return "Global (legacy)\n" + tui.Dim("removed on xdlc update")
	}
	cell := fsutil.Short(r.Path)
	if cwdUnderSetup(r.Path, cwd) {
		cell += tui.Dim(" (current repo)")
	}
	return cell
}

func repositoryID(root string) string {
	id, origin := gitx.RepoIdentity(root)
	if id != "" && !strings.HasPrefix(id, "path:") {
		return id
	}
	if origin != "" {
		if short := gitx.NormalizeRemote(origin); short != "" {
			return short
		}
		return origin
	}
	return fsutil.Short(root)
}

func cwdUnderSetup(setupRoot, cwd string) bool {
	if samePath(setupRoot, cwd) {
		return true
	}
	root, errR := filepath.EvalSymlinks(setupRoot)
	cwdAbs, errC := filepath.EvalSymlinks(cwd)
	if errR != nil {
		root = setupRoot
	}
	if errC != nil {
		cwdAbs = cwd
	}
	root, _ = filepath.Abs(root)
	cwdAbs, _ = filepath.Abs(cwdAbs)
	rel, err := filepath.Rel(root, cwdAbs)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func normalizeArgv(argv []string) []string {
	short := map[string]string{"-a": "--agent"}
	var out []string
	for i := 0; i < len(argv); i++ {
		a := argv[i]
		if v, ok := short[a]; ok {
			out = append(out, v)
			continue
		}
		if k, v, ok := strings.Cut(a, "="); ok && strings.HasPrefix(k, "--") {
			out = append(out, k, v)
			continue
		}
		out = append(out, a)
	}
	return out
}

func optionsHelp(allowed []string) string {
	var parts []string
	for _, n := range allowed {
		short := ""
		for s, f := range shortFlags {
			if f == n {
				short = s + ", "
				break
			}
		}
		parts = append(parts, short+"--"+n)
	}
	return strings.Join(parts, "  ")
}