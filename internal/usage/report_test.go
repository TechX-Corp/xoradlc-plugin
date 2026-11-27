package usage

import (
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/TechX-Corp/xoradlc-plugin/internal/gitx"
	"github.com/TechX-Corp/xoradlc-plugin/internal/link"
)

func TestHookPlan(t *testing.T) {
	sess := func(ev string) HookPayload { return HookPayload{HookEventName: ev, SessionID: "s1"} }
	cases := []struct {
		name                  string
		hook                  string
		payload               HookPayload
		flush, deferred       bool
		wantFlush, wantDetach bool
	}{
		{"kiro stop goes to background", "kiro", sess("Stop"), false, false, false, true},
		{"deferred child parses in place", "kiro", sess("Stop"), false, true, false, false},
		{"claude stop parses now", "claude", sess("Stop"), false, false, false, false},
		{"session end parses despite --flush", "kiro", sess("SessionEnd"), true, false, false, false},
		{"claude session end parses too", "claude", sess("SessionEnd"), true, false, false, false},
		{"session start only flushes", "kiro", sess("SessionStart"), true, false, true, false},
		{"no session only flushes", "kiro", HookPayload{HookEventName: "Stop"}, false, false, true, false},
		{"--flush on another event", "claude", sess("Notification"), true, false, true, false},
	}
	for _, c := range cases {
		flush, detach := hookPlan(c.hook, c.payload, c.flush, c.deferred)
		if flush != c.wantFlush || detach != c.wantDetach {
			t.Errorf("%s: got flush=%v background=%v", c.name, flush, detach)
		}
	}
}

type lateHost struct{ calls int }

func (h *lateHost) Tool() string        { return ToolKiro }
func (h *lateHost) AgentEmail() *string { return nil }
func (h *lateHost) ParseNew(sessionID, _ string, _ bool, after string) ([]Record, string, error) {
	h.calls++
	if h.calls < 2 {
		return nil, after, nil
	}
	c := 0.25
	return []Record{{Tool: ToolKiro, SessionID: sessionID, TurnID: "t1", Credits: &c, OccurredAt: time.Now()}}, "t1", nil
}

func TestReportWhenSettledWaitsForTheLateSummary(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	old := settleDelays
	settleDelays = []time.Duration{time.Millisecond, time.Millisecond, time.Millisecond}
	defer func() { settleDelays = old }()

	h := &lateHost{}
	if err := reportWhenSettled(h, HookPayload{HookEventName: "Stop", SessionID: "s1"}); err != nil {
		t.Fatal(err)
	}
	// Not logged in under the temp HOME: the turn found on the second parse lands in the queue.
	pending, err := LoadPending()
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 1 || pending[0].TurnID != "t1" {
		t.Fatalf("pending = %+v", pending)
	}
	if h.calls != 3 { // two polls until the summary shows up, then Report's own parse
		t.Fatalf("ParseNew calls = %d", h.calls)
	}
}

func TestReportStampsTheWorkOrderOfTheLinkedBranch(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	repo := t.TempDir()
	for _, args := range [][]string{{"init", "-q", "-b", "feat/x"}, {"commit", "-q", "--allow-empty", "-m", "init"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	repoID, _ := gitx.RepoIdentity(gitx.GitRoot(repo))
	if err := link.SaveLink(repoID, "feat/x", map[string]string{"work_order_id": "WO-2026-031", "step_id": "impl"}); err != nil {
		t.Fatal(err)
	}

	h := &lateHost{calls: 1}
	if err := Report(h, HookPayload{HookEventName: "Stop", SessionID: "s1", Cwd: repo}, false); err != nil {
		t.Fatal(err)
	}
	pending, err := LoadPending()
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 1 || pending[0].WorkOrderID == nil || *pending[0].WorkOrderID != "WO-2026-031" {
		t.Fatalf("pending = %+v", pending)
	}
	if wo := workOrderFor(t.TempDir()); wo != nil { // not a repo: no WO
		t.Fatalf("workOrderFor(non-repo) = %q", *wo)
	}
}
