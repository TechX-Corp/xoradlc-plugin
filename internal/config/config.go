package config

import (
	"os"
	"path/filepath"
	"runtime"
)

const (
	Version  = "0.3.0"
	Repo     = "TechX-Corp/xoradlc-plugin"
	ProdURL  = "https://xoradlc-platform-api.techxcorp.com"
	ClientID = "xora-plugin"
	Server   = "xora-dlc"
	ExcludeEnd = "# xdlc:end"
)

var (
	Agents = map[string]string{
		"kiro":   "Kiro CLI",
		"claude": "Claude Code",
	}
	Scopes    = []string{"global", "local"}
	ScopeName = map[string]string{"global": "Global", "local": "Local"}
)

func Home() string {
	h, err := os.UserHomeDir()
	if err != nil || h == "" {
		return "."
	}
	return h
}

func XDLC() string {
	return filepath.Join(Home(), ".xdlc")
}

func Bundle() string {
	return filepath.Join(XDLC(), "plugin")
}

func BinPath() string {
	if d := os.Getenv("XDLC_BIN_DIR"); d != "" {
		if runtime.GOOS == "windows" {
			return filepath.Join(d, "xdlc.exe")
		}
		return filepath.Join(d, "xdlc")
	}
	if runtime.GOOS == "windows" {
		base := os.Getenv("LOCALAPPDATA")
		if base == "" {
			base = filepath.Join(Home(), "AppData", "Local")
		}
		return filepath.Join(base, "Programs", "xdlc", "xdlc.exe")
	}
	return filepath.Join(Home(), ".local", "bin", "xdlc")
}

func BaseURL() string {
	if u := os.Getenv("XDLC_API_URL"); u != "" {
		for len(u) > 0 && u[len(u)-1] == '/' {
			u = u[:len(u)-1]
		}
		return u
	}
	return ProdURL
}

func Windows() bool { return runtime.GOOS == "windows" }

// AgentKeys lists setup agent ids.
func AgentKeys() []string {
	out := make([]string, 0, len(Agents))
	for k := range Agents {
		out = append(out, k)
	}
	return out
}

func InstallCmd() string {
	if Windows() {
		return `powershell -ExecutionPolicy ByPass -c "irm https://raw.githubusercontent.com/` + Repo + `/main/install.ps1 | iex"`
	}
	return "curl -fsSL https://raw.githubusercontent.com/" + Repo + "/main/install.sh | sh"
}
