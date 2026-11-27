package usage

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

// HookPayload is the JSON stdin from Claude/Kiro Stop-like hooks.
type HookPayload struct {
	HookEventName       string `json:"hook_event_name"`
	SessionID           string `json:"session_id"`
	TranscriptPath      string `json:"transcript_path"`
	AgentTranscriptPath string `json:"agent_transcript_path"`
	Cwd                 string `json:"cwd"`
}

// Host parses turns and resolves agent email for one IDE agent.
type Host interface {
	Tool() string
	AgentEmail() *string
	// ParseNew returns turns after cursor (high-water). subagent from SubagentStop path.
	ParseNew(sessionID, transcriptPath string, subagent bool, afterTurnID string) ([]Record, string, error)
}

// Report runs flush of queue + parse new turns for a host event.
func Report(h Host, payload HookPayload, flushOnly bool) error {
	email := h.AgentEmail()
	pending, err := LoadPending()
	if err != nil {
		fmt.Fprintf(os.Stderr, "xdlc usage: queue read: %v\n", err)
	}
	applyEmail(pending, email)

	var fresh []Record
	lastCursor := ""
	if !flushOnly && payload.SessionID != "" {
		path, sub := transcriptFor(payload)
		ev := strings.ToLower(payload.HookEventName)
		after := LoadCursor(h.Tool(), payload.SessionID)
		var newCursor string
		fresh, newCursor, err = h.ParseNew(payload.SessionID, path, sub, after)
		if err != nil {
			fmt.Fprintf(os.Stderr, "xdlc usage: parse: %v\n", err)
		}
		applyEmail(fresh, email)
		lastCursor = newCursor
		// B+C: brief retry if Stop and nothing new but transcript may lag.
		if len(fresh) == 0 && path != "" && (ev == "stop" || strings.HasSuffix(ev, "stop")) {
			time.Sleep(300 * time.Millisecond)
			fresh, newCursor, err = h.ParseNew(payload.SessionID, path, sub, after)
			if err == nil {
				applyEmail(fresh, email)
				if newCursor != "" {
					lastCursor = newCursor
				}
			}
		}
	}
	if len(fresh) > 0 {
		applyWorkOrder(fresh, workOrderFor(payload.Cwd))
	}

	all := append(pending, fresh...)
	all = dedupe(all)
	if len(all) == 0 {
		return nil
	}
	if Post(all) {
		if lastCursor != "" && payload.SessionID != "" {
			_ = SaveCursor(h.Tool(), payload.SessionID, lastCursor)
		}
		return nil
	}
	// Only enqueue fresh (pending already on disk).
	if len(fresh) > 0 {
		_ = Enqueue(fresh)
	}
	return nil
}

func transcriptFor(payload HookPayload) (path string, subagent bool) {
	ev := strings.ToLower(payload.HookEventName)
	if strings.Contains(ev, "subagent") && payload.AgentTranscriptPath != "" {
		return payload.AgentTranscriptPath, true
	}
	return payload.TranscriptPath, false
}

// settleDelays: Kiro V3 writes a turn's usage_summary only after its Stop hooks have returned.
var settleDelays = []time.Duration{time.Second, 2 * time.Second, 4 * time.Second}

// reportWhenSettled waits until the transcript has a turn past the cursor (or gives up), then reports.
func reportWhenSettled(h Host, payload HookPayload) error {
	path, sub := transcriptFor(payload)
	after := LoadCursor(h.Tool(), payload.SessionID)
	for _, d := range settleDelays {
		time.Sleep(d)
		if recs, _, err := h.ParseNew(payload.SessionID, path, sub, after); err == nil && len(recs) > 0 {
			break
		}
	}
	return Report(h, payload, false)
}

func isEvent(ev string, names ...string) bool {
	ev = strings.ToLower(strings.ReplaceAll(ev, "_", ""))
	for _, n := range names {
		if ev == n {
			return true
		}
	}
	return false
}

