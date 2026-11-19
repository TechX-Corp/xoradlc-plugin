package tui

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/TechX-Corp/xoradlc-plugin/internal/cli"
)

// Option is a labeled choice for Choose.
type Option struct {
	Key   string
	Label string
}

// Menu shows an interactive list; multi selects many indices with Space / a.
func Menu(title string, labels []string, multi bool) ([]int, error) {
	if len(labels) == 0 {
		return nil, cli.Fail("empty menu")
	}
	if !ArrowKeys() {
		return menuPlain(title, labels, multi)
	}
	n := len(labels)
	// Title + n options + tip row — kept in one redraw/clear block so back-to-back menus do not overlap.
	menuLines := n + 2
	tips := menuTips(multi)
	cur := 0
	marked := map[int]struct{}{}
	var hint *string
	out := os.Stdout

	width := 10000
	if sz, err := terminalWidth(); err == nil && sz >= 40 {
		width = sz - 1
	}

	draw := func(first bool) {
		if !first {
			_, _ = fmt.Fprintf(out, "\x1b[%dA", menuLines)
		}
		_, _ = fmt.Fprint(out, "\r\x1b[2K"+title+"\n")
		for i, text := range labels {
			var prefix string
			if multi {
				box := "□ "
				if _, ok := marked[i]; ok {
					box = "▣ "
				}
				arrow := " "
				if i == cur {
					arrow = ">"
				}
				prefix = arrow + " " + box
			} else if i == cur {
				prefix = "◉ "
			} else {
				prefix = "○ "
			}
			room := width - len(prefix)
			if room < 0 {
				room = 0
			}
			label := text
			if len(label) > room {
				label = label[:room]
			}
			line := prefix + label
			if i == cur {
				line = Focus(line)
			}
			_, _ = fmt.Fprint(out, "\r\x1b[2K"+line+"\n")
		}
		tipLine := tips
		if hint != nil {
			tipLine = *hint
		}
		_, _ = fmt.Fprint(out, "\r\x1b[2K"+tipLine+"\n")
		_ = out.Sync()
	}

	_, _ = fmt.Fprint(out, "\x1b[?25l")
	raw, err := enterRaw()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = raw.Close()
		_, _ = fmt.Fprint(out, fmt.Sprintf("\x1b[%dA\x1b[J\x1b[?25h", menuLines))
		_ = out.Sync()
	}()

	var result []int
	draw(true)
	for result == nil {
		k, err := readKey(raw)
		if err != nil {
			return nil, err
		}
		hint = nil
		switch k {
		case "up":
			cur = (cur - 1 + n) % n
		case "down":
			cur = (cur + 1) % n
		case "space":
			if multi {
				if _, ok := marked[cur]; ok {
					delete(marked, cur)
				} else {
					marked[cur] = struct{}{}
				}
			}
		case "all":
			if multi {
				if len(marked) == n {
					marked = map[int]struct{}{}
				} else {
					marked = map[int]struct{}{}
					for i := 0; i < n; i++ {
						marked[i] = struct{}{}
					}
				}
			}
		case "enter":
			if !multi {
				result = []int{cur}
			} else if len(marked) > 0 {
				result = make([]int, 0, len(marked))
				for i := range marked {
					result = append(result, i)
				}
				sort.Ints(result)
			} else {
				h := Dim("Nothing selected: ") + Em("Space") + Dim(" to select, ") + Em("Esc") + Dim(" to cancel")
				hint = &h
			}
		case "esc":
			return nil, cli.Cancel("Cancelled.")
		}
		if result == nil {
			draw(false)
		}
	}
	for _, i := range result {
		fmt.Printf("  › %s\n", Bold(labels[i]))
	}
	// Callers often Println next; ensure cursor is on a fresh line after raw-mode redraw.
	_, _ = fmt.Fprint(out, "\r\x1b[2K\n")
	_ = out.Sync()
	return result, nil
}

func menuPlain(title string, labels []string, multi bool) ([]int, error) {
	fmt.Println(title)
	n := len(labels)
	for i, text := range labels {
		fmt.Printf("  %d) %s\n", i+1, text)
	}
	if multi {
		fmt.Println(Dim("Select items: ") + Em("1,3") + Dim(" | ") + Em("1-2") + Dim(" | ") + Em("all"))
	}
	answer, err := Ask("> ")
	if err != nil {
		return nil, err
	}
	if multi {
		picked := PickMany(n, answer)
		if picked == nil {
			return nil, cli.Fail("Invalid selection.")
		}
		for _, i := range picked {
			fmt.Printf("  › %s\n", Bold(labels[i]))
		}
		return picked, nil
	}
	num, err := strconv.Atoi(strings.TrimSpace(answer))
	if err != nil || num < 1 || num > n {
		return nil, cli.Fail(fmt.Sprintf("Enter a number from 1 to %d.", n))
	}
	idx := num - 1
	fmt.Printf("  › %s\n", Bold(labels[idx]))
	return []int{idx}, nil
}

// Choose picks one option by key (arrow menu or numbered prompt).
func Choose(title string, options []Option) (string, error) {
	if ArrowKeys() {
		labels := make([]string, len(options))
		for i, o := range options {
			labels[i] = o.Label
		}
		idx, err := Menu(title, labels, false)
		if err != nil {
			return "", err
		}
		return options[idx[0]].Key, nil
	}
	fmt.Println(title)
	for i, o := range options {
		fmt.Printf("  %d) %s\n", i+1, o.Label)
	}
	for {
		a, err := Ask("> ")
		if err != nil {
			return "", err
		}
		if n, err := strconv.Atoi(a); err == nil && n >= 1 && n <= len(options) {
			return options[n-1].Key, nil
		}
		fmt.Printf("Enter a number from 1 to %d.\n", len(options))
	}
}

// Confirm is a Yes/No menu; default No.
func Confirm(prompt string) (bool, error) {
	key, err := Choose(prompt, []Option{{Key: "no", Label: "No"}, {Key: "yes", Label: "Yes"}})
	if err != nil {
		return false, err
	}
	return key == "yes", nil
}

// PickMany parses "1,3" | "1-2" | "all" into 0-based indices, or nil if invalid.
func PickMany(n int, answer string) []int {
	answer = strings.TrimSpace(answer)
	if strings.EqualFold(answer, "all") {
		out := make([]int, n)
		for i := range out {
			out[i] = i
		}
		return out
	}
	answer = strings.ReplaceAll(answer, " ", "")
	var out []int
	for _, part := range strings.Split(answer, ",") {
		if part == "" {
			continue
		}
		lo, hi, ok := strings.Cut(part, "-")
		if !ok {
			hi = lo
		}
		if !isDigits(lo) || (part != lo && !isDigits(hi)) {
			return nil
		}
		a, _ := strconv.Atoi(lo)
		b, _ := strconv.Atoi(hi)
		if a < 1 || b < a || b > n {
			return nil
		}
		for i := a; i <= b; i++ {
			idx := i - 1
			found := false
			for _, x := range out {
				if x == idx {
					found = true
					break
				}
			}
			if !found {
				out = append(out, idx)
			}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}