# Plugin theo vai (Claude Code ↔ agent AAP)

Sáu plugin theo vai: `xora-ba`, `xora-techlead`, `xora-dev`, `xora-qa`, `xora-reviewer`, `xora-delivery`.
Mỗi plugin có `AGENTS.md` (quy tắc của vai, nạp lúc mở phiên qua hook `SessionStart`), skills của vai
và một MCP server stdio. Cả 6 server chạy chung file `mcp/xora_mcp.py` (mỗi plugin trỏ tới nó qua symlink),
chỉ khác biến `XORA_ROLE`, và đều gọi về **cùng một agent AAP**. AAP giữ phần đọc/ghi KB và S3,
kiểm quyền theo người dùng và vai, và ghi audit.

Phần này độc lập với plugin `techx-dlc` (scope/overlay AI-DLC) ở thư mục `../aidlc-plugin`: không cần cài `aidlc`, không gọi backend xora-dlc.

## Cài cho Claude Code

```bash
export XORA_BASE_URL=https://<platform-host>/api/v1
export XORA_API_KEY=ak_...
export XORA_USER_ID=<email công việc>
# XORA_AGENT mặc định là developer
```

```
/plugin marketplace add <đường dẫn clone xoradlc-plugin>/agents-plugin
/plugin install xora-dev@xora-role-plugins
```

Kiểm tra kết nối: `python3 mcp/xora_mcp.py --check "hello"`.

## Kiểm tra

`python3 agents-plugin/test_roles.py` (repo chưa có CI, chạy tay trước khi mở PR): kiểm tra manifest, symlink, frontmatter
của skill và giao thức MCP, không gọi mạng.

## Còn để ngỏ

- Nội dung `AGENTS.md` và `SKILL.md` là khung, có `TODO(BA)`; thay khi BA gửi tài liệu.
- `xora_get_run` chưa nối: chờ AAP công bố API run.
- Danh tính đang tự khai (`XORA_USER_ID` + API key chung); chuyển sang token SSO của người dùng khi AAP có luồng xác thực.
- Tên trường `role` gửi sang AAP chưa được AAP xác nhận.
- Chưa có bản Kiro (steering + `.kiro/settings/mcp.json`) và chưa đóng gói vào tarball release.
