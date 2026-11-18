//go:build !windows

package tui

type keyReader interface {
	Read([]byte) (int, error)
	Close() error
}