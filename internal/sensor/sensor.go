package sensor

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/TechX-Corp/xoradlc-plugin/internal/api"
	"github.com/TechX-Corp/xoradlc-plugin/internal/cli"
	"github.com/TechX-Corp/xoradlc-plugin/internal/config"
	"github.com/TechX-Corp/xoradlc-plugin/internal/gitx"
	"github.com/TechX-Corp/xoradlc-plugin/internal/link"
)

// CmdSensor implements xdlc sensor and xdlc sensor report [--pre-push].
func CmdSensor(argv []string) error {
	if len(argv) > 0 && argv[0] == "--pre-push" {
		fmt.Println("xdlc: convention checks run in the agent on commit/push, not in git pre-push.")
		return nil
	}
	if len(argv) > 0 && argv[0] == "report" {
		return cmdReport(argv[1:])
	}
	if len(argv) > 0 {
		return cli.Usage("xdlc sensor | xdlc sensor report [--file PATH]")
	}
	cwd, _ := os.Getwd()
	root := gitx.GitRoot(cwd)
	if root == "" {
		root = cwd
	}
	text, err := api.AskAAP(api.ConventionAskPrompt, 120*time.Second)
	if err != nil {
		return err
	}
	fmt.Println("Convention (from AAP / KB):")
	fmt.Println(strings.TrimRight(text, "\n"))
	fmt.Println()
	if gitx.GitRoot(cwd) != "" {
		files := gitx.ChangedFiles(gitx.GitRoot(cwd))
		if len(files) > 0 {
			fmt.Println("Changed files:")
			for _, f := range files {
				fmt.Printf("  %s\n", f)
			}
		} else {
			fmt.Println("Changed files: (none detected)")
		}
	} else {
		fmt.Println("Changed files: (none detected)")
	}
	return nil
}

func cmdReport(argv []string) error {
	var raw []byte
	var err error
	if len(argv) >= 2 && argv[0] == "--file" {
		raw, err = os.ReadFile(argv[1])
	} else if len(argv) > 0 {
		return cli.Usage("xdlc sensor report [--file PATH]  (JSON findings on stdin if no --file)")
	} else {
		raw, err = io.ReadAll(os.Stdin)
	}
	if err != nil {
		return cli.Fail(err.Error())
	}
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		return cli.Fail(fmt.Sprintf("Invalid findings JSON: %v", err))
	}
	found, _ := data["findings"].([]any)
	if len(found) == 0 {
		return cli.Fail("findings must be a non-empty list.")
	}
	cwd, _ := os.Getwd()
	root := gitx.GitRoot(cwd)
	if root == "" {
		return cli.Fail("Not a git repo: cd into the repo and run again.")
	}
	branch := gitx.GitOut(root, "branch", "--show-current")
	if branch == "" {
		branch = "(detached)"
	}
	repoID, origin := gitx.RepoIdentity(root)
	st, ok := link.LinkOf(repoID, branch)
	if !ok {
		return cli.Fail(fmt.Sprintf("Branch %s has no work order mapped: run xdlc link.", branch))
	}
	commit := gitx.GitOut(root, "rev-parse", "HEAD")
	if len(commit) > 64 {
		commit = commit[:64]
	}
	now := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	var findings []map[string]any
	for _, item := range found {
		if len(findings) >= 500 {
			break
		}
		x, _ := item.(map[string]any)
		if x == nil {
			continue
		}
		sev := fmt.Sprint(x["severity"])
		switch sev {
		case "🔴", "red":
			sev = "red"
		default:
			sev = "yellow"
		}
		file := fmt.Sprint(x["file"])
		if file == "" {
			file = ""
		}
		rc := fmt.Sprint(x["rule_code"])
		if rc == "" || rc == "<nil>" {
			rc = "AAP-NL"
		}
		findings = append(findings, map[string]any{
			"file":      trunc(file, 500),
			"line":      x["line"],
			"rule_code": trunc(rc, 32),
			"severity":  sev,
			"message":   trunc(fmt.Sprint(x["message"]), 300),
			"at":        now,
		})
	}
	if len(findings) == 0 || findings[0]["file"] == "" {
		return cli.Fail("Each finding needs file, rule_code, severity, message.")
	}
	host, _ := os.Hostname()
	repo := origin
	if repo == "" {
		repo = repoID
	}
	body := map[string]any{
		"work_order_id": st["work_order_id"],
		"step_id":       st["step_id"],
		"machine":       trunc(host, 120),
		"session_id":    trunc("agent-push:"+branch, 120),
		"repo":          trunc(repo, 300),
		"commit":        commit,
		"findings":      findings,
	}
	status, out := api.API("POST", "/governance/sensor-events", body, 30*time.Second)
	stepName := st["step_name"]
	if stepName == "" {
		stepName = st["step_id"]
	}
	where := fmt.Sprintf("%s · %s", st["work_order_id"], stepName)
	if status == -1 {
		return cli.Usage("Not logged in to Xora DLC (xdlc login): not sending alerts.")
	}
	if status == 201 {
		fmt.Printf("xdlc: sent %d alert(s) to Xora DLC (%s).\n", len(findings), where)
		return nil
	}
	if status == 403 {
		return cli.Fail(fmt.Sprintf("You are not a Dev on %s: remap with xdlc link.", where))
	}
	if status == 0 {
		return cli.Fail(fmt.Sprintf("Could not connect to %s (%v).", config.BaseURL(), out))
	}
	detail := out
	if m, ok := out.(map[string]any); ok {
		if errObj, ok := m["error"].(map[string]any); ok {
			detail = errObj["detail"]
		}
	}
	return cli.Fail(fmt.Sprintf("Xora DLC rejected (HTTP %d): %v", status, detail))
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}