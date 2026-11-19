package kiro

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseUsageSummary(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "messages.jsonl")
	content := `{"timestamp":"2026-11-16T10:15:23.242Z","payload":{"type":"usage_summary","promptTurnSummaries":[{"unit":"credit","usage":0.18}],"executionId":"exec_1","turnId":"turn_1","status":"success"}}
{"timestamp":"2026-11-16T10:16:00Z","payload":{"type":"usage_summary","promptTurnSummaries":[{"unit":"credit","usage":0.05}],"turnId":"turn_2","status":"success"}}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(dir, "session.json"), []byte(`{"modelId":"auto"}`), 0o644)
	recs, last, err := Host{}.ParseNew("sess_x", path, false, "turn_1")
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 1 || recs[0].TurnID != "turn_2" || last != "turn_2" {
		t.Fatalf("got %+v last=%s", recs, last)
	}
	if recs[0].Credits == nil || *recs[0].Credits < 0.049 || *recs[0].Credits > 0.051 {
		t.Fatalf("credits %+v", recs[0].Credits)
	}
}
