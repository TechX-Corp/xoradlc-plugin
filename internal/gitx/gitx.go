package gitx

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/TechX-Corp/xoradlc-plugin/internal/config"
	"github.com/TechX-Corp/xoradlc-plugin/internal/tui"
)

// GitRoot returns repository root for path, or "".
func GitRoot(path string) string {
	out, ok := runGit(path, "rev-parse", "--show-toplevel")
	if !ok {
		return ""
	}
	abs, _ := filepath.Abs(strings.TrimSpace(out))
	return abs
}

// GitTracked returns which relative paths are tracked under base.
func GitTracked(base string, rels []string) []string {
	if len(rels) == 0 || GitRoot(base) == "" {
		return nil
	}
	args := append([]string{"ls-files", "--"}, rels...)
	out, ok := runGit(base, args...)
	if !ok {
		return nil
	}
	set := map[string]struct{}{}
	for _, r := range rels {
		set[r] = struct{}{}
	}
	var tracked []string
	for _, line := range strings.Fields(out) {
		if _, ok := set[line]; ok {
			tracked = append(tracked, line)
		}
	}
	sort.Strings(tracked)
	return tracked
}

// ExcludeFile path to .git/info/exclude.
func ExcludeFile(base string) string {
	out, ok := runGit(base, "rev-parse", "--git-path", "info/exclude")
	if !ok {
		return filepath.Join(base, ".git", "info", "exclude")
	}
	p := strings.TrimSpace(out)
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(base, p)
}

// SetExclude updates the xdlc block in info/exclude; empty rels removes the block.
func SetExclude(base string, rels []string) error {
	root := GitRoot(base)
	if root == "" {
		return nil
	}
	baseAbs, _ := filepath.Abs(base)
	rel, err := filepath.Rel(root, baseAbs)
	if err != nil {
		rel = "."
	}
	prefix := filepath.ToSlash(rel)
	begin := "# xdlc:begin / (Xora DLC, Local)"
	if prefix != "." {
		begin = "# xdlc:begin " + prefix + " (Xora DLC, Local)"
	}
	f := ExcludeFile(base)
	var lines []string
	if b, err := os.ReadFile(f); err == nil {
		lines = strings.Split(strings.TrimRight(string(b), "\n"), "\n")
		if len(lines) == 1 && lines[0] == "" {
			lines = nil
		}
	}
	if idx := indexOf(lines, begin); idx >= 0 {
		end := idx
		for i := idx; i < len(lines); i++ {
			if lines[i] == config.ExcludeEnd {
				end = i
				break
			}
		}
		lines = append(lines[:idx], lines[end+1:]...)
	}
	if len(rels) > 0 {
		pre := ""
		if prefix != "." {
			pre = prefix + "/"
		}
		sorted := append([]string(nil), rels...)
		sort.Strings(sorted)
		lines = append(lines, begin)
		for _, r := range sorted {
			lines = append(lines, "/"+pre+r)
		}
		lines = append(lines, config.ExcludeEnd)
	}
	text := strings.Join(lines, "\n")
	if len(lines) > 0 {
		text += "\n"
	}
	if err := os.MkdirAll(filepath.Dir(f), 0o755); err != nil {
		return err
	}
	return os.WriteFile(f, []byte(text), 0o644)
}

// LocalGit handles tracked files for local scope installs.
func LocalGit(base string, rels []string, old map[string]any, interactive bool) (string, error) {
	tracked := GitTracked(base, rels)
	if len(tracked) == 0 {
		return "", nil
	}
	if old != nil {
		if c, _ := old["git_tracked"].(string); c == "untrack" || c == "keep" {
			if c == "untrack" {
				args := append([]string{"rm", "--cached", "-q", "--"}, tracked...)
				_, _ = runGitCheck(base, args...)
				fmt.Println("Untracked from git (not committed). Commit this change if you want them removed from the team repo.")
			}
			return c, nil
		}
	}
	if !interactive {
		return "keep", nil
	}
	fmt.Println("These files are already tracked by git in this repo:")
	for _, r := range tracked {
		fmt.Printf("  %s\n", r)
	}
	choice, err := tui.Choose("Local means git does not track xdlc files. Choose:", []tui.Option{
		{Key: "untrack", Label: "Untrack (git rm --cached): files stay on disk; when you commit, they leave the team repo too"},
		{Key: "keep", Label: "Keep as-is: git still tracks them; xdlc changes will show in git status"},
	})
	if err != nil {
		return "", err
	}
	if choice == "untrack" {
		args := append([]string{"rm", "--cached", "-q", "--"}, tracked...)
		_, _ = runGitCheck(base, args...)
		fmt.Println("Untracked from git (not committed). Commit this change if you want them removed from the team repo.")
	}
	return choice, nil
}

