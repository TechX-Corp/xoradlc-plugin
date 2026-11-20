//go:build windows

package tui

func readAfterEsc(r keyReader) string { return "esc" }

func drainStdin() {}