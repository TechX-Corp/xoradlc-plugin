package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/TechX-Corp/xoradlc-plugin/internal/api"
	"github.com/TechX-Corp/xoradlc-plugin/internal/cli"
	"github.com/TechX-Corp/xoradlc-plugin/internal/config"
	"github.com/TechX-Corp/xoradlc-plugin/internal/fsutil"
	"github.com/TechX-Corp/xoradlc-plugin/internal/installs"
	"github.com/TechX-Corp/xoradlc-plugin/internal/pack"
	"github.com/TechX-Corp/xoradlc-plugin/internal/setup"
	"github.com/TechX-Corp/xoradlc-plugin/internal/tui"
)

// CmdUpdate replaces the binary from GitHub Releases (no main fallback) and reapplys setups.
func CmdUpdate(argv []string) error {
	if len(argv) > 0 {
		return cli.Usage("update takes no options.")
	}
	ref := os.Getenv("XDLC_REF")
	pluginDir := os.Getenv("XDLC_PLUGIN_DIR")
	if pluginDir != "" {
		if _, err := pack.EnsureBundle(); err != nil {
			return err
		}
		fmt.Printf("%s %s (%s).\n", tui.Ok("Updated"), tui.Em("xdlc"), tui.Em("local pack"))
		return execReapply()
	}
	if ref != "" {
		if err := installPackFromRef(ref); err != nil {
			return err
		}
		fmt.Printf("%s %s and plugin (%s).\n", tui.Ok("Updated"), tui.Em("xdlc"), tui.Em(ref))
		return execReapply()
	}
	tag := api.LatestReleaseNewerThan(config.Version)
	if tag == "" {
		return cli.Fail("No release newer than " + config.Version + " yet. Wait for a GitHub release (vX.Y.Z) with a binary for " + runtime.GOOS + "/" + runtime.GOARCH + ".")
	}
	bin, err := downloadReleaseBinary(tag)
	if err != nil {
		return err
	}
	target := setup.XdlcArgv()[0]
	tmp := target + ".new"
	if err := writeExecutable(bin, tmp); err != nil {
		return err
	}
	if err := pack.MaterializeEmbed(true); err != nil {
		return err
	}
	if err := os.Rename(tmp, target); err != nil {
		return err
	}
	fmt.Printf("%s %s and plugin (%s).\n", tui.Ok("Updated"), tui.Em("xdlc"), tui.Em(tag))
	fmt.Println(tui.Dim("Reinstalling into existing setups:"))
	return execReapply()
}

// CmdUninstall removes all setups and ~/.xdlc.
func CmdUninstall(argv []string) error {
	if len(argv) > 0 {
		return cli.Usage("uninstall takes no options.")
	}
	items := installs.Load()
	fmt.Printf("%s all of %s from this machine:\n\n", tui.Bold("Will remove"), tui.Em("Xora DLC"))
	for _, r := range items {
		fmt.Printf("  - %s: %s\n", tui.Bold("setup"), strings.TrimSpace(installs.Label(r)))
	}
	fmt.Printf("  - %s: login token, work order maps, and plugin bundle (%s)\n", tui.Bold("data"), fsutil.Short(config.XDLC()))
	fmt.Printf("  - %s: %s\n", tui.Bold("command"), fsutil.Short(config.BinPath()))
	fmt.Println()
	ok, err := tui.Confirm("Continue?")
	if err != nil {
		return err
	}
	if !ok {
		return cli.Cancel("Cancelled.")
	}
	fmt.Println()
	for _, r := range items {
		sr := installs.ToSetupRecord(r)
		switch r.Agent {
		case "kiro":
			_ = setup.RemoveKiro(sr, true)
		case "claude":
			_ = setup.RemoveClaude(sr, true)
		}
		fmt.Printf("%s %s\n", tui.Ok("Removed"), strings.TrimSpace(installs.Label(r)))
	}
	removeDataDir(config.XDLC())
	removeInstalledBinary(config.BinPath())
	fmt.Println()
	fmt.Printf("%s %s.\n", tui.Ok("Removed"), tui.Em("Xora DLC"))
	fmt.Println(tui.Dim("Reinstall: " + config.InstallCmd()))
	return nil
}

func removeDataDir(dir string) {
	if dir == "" {
		return
	}
	if runtime.GOOS == "windows" {
		_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err == nil && info != nil && !info.IsDir() {
				_ = os.Chmod(path, 0o666)
			}
			return nil
		})
	}
	_ = os.RemoveAll(dir)
}

// removeInstalledBinary deletes the on-disk xdlc command. On Windows a running
// .exe cannot be unlinked, but it can be renamed aside so the PATH name is gone.
func removeInstalledBinary(bin string) {
	st, err := os.Stat(bin)
	if err != nil || st.IsDir() {
		return
	}
	if err := os.Remove(bin); err == nil {
		return
	}
	if runtime.GOOS != "windows" {
		return
	}
	old := bin + ".old"
	_ = os.Remove(old)
	if err := os.Rename(bin, old); err != nil {
		return
	}
	_ = os.Remove(old) // may still be locked until this process exits; name xdlc.exe is gone
}

// CmdReapply hidden command after update.
func CmdReapply(argv []string) error {
	return installs.Reapply(installs.Load(), false)
}

func execReapply() error {
	exe, err := os.Executable()
	if err != nil {
		return installs.Reapply(installs.Load(), false)
	}
	if runtime.GOOS == "windows" {
		c := exec.Command(exe, "_reapply")
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	}
	return syscall.Exec(exe, []string{exe, "_reapply"}, os.Environ())
}

