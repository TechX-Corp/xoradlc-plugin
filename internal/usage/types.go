package usage

import "time"

// Tool values match POST /usage (UPPER_SNAKE).
const (
	ToolClaude = "CLAUDE_CODE"
	ToolKiro   = "KIRO_CLI"
)

// Record is one model turn for POST /usage.
type Record struct {
	Tool               string     `json:"tool"`
	SessionID          string     `json:"session_id"`
	TurnID             string     `json:"turn_id"`
	IsSubagent         bool       `json:"is_subagent"`
	Model              *string    `json:"model"`
	InputTokens        int64      `json:"input_tokens"`
	OutputTokens       int64      `json:"output_tokens"`
	CacheReadTokens    int64      `json:"cache_read_tokens"`
	CacheWriteTokens   int64      `json:"cache_write_tokens"`
	Credits            *float64   `json:"credits"`
	AgentAccountEmail  *string    `json:"agent_account_email"`
	WorkOrderID        *string    `json:"work_order_id,omitempty"` // WO code from `xdlc link`
	OccurredAt         time.Time  `json:"occurred_at"`
}

// Queued is a pending queue line (same as Record plus queued_at for TTL).
type Queued struct {
	Record
	QueuedAt time.Time `json:"queued_at"`
}
