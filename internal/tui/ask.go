package tui

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/TechX-Corp/xoradlc-plugin/internal/cli"
)

var stdinReader = bufio.NewReader(os.Stdin)

func askLine(prompt string) (string, error) {
	_, _ = os.Stdout.WriteString(prompt)
	line, err := stdinReader.ReadString('\n')
	line = strings.TrimRight(line, "\r\n")
	if err != nil && err != io.EOF {
		return "", err
	}
	if line == "" && err == io.EOF {
		return "", cli.Usage("Need a terminal answer: run the command in an interactive terminal.")
	}
	return strings.TrimSpace(line), nil
}