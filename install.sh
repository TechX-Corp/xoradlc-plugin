#!/bin/sh
# Cài plugin xora-dev (vai Developer) cho Kiro CLI hoặc Claude Code, một lệnh:
#   curl -fsSL https://raw.githubusercontent.com/TechX-Corp/xoradlc-plugin/main/install.sh | sh -s -- kiro <workspace>
#   curl -fsSL https://raw.githubusercontent.com/TechX-Corp/xoradlc-plugin/main/install.sh | sh -s -- kiro --global
#   curl -fsSL https://raw.githubusercontent.com/TechX-Corp/xoradlc-plugin/main/install.sh | sh -s -- claude [user|project|local]
# Không cần key: lần đầu dùng, Kiro / Claude mở trình duyệt để đăng nhập Xora DLC (SSO hoặc tài khoản local).
# Tuỳ chọn: XORA_URL (backend khác, vd. http://localhost:8000), XORA_PLUGIN_REF (nhánh / tag, mặc định main),
#           XORA_PLUGIN_DIR (dùng bản clone có sẵn thay vì tải về).
set -eu

TOOL="${1:-}"; TARGET="${2:-}"
REF="${XORA_PLUGIN_REF:-main}"
HOME_DIR="$HOME/.xora/plugin"  # bản plugin cố định: Claude Code nạp marketplace tại chỗ từ đây

die() { echo "xora-dev: $*" >&2; exit 1; }
need() { command -v "$1" >/dev/null 2>&1 || die "cần '$1' trên máy"; }

case "$TOOL" in
  kiro) [ -n "$TARGET" ] || die "kiro cần <workspace> hoặc --global" ;;
  claude) TARGET="${TARGET:-user}"
          case "$TARGET" in user|project|local) ;; *) die "claude: scope là user, project hoặc local" ;; esac ;;
  *) die "Dùng: install.sh kiro <workspace>|--global   hoặc   install.sh claude [user|project|local]" ;;
esac
need python3

# 1. Lấy plugin: bản clone có sẵn, hoặc tải gói của nhánh REF từ GitHub (repo public, không cần token).
if [ -n "${XORA_PLUGIN_DIR:-}" ]; then
  SRC="$(cd "$XORA_PLUGIN_DIR" && pwd)"
else
  need curl; need tar
  TMP="$(mktemp -d)"; trap 'rm -rf "$TMP"' EXIT
  curl -fsSL "https://codeload.github.com/TechX-Corp/xoradlc-plugin/tar.gz/$REF" | tar -xz -C "$TMP"
  rm -rf "$HOME_DIR" && mkdir -p "$(dirname "$HOME_DIR")"
  mv "$TMP"/xoradlc-plugin-* "$HOME_DIR"
  SRC="$HOME_DIR"
fi
[ -f "$SRC/agents-plugin/install-kiro.sh" ] || die "không thấy agents-plugin/ trong $SRC"

# 2. Cài cho công cụ đã chọn.
if [ "$TOOL" = kiro ]; then
  sh "$SRC/agents-plugin/install-kiro.sh" dev "$TARGET"
else
  need claude
  claude plugin marketplace add "$SRC/agents-plugin" >/dev/null 2>&1 || claude plugin marketplace update xora-role-plugins
  claude plugin install xora-dev@xora-role-plugins --scope "$TARGET"
  echo "Đã cài xora-dev cho Claude Code (scope $TARGET). Trong Claude Code gõ /mcp để đăng nhập Xora DLC lần đầu."
fi
