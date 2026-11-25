//go:build !windows

package usage

import (
	"os/exec"
	"syscall"
)

// detach starts cmd in its own session so it outlives the hook's process group.
func detach(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
