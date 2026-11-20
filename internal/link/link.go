package link

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TechX-Corp/xoradlc-plugin/internal/api"
	"github.com/TechX-Corp/xoradlc-plugin/internal/cli"
	"github.com/TechX-Corp/xoradlc-plugin/internal/config"
	"github.com/TechX-Corp/xoradlc-plugin/internal/fsutil"
	"github.com/TechX-Corp/xoradlc-plugin/internal/gitx"
	"github.com/TechX-Corp/xoradlc-plugin/internal/tui"
)

type step map[string]string

func linksPath() string {
	return filepath.Join(config.XDLC(), "links.json")
}

// Links loads branch→work-order maps.
func Links() map[string]map[string]step {
	var data map[string]any
	_ = fsutil.LoadJSON(linksPath(), &data)
	if data == nil {
		return map[string]map[string]step{}
	}
	if _, ok := data["prod"]; ok {
		pm, _ := data["prod"].(map[string]any)
		return convertLinks(pm)
	}
	return convertLinks(data)
}

func convertLinks(data map[string]any) map[string]map[string]step {
	out := map[string]map[string]step{}
	for repo, branches := range data {
		bm, _ := branches.(map[string]any)
		if bm == nil {
			continue
		}
		out[repo] = map[string]step{}
		for branch, st := range bm {
			sm, _ := st.(map[string]any)
			if sm == nil {
				continue
			}
			out[repo][branch] = step{
				"work_order_id":   fmt.Sprint(sm["work_order_id"]),
				"work_order_name": fmt.Sprint(sm["work_order_name"]),
				"step_id":         fmt.Sprint(sm["step_id"]),
				"step_name":       fmt.Sprint(sm["step_name"]),
			}
		}
	}
	return out
}

// LinkOf returns mapped step for repo/branch.
func LinkOf(repoID, branch string) (step, bool) {
	s, ok := Links()[repoID][branch]
	return s, ok
}

// SaveLink stores branch mapping.
func SaveLink(repoID, branch string, st step) error {
	data := map[string]any{}
	_ = fsutil.LoadJSON(linksPath(), &data)
	if data == nil {
		data = map[string]any{}
	}
	branches, _ := data[repoID].(map[string]any)
	if branches == nil {
		branches = map[string]any{}
		data[repoID] = branches
	}
	branches[branch] = map[string]any{
		"work_order_id":   st["work_order_id"],
		"work_order_name": st["work_order_name"],
		"step_id":         st["step_id"],
		"step_name":       st["step_name"],
	}
	return fsutil.SaveJSON(linksPath(), data, true)
}

func stepLabel(s step) string {
	wn := s["work_order_name"]
	if wn == "" {
		wn = ""
	}
	sn := s["step_name"]
	if sn == "" {
		sn = s["step_id"]
	}
	return fmt.Sprintf("%s · %s · %s", s["work_order_id"], wn, sn)
}

// MappingLabel returns the linked work-order line for repo/branch, or "" when unmapped.
func MappingLabel(repoID, branch string) string {
	st, ok := LinkOf(repoID, branch)
	if !ok {
		return ""
	}
	return stepLabel(st)
}

