package claude

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseNewSkipsUntilCursor(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "t.jsonl")
	content := `{"type":"assistant","timestamp":"2026-11-19T10:00:00Z","message":{"id":"msg_a","model":"claude-sonnet-5","usage":{"input_tokens":1,"output_tokens":2,"cache_read_input_tokens":3,"cache_creation_input_tokens":4}}}
{"type":"assistant","timestamp":"2026-11-19T10:01:00Z","message":{"id":"msg_b","model":"claude-sonnet-5","usage":{"input_tokens":10,"output_tokens":20,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	recs, last, err := Host{}.ParseNew("sess1", path, false, "msg_a")
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 1 || recs[0].TurnID != "msg_b" || last != "msg_b" {
		t.Fatalf("got %+v last=%s", recs, last)
	}
	if recs[0].InputTokens != 10 || recs[0].OutputTokens != 20 {
		t.Fatalf("tokens %+v", recs[0])
	}
}
