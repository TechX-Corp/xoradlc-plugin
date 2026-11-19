package usage

import (
	"fmt"
	"os"
	"time"

	"github.com/TechX-Corp/xoradlc-plugin/internal/api"
)

const maxBatch = 500

// Post sends records to POST /usage in batches. On success ACKs those turns from the queue.
// Returns false when not logged in or the API rejected the batch (caller should enqueue).
func Post(recs []Record) (ok bool) {
	if len(recs) == 0 {
		return true
	}
	for i := 0; i < len(recs); i += maxBatch {
		end := i + maxBatch
		if end > len(recs) {
			end = len(recs)
		}
		chunk := recs[i:end]
		body := map[string]any{"records": chunkToMaps(chunk)}
		status, out := api.API("POST", "/usage", body, 30*time.Second)
		if status == -1 {
			fmt.Fprintln(os.Stderr, "xdlc usage: not logged in (xdlc login): queueing")
			return false
		}
		if status == 0 {
			fmt.Fprintf(os.Stderr, "xdlc usage: could not reach API: %v\n", out)
			return false
		}
		if status != 200 {
			fmt.Fprintf(os.Stderr, "xdlc usage: API HTTP %d: %v\n", status, out)
			return false
		}
		_ = AckRemoves(chunk)
	}
	return true
}

func chunkToMaps(recs []Record) []map[string]any {
	out := make([]map[string]any, 0, len(recs))
	for _, r := range recs {
		m := map[string]any{
			"tool":                 r.Tool,
			"session_id":           r.SessionID,
			"turn_id":              r.TurnID,
			"is_subagent":          r.IsSubagent,
			"input_tokens":         r.InputTokens,
			"output_tokens":        r.OutputTokens,
			"cache_read_tokens":    r.CacheReadTokens,
			"cache_write_tokens":   r.CacheWriteTokens,
			"occurred_at":          r.OccurredAt.UTC().Format(time.RFC3339Nano),
			"agent_account_email":  r.AgentAccountEmail,
		}
		if r.Model != nil {
			m["model"] = *r.Model
		}
		if r.Credits != nil {
			m["credits"] = *r.Credits
		}
		out = append(out, m)
	}
	return out
}
