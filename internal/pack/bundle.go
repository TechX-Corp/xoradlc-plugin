package pack

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/TechX-Corp/xoradlc-plugin/internal/config"
	packembed "github.com/TechX-Corp/xoradlc-plugin"
)

// EnsureBundle returns the plugin pack root (roles/, host-agents/, standards/).
func EnsureBundle() (string, error) {
	if d := os.Getenv("XDLC_PLUGIN_DIR"); d != "" {
		root := filepath.Clean(d)
		if st, err := os.Stat(filepath.Join(root, "roles")); err != nil || !st.IsDir() {
			return "", fmt.Errorf("roles/ not found in XDLC_PLUGIN_DIR %s", root)
		}
		return root, nil
	}
	root := config.Bundle()
	if st, err := os.Stat(filepath.Join(root, "roles")); err == nil && st.IsDir() {
		return root, nil
	}
	if err := MaterializeEmbed(true); err != nil {
		return "", err
	}
	return root, nil
}

// MaterializeEmbed copies the embedded pack to ~/.xdlc/plugin (atomic rename).
func MaterializeEmbed(force bool) error {
	dst := config.Bundle()
	if !force {
		if st, err := os.Stat(filepath.Join(dst, "roles")); err == nil && st.IsDir() {
			return nil
		}
	}
	if err := os.MkdirAll(config.XDLC(), 0o755); err != nil {
		return err
	}
	newDir := filepath.Join(config.XDLC(), "plugin.new")
	oldDir := filepath.Join(config.XDLC(), "plugin.old")
	_ = os.RemoveAll(newDir)
	_ = os.RemoveAll(oldDir)
	if err := os.MkdirAll(newDir, 0o755); err != nil {
		return err
	}
	if err := copyEmbed(packembed.Pack, ".", newDir); err != nil {
		_ = os.RemoveAll(newDir)
		return err
	}
	if _, err := os.Stat(filepath.Join(newDir, "roles")); err != nil {
		_ = os.RemoveAll(newDir)
		return fmt.Errorf("roles/ missing after materialize")
	}
	if _, err := os.Stat(dst); err == nil {
		_ = os.Rename(dst, oldDir)
	}
	if err := os.Rename(newDir, dst); err != nil {
		return err
	}
	_ = os.RemoveAll(oldDir)
	return nil
}

func copyEmbed(efs fs.FS, dir, dstRoot string) error {
	return fs.WalkDir(efs, dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}
		rel := path
		if dir == "." {
			rel = path
		}
		target := filepath.Join(dstRoot, filepath.FromSlash(rel))
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		// Never materialize Go sources under host-agents (compiled into xdlc).
		if strings.HasSuffix(path, ".go") {
			return nil
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		src, err := efs.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()
		out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		_, cpErr := io.Copy(out, src)
		closeErr := out.Close()
		if cpErr != nil {
			return cpErr
		}
		return closeErr
	})
}