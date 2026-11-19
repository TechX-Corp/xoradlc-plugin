package setup

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/TechX-Corp/xoradlc-plugin/internal/config"
	"github.com/TechX-Corp/xoradlc-plugin/internal/fsutil"
)

// XdlcArgv is the command to run xdlc (Go binary path only).
func XdlcArgv() []string {
	p := config.BinPath()
	if _, err := os.Stat(p); err != nil {
		if ex, e := os.Executable(); e == nil {
			p = ex
		}
	}
	return []string{p}
}

// ShellCmd quotes argv for hook command lines.
func ShellCmd(argv []string) string {
	var parts []string
	for _, a := range argv {
		a = strings.ReplaceAll(a, "\\", "/")
		if strings.ContainsAny(a, " /\\") {
			parts = append(parts, `"`+a+`"`)
		} else {
			parts = append(parts, a)
		}
	}
	return strings.Join(parts, " ")
}

// McpEntry is the MCP server block for agents.
func McpEntry() map[string]any {
	argv := XdlcArgv()
	args := []string{"mcp"}
	if len(argv) > 1 {
		args = append(argv[1:], args...)
	}
	return map[string]any{
		"command": argv[0],
		"args":    args,
	}
}

// WriteMcpConfig merges the xora-dlc server into an MCP config file.
func WriteMcpConfig(configPath string, entry map[string]any) error {
	var cfg map[string]any
	_ = fsutil.LoadJSON(configPath, &cfg)
	if cfg == nil {
		cfg = map[string]any{}
	}
	servers, _ := cfg["mcpServers"].(map[string]any)
	if servers == nil {
		servers = map[string]any{}
	}
	block := map[string]any{
		"command":     entry["command"],
		"args":        entry["args"],
		"disabled":    false,
		"autoApprove": []any{},
	}
	servers[config.Server] = block
	cfg["mcpServers"] = servers
	return fsutil.SaveJSON(configPath, cfg, false)
}

// StripMcpConfig removes the xora-dlc block; deletes file if we created an empty shell.
func StripMcpConfig(configPath string, created bool) error {
	var cfg map[string]any
	if err := fsutil.LoadJSON(configPath, &cfg); err != nil || cfg == nil {
		return nil
	}
	servers, _ := cfg["mcpServers"].(map[string]any)
	if servers != nil {
		delete(servers, config.Server)
		cfg["mcpServers"] = servers
	}
	if created {
		if len(servers) == 0 {
			onlyMcp := true
			for k := range cfg {
				if k != "mcpServers" {
					onlyMcp = false
					break
				}
			}
			if onlyMcp {
				return os.Remove(configPath)
			}
		}
	}
	return fsutil.SaveJSON(configPath, cfg, false)
}

func pathExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func relOrAbs(base, target string) string {
	if rel, err := filepath.Rel(base, target); err == nil && !strings.HasPrefix(rel, "..") {
		return filepath.ToSlash(rel)
	}
	return filepath.ToSlash(target)
}