// GitOut runs git in root and returns trimmed stdout or "".
func GitOut(root string, args ...string) string {
	out, ok := runGit(root, args...)
	if !ok {
		return ""
	}
	return strings.TrimSpace(out)
}

var (
	scpHost   = regexp.MustCompile(`^[^/\\]+:[^/\\]`)
	winDrive  = regexp.MustCompile(`^[A-Za-z]:[/\\]`)
)

// NormalizeRemote canonicalizes git remote URLs.
func NormalizeRemote(raw string) string {
	u := strings.TrimSpace(raw)
	if u == "" {
		return ""
	}
	if strings.Contains(u, "://") {
		parsed, err := url.Parse(u)
		if err != nil {
			return ""
		}
		if parsed.Scheme == "file" {
			return ""
		}
		host := strings.ToLower(parsed.Hostname())
		path := strings.Trim(parsed.Path, "/")
		if strings.HasSuffix(path, ".git") {
			path = path[:len(path)-4]
		}
		if host == "" || path == "" {
			return ""
		}
		return host + "/" + strings.ToLower(path)
	}
	if scpHost.MatchString(u) && !winDrive.MatchString(u) {
		host, path, _ := strings.Cut(u, ":")
		host = strings.ToLower(strings.TrimPrefix(host, "user@"))
		path = strings.Trim(path, "/")
		if strings.HasSuffix(path, ".git") {
			path = path[:len(path)-4]
		}
		if host == "" || path == "" {
			return ""
		}
		return host + "/" + strings.ToLower(path)
	}
	return ""
}

// RepoIdentity returns (repo key, origin URL).
func RepoIdentity(root string) (string, string) {
	origin := GitOut(root, "remote", "get-url", "origin")
	if id := NormalizeRemote(origin); id != "" {
		return id, origin
	}
	abs, _ := filepath.Abs(root)
	return "path:" + filepath.ToSlash(abs), origin
}

// ChangedFiles lists porcelain paths relative to root.
func ChangedFiles(root string) []string {
	out := GitOut(root, "status", "--porcelain")
	if out == "" {
		return nil
	}
	var files []string
	seen := map[string]struct{}{}
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		ln := sc.Text()
		if len(ln) < 4 {
			continue
		}
		path := ln[3:]
		if i := strings.Index(path, " -> "); i >= 0 {
			path = path[i+4:]
		}
		if path == "" {
			continue
		}
		if _, ok := seen[path]; !ok {
			seen[path] = struct{}{}
			files = append(files, path)
		}
	}
	return files
}

func runGit(cwd string, args ...string) (string, bool) {
	git, err := exec.LookPath("git")
	if err != nil {
		return "", false
	}
	cmd := exec.Command(git, args...)
	cmd.Dir = cwd
	b, err := cmd.Output()
	if err != nil {
		return "", false
	}
	return string(b), true
}

func runGitCheck(cwd string, args ...string) (string, bool) {
	git, err := exec.LookPath("git")
	if err != nil {
		return "", false
	}
	cmd := exec.Command(git, args...)
	cmd.Dir = cwd
	b, err := cmd.CombinedOutput()
	return string(b), err == nil
}

func indexOf(lines []string, s string) int {
	for i, l := range lines {
		if l == s {
			return i
		}
	}
	return -1
}