package usage

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
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
		path := payload.TranscriptPath
		sub := false
		ev := strings.ToLower(payload.HookEventName)
		if strings.Contains(ev, "subagent") && payload.AgentTranscriptPath != "" {
			path = payload.AgentTranscriptPath
			sub = true
		}
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
	if !flushOnly {
		p, err := ReadHookStdin(os.Stdin)
		if err != nil && err != io.EOF {
			// SessionStart may have empty/non-usage stdin — treat as flush.
			flushOnly = true
		} else {
			payload = p
		}
	}
	ev := strings.ToLower(payload.HookEventName)
	if strings.Contains(ev, "sessionstart") || strings.Contains(ev, "session_start") {
		flushOnly = true
	}
	return Report(h, payload, flushOnly)
}
