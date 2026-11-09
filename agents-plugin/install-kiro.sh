#!/bin/sh
# Cài một plugin theo vai vào Kiro: skills + steering + hook + MCP server.
#   sh install-kiro.sh <vai> <đường-dẫn-workspace>   # ví dụ: sh install-kiro.sh dev ~/code/core-banking
#   sh install-kiro.sh <vai> --global                # mọi workspace (~/.kiro); hook của Kiro chỉ có ở workspace
# <vai>: ba | techlead | dev | qa | reviewer | delivery. Chạy lại để cập nhật.
set -eu

ROOT="$(cd "$(dirname "$0")" && pwd)"
ROLE="${1:-}"; TARGET="${2:-}"
SRC="$ROOT/$ROLE"
if [ -z "$ROLE" ] || [ -z "$TARGET" ] || [ ! -f "$SRC/mcp.json" ]; then
  echo "Dùng: sh install-kiro.sh <ba|techlead|dev|qa|reviewer|delivery> <workspace|--global>" >&2; exit 1
fi
NAME="xora-$ROLE"
case "$TARGET" in
  --global) KIRO="$HOME/.kiro" ;;
  *) WS="$(cd "$TARGET" && pwd)"
     case "$WS/" in "$ROOT"/*) echo "Workspace không được nằm trong agents-plugin/ — chọn repo dự án." >&2; exit 1 ;; esac
     KIRO="$WS/.kiro" ;;
esac

# 1. Bản cài cố định cho MCP server / sensor (mcp.json trỏ tới đây). -L: chép file thật thay cho symlink mcp/xora_mcp.py
DEST="$HOME/.agent-plugins/$NAME"
rm -rf "$DEST" && mkdir -p "$DEST" && cp -RL "$SRC/." "$DEST/"

# 2. Skills, steering, hook
mkdir -p "$KIRO/skills" "$KIRO/steering" "$KIRO/settings"
cp -R "$SRC/skills/." "$KIRO/skills/"
cp "$SRC/AGENTS.md" "$KIRO/steering/$NAME.md"
[ -d "$SRC/steering" ] && cp "$SRC/steering/"*.md "$KIRO/steering/"
if [ -d "$SRC/kiro-hooks" ] && [ "$TARGET" != "--global" ]; then
  mkdir -p "$KIRO/hooks" && cp "$SRC/kiro-hooks/"*.kiro.hook "$KIRO/hooks/"
fi

# 3. Gộp MCP server của vai vào mcp.json, giữ nguyên server khác
python3 - "$SRC/mcp.json" "$KIRO/settings/mcp.json" <<'PY'
import json, sys, pathlib
src, dst = pathlib.Path(sys.argv[1]), pathlib.Path(sys.argv[2])
cfg = json.loads(dst.read_text()) if dst.exists() and dst.read_text().strip() else {}
cfg.setdefault("mcpServers", {}).update(json.loads(src.read_text())["mcpServers"])
dst.write_text(json.dumps(cfg, ensure_ascii=False, indent=2) + "\n")
PY

echo "Đã cài $NAME vào $KIRO"
[ -s "$HOME/.xora/env" ] || echo "Chưa có ~/.xora/env: tạo file chứa XORA_API, XORA_BASE_URL, XORA_AGENT, XORA_API_KEY (xem INSTALL-KIRO.md)."
