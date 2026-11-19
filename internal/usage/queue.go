package usage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/TechX-Corp/xoradlc-plugin/internal/config"
)

const (
	queueName = "usage-pending.jsonl"
	ttlDays   = 30
)

func queuePath() string {
	return filepath.Join(config.XDLC(), queueName)
}

// Enqueue appends records to the pending queue (preserves turn_id).
func Enqueue(recs []Record) error {
	if len(recs) == 0 {
		return nil
	}
	if err := os.MkdirAll(config.XDLC(), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(queuePath(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	now := time.Now().UTC()
	enc := json.NewEncoder(f)
	for _, r := range recs {
		if err := enc.Encode(Queued{Record: r, QueuedAt: now}); err != nil {
			return err
		}
	}
	return nil
}

// LoadPending reads the queue, drops expired lines (TTL), returns live records.
func LoadPending() ([]Record, error) {
	path := queuePath()
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	cutoff := time.Now().UTC().Add(-ttlDays * 24 * time.Hour)
	var keep []Queued
	var out []Record
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var q Queued
		if err := json.Unmarshal(line, &q); err != nil {
			continue
		}
		if q.QueuedAt.IsZero() {
			q.QueuedAt = time.Now().UTC()
		}
		if q.QueuedAt.Before(cutoff) {
			fmt.Fprintf(os.Stderr, "xdlc usage: dropped expired turn %s (queued %s)\n", q.TurnID, q.QueuedAt.Format(time.RFC3339))
			continue
		}
		keep = append(keep, q)
		out = append(out, q.Record)
	}
	if err := sc.Err(); err != nil {
		return out, err
	}
	// Rewrite file without expired rows.
	if err := rewriteQueue(keep); err != nil {
		return out, err
	}
	return out, nil
}

func rewriteQueue(rows []Queued) error {
	path := queuePath()
	if len(rows) == 0 {
		_ = os.Remove(path)
		return nil
	}
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	for _, q := range rows {
		if err := enc.Encode(q); err != nil {
			f.Close()
			_ = os.Remove(tmp)
			return err
		}
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}

// AckRemoves deletes pending rows whose turn_id (and tool) were accepted.
func AckRemoves(sent []Record) error {
	if len(sent) == 0 {
		return nil
	}
	done := map[string]struct{}{}
	for _, r := range sent {
		done[r.Tool+"|"+r.TurnID] = struct{}{}
	}
	path := queuePath()
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()
	var keep []Queued
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var q Queued
		if err := json.Unmarshal(line, &q); err != nil {
			continue
		}
		key := q.Tool + "|" + q.TurnID
		if _, ok := done[key]; ok {
			continue
		}
		keep = append(keep, q)
	}
	return rewriteQueue(keep)
}

// Cursor paths remember last sent turn watermark per session (high-water mark).
func cursorPath(tool, sessionID string) string {
	safe := tool + "-" + sessionID
	return filepath.Join(config.XDLC(), "usage-cursor", safe+".txt")
}

// LoadCursor returns the last acknowledged turn_id for a session (may be empty).
func LoadCursor(tool, sessionID string) string {
	b, err := os.ReadFile(cursorPath(tool, sessionID))
	if err != nil {
		return ""
	}
	return string(b)
}

// SaveCursor stores the high-water turn_id.
func SaveCursor(tool, sessionID, turnID string) error {
	path := cursorPath(tool, sessionID)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(turnID), 0o600)
}
