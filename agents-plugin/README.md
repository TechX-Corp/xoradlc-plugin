# Plugin theo vai (Claude Code / Kiro ↔ agent AAP)

Hiện chỉ dùng plugin `xora-dev` (vai Developer): skill `implement`, `changeset`, `upload-tech-doc`, sensor convention, và MCP server `xora-dev` chạy `mcp/xora_mcp.py` để gọi về **agent AAP**. AAP lo đọc / ghi KB, S3, kiểm quyền và ghi audit. Chi tiết: [dev/README.md](dev/README.md).

Các vai khác (`ba`, `techlead`, `qa`, `reviewer`, `delivery`) đã chuyển vào [archived/](archived/), không có trong marketplace và không được kiểm. Muốn dùng lại: chuyển thư mục về cạnh `dev/`, sửa symlink `mcp/xora_mcp.py` thành `../../mcp/xora_mcp.py`, thêm vào `marketplace.json` và `ROLES` trong `test_roles.py`.

## Kết nối AAP

Đặt các biến sau (Claude Code: `export` trong shell; Kiro: file `~/.xora/env`, xem [INSTALL-KIRO.md](INSTALL-KIRO.md)):

```bash
export XORA_API=runs
export XORA_BASE_URL=https://d1xbzyews27i3w.cloudfront.net/v1
export XORA_AGENT=agent_mw3at9y00
export XORA_API_KEY=xap_...
```

Kiểm tra: `python3 mcp/xora_mcp.py --check "hello"`.

## Cài cho Claude Code

```
/plugin marketplace add <đường dẫn clone xoradlc-plugin>/agents-plugin
/plugin install xora-dev@xora-role-plugins
```

## Cài cho Kiro

`sh agents-plugin/install-kiro.sh dev <repo dự án>`. Chi tiết: [INSTALL-KIRO.md](INSTALL-KIRO.md).

## Sensor convention (`dev`)

Sau mỗi lần agent ghi / sửa code, sensor kiểm theo `convention.md` ở gốc repo và **chỉ cảnh báo**.

- Tạo `convention.md`: copy `dev/convention.sample.md` về gốc repo, sửa cho khớp dự án.
- Luật `forbid` chỉ báo trên dòng đã đổi so với HEAD; `--full` để rà cả code cũ.
- Chạy tay / CI: `python3 dev/sensor/convention_sensor.py [file …] [--staged] [--full]` (`--staged` cho git pre-commit).
- Plugin dev làm được gì: [dev/README.md](dev/README.md).
- Bỏ qua một dòng: `convention-ignore: <ID> <lý do>`.

## Kiểm tra trước khi mở PR

`python3 agents-plugin/test_roles.py`

## Còn để ngỏ

- Danh tính người dùng: AAP chỉ nhận ID token Cognito qua header `X-Xora-User-Token`; MCP chưa gửi token này.
- `XORA_API=invoke` (mặc định): API chưa có host công khai.
