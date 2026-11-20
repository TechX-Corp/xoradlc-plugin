//go:build !windows

package tui

import (
	"io"
	"os"
	"runtime"
	"syscall"
	"unsafe"
)

func fionreadReq() uintptr {
	switch runtime.GOOS {
	case "darwin":
		return 0x4004667f
	case "linux":
		return 0x541B
	default:
		return 0
	}
}

func stdinWaitingBytes(fd int) int {
	req := fionreadReq()
	if req == 0 {
		return -1
	}
	var n int32
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), req, uintptr(unsafe.Pointer(&n)))
	if errno != 0 {
		return -1
	}
	return int(n)
}

func readAfterEsc(r keyReader) string {
	fd := int(os.Stdin.Fd())
	waiting := stdinWaitingBytes(fd)
	if waiting == 0 {
		return "esc"
	}
	if waiting > 0 {
		buf := make([]byte, waiting)
		if _, err := io.ReadFull(r, buf); err != nil {
			return "esc"
		}
		return parseEscSuffix(buf)
	}
	rest := stdinQueued(r)
	if len(rest) == 0 {
		return "esc"
	}
	return parseEscSuffix(rest)
}

func drainStdin() {
	fd := int(os.Stdin.Fd())
	for {
		n := stdinWaitingBytes(fd)
		if n <= 0 {
			return
		}
		buf := make([]byte, n)
		_, _ = io.ReadFull(os.Stdin, buf)
	}
}