// hookPlan decides what one hook invocation does: flush the queue only, parse the transcript now,
// or hand the parse to a detached child (Kiro Stop, whose usage_summary lands after the hook).
func hookPlan(hook string, payload HookPayload, flushFlag, deferred bool) (flushOnly, background bool) {
	ev := payload.HookEventName
	switch {
	case payload.SessionID == "", isEvent(ev, "sessionstart"):
		return true, false
	case isEvent(ev, "sessionend"):
		return false, false // last turn of the session: no later Stop will pick it up
	case flushFlag:
		return true, false
	case hook == "kiro" && isEvent(ev, "stop") && !deferred:
		return false, true
	}
	return false, false
}

func applyEmail(recs []Record, email *string) {
	for i := range recs {
		if recs[i].AgentAccountEmail == nil {
			recs[i].AgentAccountEmail = email
		}
	}
}

func dedupe(recs []Record) []Record {
	seen := map[string]struct{}{}
	var out []Record
	for _, r := range recs {
		key := r.Tool + "|" + r.TurnID
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, r)
	}
	return out
}

// ReadHookStdin parses hook JSON from r (usually os.Stdin).
func ReadHookStdin(r io.Reader) (HookPayload, error) {
	var p HookPayload
	dec := json.NewDecoder(r)
	if err := dec.Decode(&p); err != nil {
		return p, err
	}
	return p, nil
}

// CmdUsage is `xdlc usage [--hook claude|kiro] [--flush]`.
func CmdUsage(argv []string, hosts map[string]Host) error {
	hook := ""
	flushOnly := false
	deferred := false
	for i := 0; i < len(argv); i++ {
		switch argv[i] {
		case "--hook":
			if i+1 >= len(argv) {
				return fmt.Errorf("usage: xdlc usage --hook claude|kiro")
			}
			i++
			hook = strings.ToLower(argv[i])
		case "--flush":
			flushOnly = true
		case deferredFlag:
			deferred = true
		case "-h", "--help":
			fmt.Println("xdlc usage [--hook claude|kiro] [--flush]\n\nReport Claude/Kiro turn usage to Xora DLC POST /usage.")
			return nil
		}
	}
	if hook == "" {
		// Flush queue only (no host parse).
		pending, err := LoadPending()
		if err != nil {
			return err
		}
		if len(pending) == 0 {
			fmt.Println("xdlc usage: nothing pending.")
			return nil
		}
		if Post(pending) {
			fmt.Printf("xdlc usage: sent %d pending record(s).\n", len(pending))
		} else {
			fmt.Println("xdlc usage: could not send; left in queue.")
		}
		return nil
	}
	h, ok := hosts[hook]
	if !ok {
		return fmt.Errorf("unknown --hook %q (claude|kiro)", hook)
	}
	var payload HookPayload
	if deferred {
		_ = json.Unmarshal([]byte(os.Getenv(payloadEnv)), &payload)
	} else if !stdinIsTerminal() {
		// Empty or non-JSON stdin leaves the payload without a session: flush only.
		payload, _ = ReadHookStdin(os.Stdin)
	}
	flushOnly, background := hookPlan(hook, payload, flushOnly, deferred)
	if background {
		if err := spawnDeferred(hook, payload); err == nil {
			return nil
		}
		// Waiting here would only hold Kiro back from writing the summary: report what is there.
		return Report(h, payload, false)
	}
	if deferred {
		return reportWhenSettled(h, payload)
	}
	return Report(h, payload, flushOnly)
}

const (
	deferredFlag = "--deferred"
	payloadEnv   = "XDLC_USAGE_PAYLOAD"
)

// spawnDeferred re-runs `xdlc usage --hook <hook> --deferred` detached, so the hook returns at once.
func spawnDeferred(hook string, payload HookPayload) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	cmd := exec.Command(exe, "usage", "--hook", hook, deferredFlag)
	cmd.Env = append(os.Environ(), payloadEnv+"="+string(b))
	detach(cmd)
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}

func stdinIsTerminal() bool {
	st, err := os.Stdin.Stat()
	return err == nil && st.Mode()&os.ModeCharDevice != 0
}
