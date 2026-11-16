# Cài plugin theo vai vào Kiro

Cần: Kiro CLI, `python3` ≥ 3.9, macOS / Linux (Windows dùng WSL), tài khoản Xora DLC (SSO hoặc local).

## 1. Cài (một lệnh)

```bash
curl -fsSL https://raw.githubusercontent.com/TechX-Corp/xoradlc-plugin/main/install.sh | sh -s -- kiro ~/code/core-banking
```

Đích là repo dự án mở trong Kiro; `--global` thay cho đường dẫn để cài vào `~/.kiro` (khi đó `dev` không có sensor, vì hook chỉ chạy ở mức workspace). Script chép skill, steering, hook vào `<repo>/.kiro/` và gộp server `xora-dev` (URL + `oauth.clientId`) vào `<repo>/.kiro/settings/mcp.json`. Chạy lại để cập nhật. Không có key nào phải tạo hay dán.

Đã clone repo thì `sh agents-plugin/install-kiro.sh dev <repo>` làm cùng việc đó; `XORA_URL=http://localhost:8000` đổi backend.

## 2. Đăng nhập (lần đầu)

Mở phiên Kiro mới trong repo và hỏi *"Việc của tôi trong Xora DLC là gì?"*. Kiro mở trình duyệt tới trang đăng nhập Xora DLC: chọn SSO hoặc nhập tài khoản local. Xong, trình duyệt báo quay lại terminal. Token được Kiro lưu và tự làm mới; `/mcp logout` để đăng xuất, `/mcp auth` để đăng nhập lại.

## 3. Kiểm tra

- Panel MCP (`/mcp`): `xora-dev` hiện tool `xora_ask`.
- Hỏi *"Việc của tôi trong Xora DLC là gì?"*: Kiro gọi `xora_ask`, câu trả lời kết thúc bằng `[session_id: … | run_id: … | evidence: …]`.
- `dev`: chép `~/.agent-plugins/xora-dev/convention.sample.md` về gốc repo thành `convention.md`.

## Sự cố

| Hiện tượng | Xử lý |
|---|---|
| Server `xora-dev` báo cần xác thực, trình duyệt không mở | `/mcp auth` trong Kiro, hoặc mở URL Kiro in ra |
| Trình duyệt báo `redirect_uri` hoặc `client_id` không hợp lệ | `oauth.clientId` trong `.kiro/settings/mcp.json` phải là `xora-plugin`: chạy lại lệnh cài |
| Đăng nhập xong vẫn 401 | Tài khoản bị khoá, hoặc đổi sang tài khoản khác: `/mcp logout` rồi `/mcp auth` |
| `HTTP 403` `write on ai_runs required` | Vai của tài khoản chưa được hỏi agent: nhờ admin Xora DLC cấp quyền |
| `agent_not_configured` | Xora DLC chưa nối agent AAP: báo admin |
| `agent_failed` | AAP lỗi hoặc quá thời gian: đợi vài phút rồi hỏi lại |
| Không thấy skill, hoặc agent theo quy tắc cũ | Mở phiên chat mới |

## Gỡ

Xoá skill, `steering/xora-<vai>.md`, hook của vai và khối `xora-<vai>` trong `<repo>/.kiro/`, rồi xoá `~/.agent-plugins/xora-<vai>` và `~/.xora/plugin`.
