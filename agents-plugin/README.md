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

## Vai BA và Dev

| Plugin | Skill | Thêm |
|---|---|---|
| `xora-ba` | `clarify-requirements`, `write-brd` (+ `references/brd-standard.md`), `write-user-stories`, `draw-process-flow`, `upload-kb` | `steering/brd-documents.md`: quy ước file BRD, nạp lúc mở phiên cùng `AGENTS.md` |
| `xora-dev` | `implement`, `changeset` (+ `references/changeset-standard.md`), `upload-kb` | `steering/changeset-documents.md`; **sensor convention** |

Lưu tài liệu vào KB / S3 (skill `upload-kb`) đi qua `xora_ask`: agent rà dữ liệu nhạy cảm và trạng thái duyệt, hỏi người dùng xác nhận, rồi mới gửi yêu cầu lưu cho AAP.

### Sensor convention (`xora-dev`)

Kiểm code theo `convention.md` ở gốc repo dự án mỗi khi agent ghi / sửa file (hook `PostToolUse` trên Write / Edit), **chỉ cảnh báo**. Code thuần, không LLM, không gọi mạng.

- Mẫu: copy `dev/convention.sample.md` về gốc repo thành `convention.md`, sửa cho khớp dự án (13 luật mẫu: bí mật viết cứng, private key, log dữ liệu cá nhân, catch rỗng, print / console, printStackTrace, TODO không ticket, tên file, file quá dài).
- Chạy tay / CI: `python3 dev/sensor/convention_sensor.py [file …]` (không có file → các file đang đổi trong git); `--strict` trả exit 1 khi có cảnh báo.
- Bỏ qua có chủ đích: `convention-ignore: <ID> <lý do>` trên dòng đó, và ghi vào mục *Nợ kỹ thuật* của changeset.

## Kiểm tra

`python3 agents-plugin/test_roles.py` (repo chưa có CI, chạy tay trước khi mở PR): kiểm tra manifest, symlink, frontmatter
của skill, file mà hook trỏ tới, giao thức MCP và sensor convention của `dev`, không gọi mạng.

## Còn để ngỏ

- `ba` và `dev` đã có nội dung đầy đủ. `techlead`, `qa`, `reviewer`, `delivery` vẫn là khung có `TODO(BA)`; thay khi BA gửi tài liệu.
- `xora_get_run` chưa nối: chờ AAP công bố API run.
- Danh tính đang tự khai (`XORA_USER_ID` + API key chung); chuyển sang token SSO của người dùng khi AAP có luồng xác thực.
- Tên trường `role` gửi sang AAP chưa được AAP xác nhận.
- Chưa có bản Kiro (steering + `.kiro/settings/mcp.json`) và chưa đóng gói vào tarball release.
