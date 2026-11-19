package claude

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/TechX-Corp/xoradlc-plugin/internal/usage"
)

// Host implements usage.Host for Claude Code.
type Host struct{}

func (Host) Tool() string { return usage.ToolClaude }

func (Host) AgentEmail() *string { return AgentEmail() }

// ParseNew reads transcript JSONL and returns assistant turns after afterTurnID.
func (Host) ParseNew(sessionID, transcriptPath string, subagent bool, afterTurnID string) ([]usage.Record, string, error) {
	if transcriptPath == "" {
		return nil, afterTurnID, nil
	}
	f, err := os.Open(transcriptPath)
	if err != nil {
		return nil, afterTurnID, err
	}
	defer f.Close()

	type msgUsage struct {
		InputTokens              int64 `json:"input_tokens"`
		OutputTokens             int64 `json:"output_tokens"`
		CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
		CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
	}
	type message struct {
		ID    string    `json:"id"`
		Model string    `json:"model"`
		Usage *msgUsage `json:"usage"`
	}
	type entry struct {
		Type      string          `json:"type"`
		Timestamp string          `json:"timestamp"`
		Message   json.RawMessage `json:"message"`
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
		if e.Type != "assistant" || len(e.Message) == 0 {
			continue
		}
		var msg message
		if err := json.Unmarshal(e.Message, &msg); err != nil || msg.ID == "" {
			continue
		}
		if !seenAfter {
			if msg.ID == afterTurnID {
				seenAfter = true
			}
			continue
		}
		if msg.Usage == nil {
			continue
		}
		if strings.HasPrefix(msg.Model, "<") { // synthetic
			continue
		}
		at := time.Now().UTC()
		if e.Timestamp != "" {
			if t, err := time.Parse(time.RFC3339Nano, e.Timestamp); err == nil {
				at = t.UTC()
			} else if t, err := time.Parse(time.RFC3339, e.Timestamp); err == nil {
				at = t.UTC()
			}
		}
		var model *string
		if msg.Model != "" {
			m := msg.Model
			model = &m
		}
		recs = append(recs, usage.Record{
			Tool:             usage.ToolClaude,
			SessionID:        sessionID,
			TurnID:           msg.ID,
			IsSubagent:       subagent,
			Model:            model,
			InputTokens:      msg.Usage.InputTokens,
			OutputTokens:     msg.Usage.OutputTokens,
			CacheReadTokens:  msg.Usage.CacheReadInputTokens,
			CacheWriteTokens: msg.Usage.CacheCreationInputTokens,
			OccurredAt:       at,
		})
		lastID = msg.ID
	}
	return recs, lastID, sc.Err()
}
