#!/bin/sh
# Install the xdlc command (Xora DLC for Claude Code and Kiro CLI): one file, into ~/.local/bin.
#   curl -fsSL https://raw.githubusercontent.com/TechX-Corp/xoradlc-plugin/main/install.sh | sh
# Then: xdlc setup, xdlc login. Optional: XDLC_BIN_DIR (install dir), XDLC_REF (branch / tag),
# XDLC_PLUGIN_DIR (use xdlc from an existing clone instead of downloading).
set -eu

REPO="TechX-Corp/xoradlc-plugin"
BIN_DIR="${XDLC_BIN_DIR:-$HOME/.local/bin}"
die() { echo "xdlc: $*" >&2; exit 1; }

command -v python3 >/dev/null 2>&1 || die "python3 required (3.9 or later)"
python3 -c 'import sys; sys.exit(sys.version_info < (3, 9))' || die "python3 3.9 or later required"
mkdir -p "$BIN_DIR"
TMP="$BIN_DIR/.xdlc.download"
trap 'rm -f "$TMP"' EXIT

if [ -n "${XDLC_PLUGIN_DIR:-}" ]; then
  cp "$XDLC_PLUGIN_DIR/xdlc" "$TMP"
else
  command -v curl >/dev/null 2>&1 || die "curl required"
  REF="${XDLC_REF:-}"
  if [ -z "$REF" ]; then  # latest release (tag vX.Y.Z), else main
    REF=$(curl -fsSL "https://api.github.com/repos/$REPO/tags?per_page=100" 2>/dev/null | python3 -c '
import json, sys
try:
    tags = [t["name"] for t in json.load(sys.stdin)]
except Exception:
    tags = []
v = []
for t in tags:
    try:
        n = tuple(int(x) for x in t.lstrip("v").split("."))
    except ValueError:
        continue
    if len(n) == 3:
        v.append((n, t))
print(max(v)[1] if v else "main")') || REF=main
  fi
  curl -fsSL "https://raw.githubusercontent.com/$REPO/$REF/xdlc" -o "$TMP" || die "could not download xdlc ($REF)"
fi
python3 -m py_compile "$TMP" 2>/dev/null || die "downloaded xdlc file is invalid"
rm -rf "$BIN_DIR/__pycache__"
chmod 755 "$TMP"
mv "$TMP" "$BIN_DIR/xdlc"

echo "Installed $("$BIN_DIR/xdlc" --version | head -1) to $BIN_DIR/xdlc"
case ":$PATH:" in
  *":$BIN_DIR:"*) echo "Next: xdlc setup (install for Kiro or Claude Code), then xdlc login." ;;
  *) echo "Add $BIN_DIR to PATH, e.g.: echo 'export PATH=\"$BIN_DIR:\$PATH\"' >> ~/.zshrc && exec zsh"
     echo "Then: xdlc setup, then xdlc login." ;;
esac
