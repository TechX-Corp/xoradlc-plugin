# Ghi chú cho người bảo trì

Phần này không cần cho người dùng plugin; hướng dẫn dùng ở [README.md](README.md).

## Cách plugin chạy

- `xdlc` (một file Python ở gốc repo) cài plugin `xora-dev` (`dev/`) cho Kiro CLI hoặc Claude Code, giữ token OAuth trong `~/.xdlc/auth.json` (Windows: DPAPI) và chạy MCP server cục bộ `xdlc mcp` (stdio). `xdlc mcp` chuyển mọi yêu cầu tới `<backend>/mcp/user` kèm token.
- Tool `xora_ask`: backend Xora DLC nhận câu hỏi như chat trên web, gắn vai và phiên MCP rồi chuyển cho agent AAP của vai đó; AAP tự chọn function MCP của Xora DLC.
- Sensor convention: [dev/README.md](dev/README.md#sensor-chạy-thế-nào). Gửi cảnh báo khi push: [dev/README.md](dev/README.md#gửi-cảnh-báo-về-xora-dlc).

## Kiểm tra trước khi mở PR

```bash
python3 agents-plugin/test_roles.py
claude plugin validate agents-plugin/dev
```

CI (`.github/workflows/xdlc-plugin-tests.yml`) chạy `test_roles.py` trên Ubuntu, macOS, Windows với Python 3.9 và 3.13. Mục "Các lệnh" của README phải khớp đúng output của `xdlc` (test kiểm). `install.ps1` chỉ được có ký tự ASCII (PowerShell 5.1 đọc theo ANSI).

## Phát hành

1. Tăng `VERSION` trong `xdlc` và `version` trong `.claude-plugin/marketplace.json` (phải khớp, test kiểm).
2. Merge vào `main`, tạo tag `vX.Y.Z`. `install.sh` / `install.ps1` và `xdlc update` lấy tag mới nhất.

Thử bản chưa phát hành: `XDLC_REF=<nhánh> xdlc update`, hoặc `XDLC_PLUGIN_DIR=<bản clone>` cho `install.sh` / `install.ps1` / `xdlc setup` / `xdlc update`.

## Các vai đã lưu trữ

`ba`, `techlead`, `qa`, `reviewer`, `delivery` ở [archived/](archived/), không có trong marketplace và không được kiểm. Dùng lại: chuyển thư mục về cạnh `dev/`, đổi `.mcp.json` và `mcp.json` sang `xdlc mcp` như `dev/`, thêm vào `kiro_files` của `xdlc`, `marketplace.json` và `ROLES` trong `test_roles.py`.

## Giới hạn đã biết

- `dev/convention.md` là convention mặc định (sensor dùng khi repo không có file riêng): hiện `rules` rỗng, sẽ bổ sung sau.

- `xora_ask` nhận tối đa 3.800 ký tự (giới hạn `question` của backend): `upload-tech-doc` chỉ lưu được tài liệu ngắn hoặc phải tách phần.
- Mỗi lần push gửi lại mọi cảnh báo còn lại: backend chưa gộp alert trùng, chưa tự đóng alert khi lỗi đã sửa.
- Pre-push kiểm bản trong thư mục làm việc của các file thay đổi, không phải đúng bản trong commit được push.
- Backend không giữ danh sách token: muốn chặn hẳn một người thì khoá tài khoản.
