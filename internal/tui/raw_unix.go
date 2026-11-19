//go:build !windows

package tui

import (
	"os"
	"syscall"
	"time"

	"golang.org/x/term"
)

func rawKeysAvailable() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

func enableVT() bool {
	return true
}

func terminalWidth() (int, error) {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	return w, err
}

type rawTerm struct {
	fd  int
	old *term.State
}

func enterRaw() (keyReader, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return nil, syscall.EINVAL
	}
	old, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}
	return &rawTerm{fd: fd, old: old}, nil
}

func (r *rawTerm) Read(p []byte) (int, error) {
	return os.Stdin.Read(p)
}

func (r *rawTerm) Close() error {
	return term.Restore(r.fd, r.old)
}

func readKey(r keyReader) (string, error) {
	buf := make([]byte, 1)
	n, err := r.Read(buf)
	if err != nil || n == 0 {
		return "", err
	}
	b := buf[0]
	switch b {
	case '\r', '\n':
		return "enter", nil
	case ' ':
		return "space", nil
	case 'k':
		return "up", nil
	case 'j':
		return "down", nil
	case 'a':
		return "all", nil
	case 'q', 4:
		return "esc", nil
	case 27:
		// Lone Esc is a single byte; arrow keys are Esc + "[A" etc. Do not block waiting for a sequence.
		_ = os.Stdin.SetReadDeadline(time.Now().Add(30 * time.Millisecond))
		defer func() { _ = os.Stdin.SetReadDeadline(time.Time{}) }()
		seq := make([]byte, 3)
		n, err := r.Read(seq)
		if err != nil || n == 0 {
			return "esc", nil
		}
		if seq[0] == '[' {
			switch {
			case n >= 2 && seq[1] == 'A':
				return "up", nil
			case n >= 2 && seq[1] == 'B':
				return "down", nil
			}
		} else if seq[0] == 'O' {
			switch {
			case n >= 2 && seq[1] == 'A':
				return "up", nil
			case n >= 2 && seq[1] == 'B':
				return "down", nil
			}
		}
		return "esc", nil
	default:
		return "", nil
	}
}