func writeExecutable(data []byte, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o755); err != nil {
		return err
	}
	return os.Chmod(path, 0o755)
}

func downloadReleaseBinary(tag string) ([]byte, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/%s", config.Repo, url.PathEscape(tag))
	status, out := api.Request("GET", apiURL, api.RequestOpts{Timeout: 30 * time.Second})
	if status != 200 {
		return nil, cli.Fail(fmt.Sprintf("Could not load release %s from GitHub (HTTP %d).", tag, status))
	}
	rel, _ := out.(map[string]any)
	assets, _ := rel["assets"].([]any)
	name := releaseAssetName()
	var downloadURL string
	for _, a := range assets {
		m, _ := a.(map[string]any)
		if m == nil {
			continue
		}
		n := fmt.Sprint(m["name"])
		if n == name || strings.HasPrefix(n, name+".") {
			downloadURL = fmt.Sprint(m["browser_download_url"])
			break
		}
	}
	if downloadURL == "" {
		return nil, cli.Fail(fmt.Sprintf("Release %s has no asset %s. Wait for CI to publish platform binaries.", tag, name))
	}
	resp, err := http.Get(downloadURL)
	if err != nil {
		return nil, cli.Fail(err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, cli.Fail(fmt.Sprintf("Download failed (HTTP %d).", resp.StatusCode))
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if strings.Contains(downloadURL, ".tar.gz") || strings.HasSuffix(name, ".tar.gz") {
		return extractTarGzSingleBin(data)
	}
	return data, nil
}

func releaseAssetName() string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	if goos == "windows" {
		return fmt.Sprintf("xdlc-windows-%s.exe", goarch)
	}
	return fmt.Sprintf("xdlc-%s-%s", goos, goarch)
}

func extractTarGzSingleBin(data []byte) ([]byte, error) {
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if h.Typeflag == tar.TypeReg && (h.Name == "xdlc" || strings.HasSuffix(h.Name, "/xdlc") || strings.HasSuffix(h.Name, "xdlc.exe")) {
			return io.ReadAll(tr)
		}
	}
	return nil, cli.Fail("tar.gz did not contain xdlc binary")
}

func installPackFromRef(ref string) error {
	root, tmp, err := fetchGitHubTarball(ref)
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	return installPackDir(root)
}

func installPackDir(root string) error {
	if st, err := os.Stat(filepath.Join(root, "roles")); err != nil || !st.IsDir() {
		return cli.Fail("roles/ not found in plugin package")
	}
	newDir := filepath.Join(config.XDLC(), "plugin.new")
	oldDir := filepath.Join(config.XDLC(), "plugin.old")
	_ = os.RemoveAll(newDir)
	_ = os.RemoveAll(oldDir)
	if err := copyDir(root, newDir); err != nil {
		return err
	}
	bundle := config.Bundle()
	_ = os.MkdirAll(config.XDLC(), 0o755)
	if _, err := os.Stat(bundle); err == nil {
		_ = os.Rename(bundle, oldDir)
	}
	if err := os.Rename(newDir, bundle); err != nil {
		return err
	}
	_ = os.RemoveAll(oldDir)
	return nil
}

func fetchGitHubTarball(ref string) (root, tmp string, err error) {
	u := fmt.Sprintf("https://codeload.github.com/%s/tar.gz/%s", config.Repo, url.PathEscape(ref))
	resp, err := http.Get(u)
	if err != nil {
		return "", "", cli.Fail(fmt.Sprintf("Could not download the plugin from GitHub (%v).", err))
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", "", cli.Fail(fmt.Sprintf("Could not download the plugin from GitHub (HTTP %d).", resp.StatusCode))
	}
	tmp, err = os.MkdirTemp("", "xdlc-")
	if err != nil {
		return "", "", err
	}
	gr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return "", tmp, err
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", tmp, err
		}
		if !strings.HasPrefix(h.Name, filepath.Base(tmp)) && !strings.Contains(h.Name, "..") {
			// safe join below
		}
		target := filepath.Join(tmp, filepath.FromSlash(h.Name))
		if !strings.HasPrefix(target, tmp) {
			continue
		}
		switch h.Typeflag {
		case tar.TypeDir:
			_ = os.MkdirAll(target, 0o755)
		case tar.TypeReg:
			_ = os.MkdirAll(filepath.Dir(target), 0o755)
			f, err := os.Create(target)
			if err != nil {
				continue
			}
			_, _ = io.Copy(f, tr)
			_ = f.Close()
		}
	}
	entries, _ := os.ReadDir(tmp)
	if len(entries) == 1 && entries[0].IsDir() {
		return filepath.Join(tmp, entries[0].Name()), tmp, nil
	}
	return "", tmp, cli.Fail("Downloaded plugin package has the wrong structure.")
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		if rel == "." {
			return nil
		}
		if strings.Contains(rel, ".git") || strings.Contains(rel, "__pycache__") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return fsutilCopy(path, target)
	})
}

func fsutilCopy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	_ = os.MkdirAll(filepath.Dir(dst), 0o755)
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	_ = out.Close()
	return err
}

// CmdPrint hidden: concatenate files to stdout (Claude hook).
func CmdPrint(argv []string) error {
	for _, f := range argv {
		b, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		_, _ = os.Stdout.Write(b)
		if fi, err := os.Stdout.Stat(); err != nil || (fi.Mode()&os.ModeCharDevice) == 0 {
			_, _ = os.Stdout.Write([]byte{'\n'})
		}
	}
	return nil
}