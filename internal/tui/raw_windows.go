//go:build windows

package tui

import (
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/term"
)

func rawKeysAvailable() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && enableVT()
}

func enableVT() bool {
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	getStdHandle := kernel32.NewProc("GetStdHandle")
	getConsoleMode := kernel32.NewProc("GetConsoleMode")
	setConsoleMode := kernel32.NewProc("SetConsoleMode")
	const stdOutputHandle = uintptr(0xFFFFFFF5)
	const enableVTProcessing = 0x0004
	h, _, _ := getStdHandle.Call(stdOutputHandle)
	if h == 0 {
		return false
	}
	var mode uint32
	if r, _, _ := getConsoleMode.Call(h, uintptr(unsafe.Pointer(&mode))); r == 0 {
		return false
	}
	mode |= enableVTProcessing
	if r, _, _ := setConsoleMode.Call(h, uintptr(mode)); r == 0 {
		return false
	}
	return true
}

func terminalWidth() (int, error) {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	return w, err
}

type rawTerm struct{}

func enterRaw() (keyReader, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return nil, syscall.EINVAL
	}
	return &rawTerm{}, nil
}

func (r *rawTerm) Read(p []byte) (int, error) {
	return 0, syscall.ENOSYS
}

func (r *rawTerm) Close() error { return nil }

func readKey(r keyReader) (string, error) {
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	getWch := kernel32.NewProc("_getwch")
	ch, _, _ := getWch.Call()
	if ch == 0 || ch == 0xe0 {
		ch2, _, _ := getWch.Call()
		switch ch2 {
		case 'H':
			return "up", nil
		case 'P':
			return "down", nil
		}
		return "", nil
	}
	if ch == 3 {
		return "", syscall.EINTR
	}
	s := string(rune(ch))
	keys := map[string]string{
		"\r": "enter", "\n": "enter", " ": "space",
		"k": "up", "j": "down", "a": "all", "q": "esc",
	}
	if k, ok := keys[s]; ok {
		return k, nil
	}
	if ch == 4 {
		return "esc", nil
	}
	if ch == 27 {
		return "esc", nil
	}
	return "", nil
}

type keyReader interface {
	Read([]byte) (int, error)
	Close() error
}