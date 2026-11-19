package tui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"

)

var (
	colorOnce  sync.Once
	colorCache bool
)

// ColorEnabled reports whether ANSI styling is allowed (TTY, no NO_COLOR / XDLC_PLAIN / TERM=dumb).
func ColorEnabled() bool {
	colorOnce.Do(probeColor)
	return colorCache
}

func probeColor() {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("XDLC_PLAIN") != "" || os.Getenv("TERM") == "dumb" {
		colorCache = false
		return
	}
	fi, err := os.Stdout.Stat()
	if err != nil || (fi.Mode()&os.ModeCharDevice) == 0 {
		colorCache = false
		return
	}
	colorCache = enableVT()
}

func paint(text string, codes ...int) string {
	if text == "" || !ColorEnabled() {
		return text
	}
	parts := make([]string, len(codes))
	for i, c := range codes {
		parts[i] = fmt.Sprint(c)
	}
	return "\x1b[" + strings.Join(parts, ";") + "m" + text + "\x1b[0m"
}

func Bold(text string) string  { return paint(text, 1) }
func Dim(text string) string   { return paint(text, 2) }
func Focus(text string) string { return paint(text, 1, 36) }
func Em(text string) string    { return paint(text, 36) }
func Ok(text string) string    { return paint(text, 32) }
func Warn(text string) string  { return paint(text, 33) }

// RoleLabel formats a pack or account role id for terminal output (uppercase + accent).
func RoleLabel(id string) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return Dim("no role")
	}
	return Em(strings.ToUpper(id))
}

// ArrowKeys is true when interactive arrow-key menus are available.
func ArrowKeys() bool {
	if os.Getenv("XDLC_PLAIN") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	inFi, err := os.Stdin.Stat()
	if err != nil || (inFi.Mode()&os.ModeCharDevice) == 0 {
		return false
	}
	outFi, err := os.Stdout.Stat()
	if err != nil || (outFi.Mode()&os.ModeCharDevice) == 0 {
		return false
	}
	return rawKeysAvailable()
}

// PrintHelp paints help blocks like the Python TUI.
func PrintHelp(text string, w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	out := bufio.NewWriter(w)
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	cmdLine := regexp.MustCompile(`^(  )(\S.*?\S|\S)( {2,})(.+)$`)
	for i, line := range lines {
		if line == "" {
			_, _ = fmt.Fprintln(out)
			continue
		}
		if i == 0 && strings.HasPrefix(line, "xdlc ") {
			body := line[5:]
			for _, name := range []string{"Xora DLC", "Claude Code", "Kiro CLI", "xora-dlc"} {
				body = strings.ReplaceAll(body, name, Em(name))
			}
			_, _ = fmt.Fprintln(out, Bold("xdlc")+" "+body)
			continue
		}
		switch line {
		case "Commands:", "Options:", "Shortcuts (skip terminal prompts):":
			_, _ = fmt.Fprintln(out, Bold(line))
			continue
		}
		if strings.HasPrefix(line, "Usage:") {
			_, _ = fmt.Fprintln(out, Bold("Usage:")+line[6:])
			continue
		}
		if strings.HasPrefix(line, "Scope:") {
			_, _ = fmt.Fprintln(out, Bold("Scope:")+Dim(line[6:]))
			continue
		}
		if strings.HasPrefix(line, "For details") {
			_, _ = fmt.Fprintln(out, Dim(line))
			continue
		}
		if m := cmdLine.FindStringSubmatch(line); m != nil {
			_, _ = fmt.Fprintln(out, m[1]+Em(m[2])+m[3]+m[4])
			continue
		}
		_, _ = fmt.Fprintln(out, line)
	}
	_ = out.Flush()
}

// PrintTable prints a bordered table (ASCII when plain).
func PrintTable(headers []string, rows [][]string, indent string) {
	if indent == "" {
		indent = "  "
	}
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	data := make([][]string, len(rows))
	for ri, row := range rows {
		line := make([]string, len(headers))
		for i := range headers {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			line[i] = cell
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
		data[ri] = line
	}
	plain := os.Getenv("XDLC_PLAIN") != "" || os.Getenv("TERM") == "dumb"
	var h, v, tl, tr, bl, br, l, r, t, b, x string
	if plain {
		h, v, x = "-", "|", "+"
		tl, tr, bl, br, l, r, t, b = x, x, x, x, x, x, x, x
	} else {
		h, v = "─", "│"
		tl, tr, bl, br = "┌", "┐", "└", "┘"
		l, r, t, b, x = "├", "┤", "┬", "┴", "┼"
	}
	border := func(left, mid, right, fill string) string {
		parts := make([]string, len(widths))
		for i, w := range widths {
			parts[i] = strings.Repeat(fill, w+2)
		}
		return indent + left + strings.Join(parts, mid) + right
	}
	rowLine := func(cells []string) string {
		parts := make([]string, len(cells))
		for i, c := range cells {
			parts[i] = fmt.Sprintf(" %-*s ", widths[i], c)
		}
		return indent + v + strings.Join(parts, v) + v
	}
	fmt.Println(border(tl, t, tr, h))
	fmt.Println(rowLine(headers))
	fmt.Println(border(l, x, r, h))
	for _, row := range data {
		fmt.Println(rowLine(row))
	}
	fmt.Println(border(bl, b, br, h))
}

func Ask(prompt string) (string, error) {
	return askLine(prompt)
}

func tipAction(key, gloss string) string {
	return Em(key) + Dim(" "+gloss)
}

func menuTips(multi bool) string {
	sep := Dim(" · ")
	if multi {
		return strings.Join([]string{
			tipAction("↑/↓", "move"),
			tipAction("Space", "select"),
			tipAction("a", "select all"),
			tipAction("Enter", "confirm"),
			tipAction("Esc", "cancel"),
		}, sep)
	}
	return strings.Join([]string{
		tipAction("↑/↓", "move"),
		tipAction("Enter", "select"),
		tipAction("Esc", "cancel"),
	}, sep)
}