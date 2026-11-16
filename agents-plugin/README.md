# Plugin theo vai (Claude Code / Kiro ↔ Xora DLC ↔ agent AAP)

Hiện chỉ dùng plugin `xora-dev` (vai Developer): skill `implement`, `changeset`, `upload-tech-doc`, sensor convention, và MCP từ xa `xora-dev` (`https://xoradlc-platform-api.techxcorp.com/mcp/user`). Tool `xora_ask` gửi câu hỏi tới **backend Xora DLC** như chat trên web: backend xác định người hỏi qua đăng nhập OAuth, gắn vai và phiên MCP rồi chuyển cho **agent AAP** của vai đó. AAP tự chọn function MCP của Xora DLC cần gọi (việc của tôi, work order, tài liệu, open question) và tra KB, S3, kiểm quyền, ghi audit. Chi tiết: [dev/README.md](dev/README.md).

Các vai khác (`ba`, `techlead`, `qa`, `reviewer`, `delivery`) đã chuyển vào [archived/](archived/), không có trong marketplace và không được kiểm. Muốn dùng lại: chuyển thư mục về cạnh `dev/`, đổi `.mcp.json` và `mcp.json` sang URL `/mcp/user` như `dev/` (các file này còn trỏ tới `xora_mcp.py` đã bỏ), thêm vào `marketplace.json` và `ROLES` trong `test_roles.py`.

## Cài đặt

Một lệnh, không cần key. Chạy trong terminal (cần `python3`; Claude Code cần thêm `claude`):

```bash
# Kiro CLI, cho một repo dự án (hoặc --global thay cho đường dẫn: mọi repo, không có sensor)
curl -fsSL https://raw.githubusercontent.com/TechX-Corp/xoradlc-plugin/main/install.sh | sh -s -- kiro ~/code/<repo>

# Claude Code (scope user mặc định; project hoặc local để giới hạn)
curl -fsSL https://raw.githubusercontent.com/TechX-Corp/xoradlc-plugin/main/install.sh | sh -s -- claude
```

Mở phiên Kiro / Claude Code mới và hỏi. **Lần đầu** Kiro mở trình duyệt tới trang đăng nhập Xora DLC (SSO Authentik hoặc tài khoản local, như trên web); trong Claude Code gõ `/mcp` rồi chọn đăng nhập. Token do Kiro / Claude lưu và tự làm mới.

Tuỳ chọn: `XORA_URL=http://localhost:8000` (backend khác), `XORA_PLUGIN_REF=<nhánh|tag>`, `XORA_PLUGIN_DIR=<bản clone có sẵn>`. Chi tiết Kiro: [INSTALL-KIRO.md](INSTALL-KIRO.md).

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

- Câu hỏi qua Xora DLC tối đa 3.800 ký tự (giới hạn `question` của backend), nên `upload-tech-doc` chỉ lưu được tài liệu ngắn hoặc phải tách phần.
- Sensor chưa gửi cảnh báo về Xora DLC (xem [dev/README.md](dev/README.md)).
