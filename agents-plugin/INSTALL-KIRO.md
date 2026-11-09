# Cài plugin theo vai vào Kiro

Cần: Kiro, `python3` ≥ 3.9, macOS / Linux (Windows dùng WSL), API key Xora.

## 1. Lưu kết nối AAP (một lần mỗi máy)

```bash
mkdir -p ~/.xora
cat > ~/.xora/env <<'ENV'
export XORA_API=runs
export XORA_BASE_URL=https://d1xbzyews27i3w.cloudfront.net/v1
export XORA_AGENT=agent_mw3at9y00
export XORA_API_KEY=xap_...
ENV
chmod 600 ~/.xora/env
```

Không commit file này, không dán key vào `mcp.json`.

## 2. Cài vai

Chạy từ thư mục clone `xoradlc-plugin`, đích là repo dự án mở trong Kiro:

```bash
sh agents-plugin/install-kiro.sh ba  ~/code/core-banking
sh agents-plugin/install-kiro.sh dev ~/code/core-banking
```

Script chép skill, steering, hook (`dev`) vào `<repo>/.kiro/` và gộp server `xora-<vai>` vào `<repo>/.kiro/settings/mcp.json`. Chạy lại để cập nhật. Dùng `--global` thay cho repo để cài vào `~/.kiro`; khi đó `dev` không có sensor, vì hook chỉ chạy ở mức workspace.

## 3. Kiểm tra

Mở lại Kiro (hoặc **MCP Servers → Retry**), mở phiên chat mới:

- Panel MCP: `xora-<vai>` màu xanh.
- Gõ `/`: thấy skill của vai.
- Hỏi *"Hỏi AAP: bạn hỗ trợ những việc gì?"*: câu trả lời kết thúc bằng `[run_id: … | evidence: …]`.
- `dev`: chép `~/.agent-plugins/xora-dev/convention.sample.md` về gốc repo thành `convention.md`.

Kiểm ngoài Kiro:

```bash
sh -c '. ~/.xora/env; XORA_ROLE=ba python3 ~/.agent-plugins/xora-ba/mcp/xora_mcp.py --check "hello"'
```

## Sự cố

| Hiện tượng | Xử lý |
|---|---|
| Server đỏ, `python3: not found` | Sửa `python3` trong `.kiro/settings/mcp.json` thành đường dẫn tuyệt đối |
| `Missing XORA_BASE_URL / XORA_API_KEY` | Làm lại bước 1, rồi Retry |
| `HTTP 401` / `403` | API key sai hoặc hết hạn |
| `HTTP 503` | AAP tạm ngừng, đợi vài phút |
| Không thấy skill, hoặc agent theo quy tắc cũ | Mở phiên chat mới |

## Gỡ

Xoá skill, `steering/xora-<vai>.md`, hook của vai và khối `xora-<vai>` trong `<repo>/.kiro/`, rồi xoá `~/.agent-plugins/xora-<vai>`.
