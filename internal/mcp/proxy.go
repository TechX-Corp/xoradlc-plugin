package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/TechX-Corp/xoradlc-plugin/internal/api"
	"github.com/TechX-Corp/xoradlc-plugin/internal/auth"
	"github.com/TechX-Corp/xoradlc-plugin/internal/config"
)

var askTool = map[string]any{
	"name": "xora_ask",
	"description": "Ask the Xora DLC agent, in plain words, anything about the work in Xora DLC or the company " +
		"knowledge base: my work, a work order's status and steps, its documents (BRD, design), open " +
		"questions, decisions, conventions and precedents. The agent picks the Xora DLC functions itself " +
		"and answers with citations; mention the work order code (WO-...) when you know it. Pass " +
		"session_id from a previous answer to continue the conversation.",
	"inputSchema": map[string]any{
		"type": "object",
		"properties": map[string]any{
			"text": map[string]any{"type": "string", "minLength": 1, "maxLength": 3800},
			"session_id": map[string]any{
				"type":        "string",
				"description": "From a previous answer, to continue",
			},
		},
		"required": []any{"text"},
	},
}

// Run is the stdio JSON-RPC MCP proxy (cmd_mcp parity).
func Run() error {
	outLock := sync.Mutex{}
	sc := bufio.NewScanner(os.Stdin)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	var wg sync.WaitGroup
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			continue
		}
		wg.Add(1)
		go func(line string) {
			defer wg.Done()
			var msg map[string]any
			if err := json.Unmarshal([]byte(line), &msg); err != nil {
				return
			}
			resp := forward(msg)
			if resp == nil {
				return
			}
			b, _ := json.Marshal(resp)
			outLock.Lock()
			_, _ = os.Stdout.Write(append(b, '\n'))
			outLock.Unlock()
		}(line)
	}
	wg.Wait()
	return sc.Err()
}

func forward(msg map[string]any) map[string]any {
	url := config.BaseURL() + "/mcp/user"
	token, _ := auth.AccessToken(false)
	if token != "" {
		status, out := api.Request("POST", url, api.RequestOpts{Body: msg, Token: token, Timeout: 600 * time.Second})
		if status == 401 {
			token, _ = auth.AccessToken(true)
			if token != "" {
				status, out = api.Request("POST", url, api.RequestOpts{Body: msg, Token: token, Timeout: 0})
			}
		}
		if status == 202 {
			return nil
		}
		if _, hasID := msg["id"]; !hasID && status == 200 {
			return nil
		}
		if status == 200 {
			if m, ok := out.(map[string]any); ok {
				return m
			}
		}
		var reason string
		if status == 0 {
			reason = fmt.Sprintf("Could not connect to Xora DLC (%s): %v. Check the network and ask again.", config.BaseURL(), out)
		} else if status == 401 {
			reason = "Xora DLC session expired. Run `xdlc login` in a terminal and ask again."
		} else {
			reason = fmt.Sprintf("Xora DLC returned HTTP %d: %v", status, out)
		}
		if _, hasID := msg["id"]; !hasID {
			return nil
		}
		return localAnswer(msg, reason)
	}
	reason := fmt.Sprintf("Not logged in to Xora DLC (%s). Run `xdlc login` in a terminal and ask again.", config.BaseURL())
	if _, hasID := msg["id"]; !hasID {
		return nil
	}
	return localAnswer(msg, reason)
}

func localAnswer(msg map[string]any, reason string) map[string]any {
	m, _ := msg["method"].(string)
	mid := msg["id"]
	var result any
	switch m {
	case "initialize":
		asked := ""
		if params, ok := msg["params"].(map[string]any); ok {
			asked, _ = params["protocolVersion"].(string)
		}
		if asked == "" {
			asked = "2025-06-18"
		}
		result = map[string]any{
			"protocolVersion": asked,
			"capabilities":    map[string]any{"tools": map[string]any{"listChanged": false}},
			"serverInfo":      map[string]any{"name": "xoradlc", "version": config.Version},
		}
	case "ping":
		result = map[string]any{}
	case "tools/list":
		result = map[string]any{"tools": []any{askTool}}
	case "tools/call":
		result = map[string]any{
			"content": []any{map[string]any{"type": "text", "text": reason}},
			"isError": true,
		}
	default:
		return map[string]any{
			"jsonrpc": "2.0",
			"id":      mid,
			"error":   map[string]any{"code": -32601, "message": fmt.Sprintf("Method not found: %s", m)},
		}
	}
	return map[string]any{"jsonrpc": "2.0", "id": mid, "result": result}
}