package app

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/TechX-Corp/xoradlc-plugin/internal/cli"
	"github.com/TechX-Corp/xoradlc-plugin/internal/config"
	"github.com/TechX-Corp/xoradlc-plugin/internal/gitx"
	"github.com/TechX-Corp/xoradlc-plugin/internal/installs"
	"github.com/TechX-Corp/xoradlc-plugin/internal/link"
	"github.com/TechX-Corp/xoradlc-plugin/internal/tui"
)

func cmdView(argv []string) error {
	if len(argv) != 1 {
		return cli.Usage("view takes exactly one setup id (see xdlc status).")
	}
	items := installs.Load()
	rec, ok := installs.ByID(items, argv[0])
	if !ok {
		return cli.Fail(fmt.Sprintf("unknown setup id %q. Check ids with xdlc status.", argv[0]))
	}
	cwd, _ := os.Getwd()
	printSetupView(rec, cwd)
	return nil
}

func printSetupView(r installs.Record, cwd string) {
	agent := config.Agents[r.Agent]
	if agent == "" {
		agent = r.Agent
	}
	fmt.Printf("%s %s\n", tui.Bold("Setup ID:"), tui.Em(r.ID))
	fmt.Printf("%s %s\n", tui.Bold("Agent:"), tui.Em(agent))
	fmt.Printf("%s %s\n", tui.Bold("Scope:"), setupScopeInline(r, cwd))
	updated := tui.Dim("unknown")
	if r.UpdatedAt > 0 {
		updated = time.Unix(r.UpdatedAt, 0).Format("2006-01-02 15:04")
	}
	fmt.Printf("%s %s\n", tui.Bold("Updated:"), updated)
	fmt.Printf("%s %s\n", tui.Bold("Repo:"), repoViewLine(r))
	fmt.Printf("%s %s\n", tui.Bold("Work order:"), workOrderViewLine(r))
}

func setupScopeInline(r installs.Record, cwd string) string {
	scope := config.ScopeName[r.Scope]
	if scope == "" {
		scope = r.Scope
	}
	if r.Scope == "global" {
		return scope + " · " + tui.Dim("all repositories on this machine")
	}
	line2 := setupScopeCell(r, cwd)
	if i := strings.IndexByte(line2, '\n'); i >= 0 {
		line2 = line2[i+1:]
	}
	return scope + " · " + line2
}

func repoViewLine(r installs.Record) string {
	if r.Scope == "global" {
		return tui.Dim("not a repo")
	}
	root := gitx.GitRoot(r.Path)
	if root == "" {
		return tui.Dim("not a git repository")
	}
	repo := repositoryID(root)
	branch := branchAtSetupRoot(root)
	if branch == "" {
		return tui.Em(repo) + " · " + tui.Dim("branch ") + "—"
	}
	return tui.Em(repo) + " · " + tui.Dim("branch ") + tui.Em(branch)
}

func workOrderViewLine(r installs.Record) string {
	if r.Scope == "global" {
		return tui.Dim("map in repo: ") + tui.Em("xdlc link -l")
	}
	root := gitx.GitRoot(r.Path)
	if root == "" {
		return tui.Dim("—")
	}
	repoID, _ := gitx.RepoIdentity(root)
	branch := branchAtSetupRoot(root)
	if branch == "" {
		return tui.Dim("not mapped") + " · " + tui.Dim("branch unknown")
	}
	if label := link.MappingLabel(repoID, branch); label != "" {
		return tui.Em(label)
	}
	return tui.Dim("not mapped") + " · " + tui.Em("xdlc link")
}

func branchAtSetupRoot(gitRoot string) string {
	return gitx.GitOut(gitRoot, "branch", "--show-current")
}