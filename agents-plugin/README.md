# Plugin theo vai (Claude Code ↔ agent AAP)

Sáu plugin theo vai: `xora-ba`, `xora-techlead`, `xora-dev`, `xora-qa`, `xora-reviewer`, `xora-delivery`.
Mỗi plugin có `AGENTS.md` (quy tắc của vai, nạp lúc mở phiên qua hook `SessionStart`), skills của vai
và một MCP server stdio. Cả 6 server chạy chung file `mcp/xora_mcp.py` (mỗi plugin trỏ tới nó qua symlink),
chỉ khác biến `XORA_ROLE`, và đều gọi về **cùng một agent AAP**. AAP giữ phần đọc/ghi KB và S3,
kiểm quyền theo người dùng và vai, và ghi audit.

Phần này độc lập với plugin `techx-dlc` (scope/overlay AI-DLC) ở thư mục `../aidlc-plugin`: không cần cài `aidlc`, không gọi backend xoradlc-platform.

## Cài cho Claude Code

```bash
# Do AAP cấp:
export XORA_BASE_URL=https://<platform-host>/api/v1
export XORA_API_KEY=ak_...
# XORA_AGENT mặc định là developer
# XORA_USER_ID, XORA_DLC_URL, XORA_DLC_REFRESH_TOKEN: nạp từ file .env tải ở trang xoradlc-platform
# (đăng nhập SSO Authentik rồi bấm "Tải .env"), ví dụ:
set -a; . ~/Downloads/.env; set +a
```

Nội dung steering (`AGENTS.md`) và skill nằm ở repo xoradlc-platform (`backend/app/features/plugin/content/<role>/`), không nằm trong plugin:
hook `SessionStart` chạy `xora_mcp.py --steering <role>` để in steering từ API (lỗi mạng/thiếu token thì dùng
`AGENTS.md` cục bộ của plugin); mỗi `SKILL.md` chỉ giữ frontmatter để Claude nhận diện skill, phần thân bảo
Claude gọi tool `xora_get_skill` để lấy nội dung. Plugin đổi refresh token (của Authentik) lấy access token qua `POST /auth/refresh-token`, dùng lại access token tới khi hết hạn, và giữ refresh token mới (Authentik có thể xoay vòng) trong `~/.config/xora/session.json` (quyền 0600). Khi plugin báo phiên hết hạn, tải lại `.env`.
Sửa nội dung thì sửa bên xoradlc-platform và phát hành lại backend, không cần cài lại plugin.

```
/plugin marketplace add <đường dẫn clone xoradlc-plugin>/agents-plugin
/plugin install xora-dev@xora-role-plugins
```

Kiểm tra kết nối: `python3 mcp/xora_mcp.py --check "hello"`.

## Kiểm tra

`python3 agents-plugin/test_roles.py` (repo chưa có CI, chạy tay trước khi mở PR): kiểm tra manifest, symlink, frontmatter
của skill và giao thức MCP, không gọi mạng.

## Còn để ngỏ

- Nội dung steering/skill vẫn là khung có `TODO(BA)`; thay ở xoradlc-platform khi BA gửi tài liệu.
- `AGENTS.md` trong plugin chỉ còn là bản dự phòng, cần giữ đồng bộ tay với bản trong xoradlc-platform.
- Thời hạn refresh token do Authentik quyết định; hết hạn thì tải lại `.env` từ trang xoradlc-platform.
- `xora_get_run` chưa nối: chờ AAP công bố API run.
- Danh tính đang tự khai (`XORA_USER_ID` + API key chung); chuyển sang token SSO của người dùng khi AAP có luồng xác thực.
- Tên trường `role` gửi sang AAP chưa được AAP xác nhận.
- Chưa có bản Kiro (steering + `.kiro/settings/mcp.json`) và chưa đóng gói vào tarball release.
