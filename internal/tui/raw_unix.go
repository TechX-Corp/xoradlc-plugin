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
		// Lone Esc is 0x1b only. Arrows are ESC [ A/B — peek quickly so Esc cancel is instant.
		defer func() { _ = os.Stdin.SetReadDeadline(time.Time{}) }()
		_ = os.Stdin.SetReadDeadline(time.Now().Add(3 * time.Millisecond))
		rest := make([]byte, 8)
		n, err := r.Read(rest)
		if err != nil || n == 0 {
			return "esc", nil
		}
		if rest[0] != '[' && rest[0] != 'O' {
			return "esc", nil
		}
		if rest[0] == '[' && n >= 2 {
			switch rest[1] {
			case 'A':
				return "up", nil
			case 'B':
				return "down", nil
			}
		}
		if rest[0] == 'O' && n >= 2 {
			switch rest[1] {
			case 'A':
				return "up", nil
			case 'B':
				return "down", nil
			}
		}
		if n == 1 {
			_ = os.Stdin.SetReadDeadline(time.Now().Add(25 * time.Millisecond))
			more := make([]byte, 2)
			n2, err2 := r.Read(more)
			if err2 == nil && n2 > 0 {
				switch rest[0] {
				case '[':
					switch more[0] {
					case 'A':
						return "up", nil
					case 'B':
						return "down", nil
					}
				case 'O':
					switch more[0] {
					case 'A':
						return "up", nil
					case 'B':
						return "down", nil
					}
				}
			}
		}
		return "esc", nil
	default:
		return "", nil
	}
}