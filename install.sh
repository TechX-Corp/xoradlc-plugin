#!/bin/sh
# Cài lệnh xdlc (Xora DLC cho Claude Code và Kiro CLI): một file, vào ~/.local/bin.
#   curl -fsSL https://raw.githubusercontent.com/TechX-Corp/xoradlc-plugin/main/install.sh | sh
# Sau đó: xdlc setup, xdlc login. Tuỳ chọn: XDLC_BIN_DIR (thư mục cài), XDLC_REF (nhánh / tag),
# XDLC_PLUGIN_DIR (lấy xdlc từ bản clone có sẵn thay vì tải về).
set -eu

REPO="TechX-Corp/xoradlc-plugin"
BIN_DIR="${XDLC_BIN_DIR:-$HOME/.local/bin}"
die() { echo "xdlc: $*" >&2; exit 1; }

command -v python3 >/dev/null 2>&1 || die "cần python3 (3.9 trở lên)"
python3 -c 'import sys; sys.exit(sys.version_info < (3, 9))' || die "cần python3 3.9 trở lên"
mkdir -p "$BIN_DIR"
TMP="$BIN_DIR/.xdlc.download"
trap 'rm -f "$TMP"' EXIT

if [ -n "${XDLC_PLUGIN_DIR:-}" ]; then
  cp "$XDLC_PLUGIN_DIR/xdlc" "$TMP"
else
  command -v curl >/dev/null 2>&1 || die "cần curl"
  REF="${XDLC_REF:-}"
  if [ -z "$REF" ]; then  # bản phát hành mới nhất (tag vX.Y.Z), chưa có thì main
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
  curl -fsSL "https://raw.githubusercontent.com/$REPO/$REF/xdlc" -o "$TMP" || die "không tải được xdlc ($REF)"
fi
python3 -m py_compile "$TMP" 2>/dev/null || die "file xdlc tải về không hợp lệ"
rm -rf "$BIN_DIR/__pycache__"
chmod 755 "$TMP"
mv "$TMP" "$BIN_DIR/xdlc"

echo "Đã cài $("$BIN_DIR/xdlc" --version | head -1) vào $BIN_DIR/xdlc"
case ":$PATH:" in
  *":$BIN_DIR:"*) echo "Tiếp theo: xdlc setup (cài cho Kiro hoặc Claude Code), rồi xdlc login." ;;
  *) echo "Thêm $BIN_DIR vào PATH, ví dụ: echo 'export PATH=\"$BIN_DIR:\$PATH\"' >> ~/.zshrc && exec zsh"
     echo "Sau đó: xdlc setup, rồi xdlc login." ;;
esac
