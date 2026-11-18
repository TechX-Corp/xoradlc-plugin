#!/bin/sh
# Install the xdlc binary (Xora DLC for Claude Code and Kiro CLI) into ~/.local/bin.
#   curl -fsSL https://raw.githubusercontent.com/TechX-Corp/xoradlc-plugin/main/install.sh | sh
# Optional: XDLC_BIN_DIR, XDLC_REF (tag vX.Y.Z), XDLC_PLUGIN_DIR (local clone with a built ./xdlc).
set -eu

REPO="TechX-Corp/xoradlc-plugin"
BIN_DIR="${XDLC_BIN_DIR:-$HOME/.local/bin}"
die() { echo "xdlc: $*" >&2; exit 1; }

mkdir -p "$BIN_DIR"
TMP="$BIN_DIR/.xdlc.download"
trap 'rm -f "$TMP"' EXIT

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)
case "$arch" in
  x86_64|amd64) arch=amd64 ;;
  aarch64|arm64) arch=arm64 ;;
  *) die "unsupported architecture: $arch" ;;
esac
case "$os" in
  darwin|linux) ;;
  mingw*|msys*|cygwin*) die "use install.ps1 on Windows" ;;
  *) die "unsupported OS: $os" ;;
esac
ASSET="xdlc-${os}-${arch}"

if [ -n "${XDLC_PLUGIN_DIR:-}" ]; then
  src="$XDLC_PLUGIN_DIR/xdlc"
  [ -f "$src" ] || die "XDLC_PLUGIN_DIR has no xdlc binary (build with: go build -o xdlc ./cmd/xdlc)"
  cp "$src" "$TMP"
else
  command -v curl >/dev/null 2>&1 || die "curl required"
  REF="${XDLC_REF:-}"
  if [ -z "$REF" ]; then
    REF=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" 2>/dev/null | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -1) || true
  fi
  [ -n "$REF" ] || die "no GitHub release yet (tag vX.Y.Z). Build from source or set XDLC_PLUGIN_DIR."
  URL="https://github.com/$REPO/releases/download/${REF}/${ASSET}"
  curl -fsSL "$URL" -o "$TMP" || die "could not download $ASSET ($REF)"
fi

chmod 755 "$TMP"
mv "$TMP" "$BIN_DIR/xdlc"

echo "Installed $("$BIN_DIR/xdlc" --version | head -1) to $BIN_DIR/xdlc"
case ":$PATH:" in
  *":$BIN_DIR:"*) echo "Next: xdlc setup (install for Kiro or Claude Code), then xdlc login." ;;
  *) echo "Add $BIN_DIR to PATH, e.g.: echo 'export PATH=\"$BIN_DIR:\$PATH\"' >> ~/.zshrc && exec zsh"
     echo "Then: xdlc setup, then xdlc login." ;;
esac