// CmdLink implements xdlc link.
func CmdLink(argv []string) error {
	argv = longOpts(argv)
	var wo, stepID, branch string
	listOnly := false
	for i := 0; i < len(argv); i++ {
		a := argv[i]
		switch a {
		case "--list":
			listOnly = true
		case "--step", "-t":
			i++
			if i < len(argv) {
				stepID = argv[i]
			}
		case "--branch", "-b":
			i++
			if i < len(argv) {
				branch = argv[i]
			}
		default:
			if strings.HasPrefix(a, "-") {
				return cli.Usage("xdlc link [WO-…] [-t, --step STEP] [-b, --branch BRANCH] [-l, --list]")
			}
			if wo == "" {
				wo = a
			} else {
				return cli.Usage("xdlc link [WO-…] [-t, --step STEP] [-b, --branch BRANCH] [-l, --list]")
			}
		}
	}
	cwd, _ := os.Getwd()
	root := gitx.GitRoot(cwd)
	if root == "" {
		return cli.Fail("Not a git repo: cd into the repo and run again.")
	}
	if branch == "" {
		branch = gitx.GitOut(root, "branch", "--show-current")
	}
	if branch == "" {
		return cli.Fail("Not on a branch (detached HEAD): add --branch <name>.")
	}
	repoID, _ := gitx.RepoIdentity(root)
	cur, hasCur := LinkOf(repoID, branch)
	steps, err := api.MySteps()
	if err != nil {
		return err
	}
	if listOnly {
		mapped := tui.Dim("no work order mapped")
		if hasCur {
			mapped = stepLabel(stepFromAPI(cur))
		}
		fmt.Printf("%s %s (%s): %s\n", tui.Bold("Branch"), tui.Em(branch), tui.Em(repoID), mapped)
		fmt.Println()
		if len(steps) > 0 {
			fmt.Println(tui.Dim("Your in-progress steps (") + tui.Em("xdlc link <WO>") + tui.Dim(" ") +
				tui.Em("--step") + tui.Dim(" <step>):"))
			for i, st := range steps {
				meta := strings.Trim(strings.Join([]string{st.WorkOrderName, st.StepName}, " · "), " ·")
				line := fmt.Sprintf("  %d. %s %s %s", i+1, tui.Em(st.WorkOrderID), tui.Em("--step"), tui.Em(st.StepID))
				if meta != "" {
					line += "   " + meta
				}
				fmt.Println(line)
			}
		} else {
			fmt.Println(tui.Dim("You have no in-progress steps."))
		}
		return nil
	}
	var chosen api.Step
	if wo != "" {
		var cands []api.Step
		for _, st := range steps {
			if strings.EqualFold(st.WorkOrderID, wo) && (stepID == "" || st.StepID == stepID) {
				cands = append(cands, st)
			}
		}
		if len(cands) == 0 {
			suffix := ""
			if stepID != "" {
				suffix = " / " + stepID
			}
			return cli.Fail(fmt.Sprintf("%s%s is not among your in-progress steps. See: xdlc link --list", wo, suffix))
		}
		if len(cands) > 1 {
			var ids []string
			for _, c := range cands {
				ids = append(ids, c.StepID)
			}
			return cli.Fail(fmt.Sprintf("%s has multiple in-progress steps; add --step: %s", wo, strings.Join(ids, " | ")))
		}
		chosen = cands[0]
	} else {
		if len(steps) == 0 {
			return cli.Fail("You have no in-progress steps in Xora DLC: nothing to map.")
		}
		labels := make([]string, len(steps))
		for i, st := range steps {
			labels[i] = api.StepLabel(st)
		}
		idx, err := tui.Menu(fmt.Sprintf("Map branch %s to which step? (My Work)", branch), labels, false)
		if err != nil {
			return err
		}
		chosen = steps[idx[0]]
	}
	st := step{
		"work_order_id": chosen.WorkOrderID, "work_order_name": chosen.WorkOrderName,
		"step_id": chosen.StepID, "step_name": chosen.StepName,
	}
	if err := SaveLink(repoID, branch, st); err != nil {
		return err
	}
	was := ""
	if hasCur && (cur["work_order_id"] != chosen.WorkOrderID || cur["step_id"] != chosen.StepID) {
		was = tui.Dim(fmt.Sprintf(" (was: %s)", cur["work_order_id"]))
	}
	fmt.Printf("%s branch %s → %s%s\n", tui.Ok("Mapped"), tui.Em(branch), tui.Em(api.StepLabel(chosen)), was)
	return nil
}

func stepFromAPI(s step) step { return s }

var shortFlags = map[string]string{
	"-t": "--step", "-b": "--branch", "-l": "--list",
}

func longOpts(argv []string) []string {
	out := make([]string, len(argv))
	for i, a := range argv {
		if v, ok := shortFlags[a]; ok {
			out[i] = v
		} else {
			out[i] = a
		}
	}
	return out
}