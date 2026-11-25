package usage

import (
	"testing"
	"time"
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
