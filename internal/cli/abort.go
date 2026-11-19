package cli

import "fmt"

// Abort stops a command with an exit code (0 = cancelled / soft stop).
type Abort struct {
	Msg  string
	Code int
}

func (e *Abort) Error() string { return e.Msg }

func Fail(msg string) error { return &Abort{Msg: msg, Code: 1} }

func Failf(format string, args ...any) error {
	return &Abort{Msg: fmt.Sprintf(format, args...), Code: 1}
}

func Usage(msg string) error { return &Abort{Msg: msg, Code: 2} }

func Cancel(msg string) error { return &Abort{Msg: msg, Code: 0} }
