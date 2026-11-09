# Plugin theo vai (Claude Code / Kiro ↔ agent AAP)

Sáu plugin theo vai: `xora-ba`, `xora-techlead`, `xora-dev`, `xora-qa`, `xora-reviewer`, `xora-delivery`. Mỗi plugin gồm `AGENTS.md` (quy tắc của vai), skill của vai và MCP server `xora-<vai>`. Cả 6 server chạy chung `mcp/xora_mcp.py`, chỉ khác `XORA_ROLE`, và gọi về **một agent AAP**. Agent này lo đọc / ghi KB, S3, kiểm quyền và ghi audit.

| Vai | Skill |
|---|---|
| `ba` | `clarify-requirements`, `write-brd`, `write-user-stories`, `draw-process-flow`, `upload-kb` |
| `dev` | `implement`, `changeset`, `upload-tech-doc` + sensor convention |
| `techlead`, `qa`, `reviewer`, `delivery` | Khung, còn `TODO(BA)` |

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

`sh agents-plugin/install-kiro.sh <vai> <repo dự án>`. Chi tiết: [INSTALL-KIRO.md](INSTALL-KIRO.md).

## Sensor convention (`dev`)

Sau mỗi lần agent ghi / sửa code, sensor kiểm theo `convention.md` ở gốc repo và **chỉ cảnh báo**.

- Tạo `convention.md`: copy `dev/convention.sample.md` về gốc repo, sửa cho khớp dự án.
- Chạy tay / CI: `python3 dev/sensor/convention_sensor.py [file …] [--strict]`.
- Bỏ qua một dòng: `convention-ignore: <ID> <lý do>`.

## Kiểm tra trước khi mở PR

`python3 agents-plugin/test_roles.py`

## Còn để ngỏ

- Danh tính người dùng: AAP chỉ nhận ID token Cognito qua header `X-Xora-User-Token`; MCP chưa gửi token này.
- `XORA_API=invoke` (mặc định): API chưa có host công khai.
