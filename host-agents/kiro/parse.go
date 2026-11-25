package kiro

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/TechX-Corp/xoradlc-plugin/internal/usage"
)

// Host implements usage.Host for Kiro CLI.
type Host struct{}

func (Host) Tool() string { return usage.ToolKiro }

func (Host) AgentEmail() *string { return AgentEmail() }

// ParseNew reads messages.jsonl (or resolves sess dir). afterTurnID is last sent turnId.
// transcriptPath from hook may be session dir, messages.jsonl, or empty (Kiro V3 sends none:
// the session is found under ~/.kiro/sessions by id).
func (Host) ParseNew(sessionID, transcriptPath string, subagent bool, afterTurnID string) ([]usage.Record, string, error) {
	path := resolveMessages(transcriptPath, sessionID)
	if path == "" {
		return nil, afterTurnID, nil
	}
	model := readModel(filepath.Dir(path))

	f, err := os.Open(path)
	if err != nil {
		return nil, afterTurnID, err
	}
	defer f.Close()

	type summary struct {
		Unit  string  `json:"unit"`
		Usage float64 `json:"usage"`
	}
	type payload struct {
		Type               string    `json:"type"`
		TurnID             string    `json:"turnId"`
		ExecutionID        string    `json:"executionId"`
		PromptTurnSummaries []summary `json:"promptTurnSummaries"`
		Status             string    `json:"status"`
	}
	type entry struct {
		Timestamp string  `json:"timestamp"`
		Payload   payload `json:"payload"`
	}

	seenAfter := afterTurnID == ""
	var recs []usage.Record
	lastID := afterTurnID
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var e entry
		if err := json.Unmarshal(line, &e); err != nil {
			continue
		}
		if e.Payload.Type != "usage_summary" {
			continue
		}
		turnID := e.Payload.TurnID
		if turnID == "" {
			turnID = e.Payload.ExecutionID
		}
		if turnID == "" {
			continue
		}
		if !seenAfter {
			if turnID == afterTurnID {
				seenAfter = true
			}
			continue
		}
		var credits float64
		for _, s := range e.Payload.PromptTurnSummaries {
			if strings.EqualFold(s.Unit, "credit") || s.Unit == "" {
				credits += s.Usage
			}
		}
		at := time.Now().UTC()
		if e.Timestamp != "" {
			if t, err := time.Parse(time.RFC3339Nano, e.Timestamp); err == nil {
				at = t.UTC()
			} else if t, err := time.Parse(time.RFC3339, e.Timestamp); err == nil {
				at = t.UTC()
			}
		}
		c := credits
		recs = append(recs, usage.Record{
			Tool:       usage.ToolKiro,
			SessionID:  sessionID,
			TurnID:     turnID,
			IsSubagent: subagent,
			Model:      model,
			Credits:    &c,
			OccurredAt: at,
		})
		lastID = turnID
	}
	return recs, lastID, sc.Err()
}

func resolveMessages(transcriptPath, sessionID string) string {
	if transcriptPath != "" {
		st, err := os.Stat(transcriptPath)
		if err == nil {
			if st.IsDir() {
				p := filepath.Join(transcriptPath, "messages.jsonl")
				if _, err := os.Stat(p); err == nil {
					return p
				}
			} else if strings.HasSuffix(transcriptPath, "messages.jsonl") {
				return transcriptPath
			}
		}
	}
	if sessionID == "" {
		return ""
	}
	// Search ~/.kiro/sessions/**/sess_<id>/messages.jsonl
	root := filepath.Join(os.Getenv("HOME"), ".kiro", "sessions")
	if h, err := os.UserHomeDir(); err == nil {
		root = filepath.Join(h, ".kiro", "sessions")
	}
	var found string
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if d.Name() != "messages.jsonl" {
			return nil
		}
		if strings.Contains(path, sessionID) {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

func readModel(sessDir string) *string {
	b, err := os.ReadFile(filepath.Join(sessDir, "session.json"))
	if err != nil {
		return nil
	}
	var doc map[string]any
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil
	}
	if m, ok := doc["modelId"].(string); ok && m != "" {
		return &m
	}
	return nil
}
