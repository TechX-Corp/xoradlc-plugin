# Plugin `xora-dev`

Plugin cho vai Developer, chạy trên Claude Code và Kiro.

## Plugin làm được gì

| Việc | Thành phần |
|---|---|
| Viết / sửa code, cấu hình, migration, test theo BRD và thiết kế đã duyệt | Skill `implement` |
| Lập hoặc rà hồ sơ changeset (mô tả PR), ghi rõ phần do AI sinh | Skill `changeset` |
| Lưu tài liệu kỹ thuật vào KB / S3 qua agent AAP, sau khi người dùng xác nhận | Skill `upload-tech-doc` |
| Hỏi Xora DLC (việc của tôi, work order, tài liệu, open question) và KB nội bộ (convention, tiền lệ), có trích dẫn; AAP tự chọn function MCP của Xora DLC. Đăng nhập bằng `xdlc login`, không key | MCP `xora-dev`, tool `xora_ask` |
| Kiểm code theo `convention.md` của dự án (không có thì theo bản mặc định của plugin) và cảnh báo vi phạm | Sensor convention |

## Sensor chạy thế nào

1. **Kích hoạt.** Agent ghi / sửa một file code → hook của agent chạy `xdlc sensor` (dev cũng chạy tay hoặc gắn vào pre-commit).
2. **Tìm luật.** Sensor tìm `convention.md` (biến `CONVENTION_FILE`, hoặc đi ngược từ file lên tới gốc repo; cả `docs/convention.md`), đọc khối ```` ```json ```` chứa luật. Repo không có file → dùng `convention.md` mặc định đi kèm plugin (hiện chưa có luật nào, sẽ bổ sung sau).
3. **Chọn file.** Claude Code: đúng file vừa sửa (lấy từ dữ liệu hook). Kiro / chạy tay không truyền file: các file đang đổi trong git (staged, chưa staged, mới). `--staged`: chỉ phần đã staged, đọc bản trong index.
4. **Chọn dòng** (cho luật `forbid`). So với HEAD bằng `git diff`: chỉ xét dòng đã đổi, để code cũ không làm ngập cảnh báo. File mới, ngoài git, repo chưa có commit, hoặc `--full`: xét mọi dòng.
5. **Kiểm.** Mỗi luật áp cho file khớp `files` (trừ `exclude`). Bốn kiểu: `forbid` (regex không được có trên dòng nào; chỉ xét dòng đã đổi), `require` (regex phải có trong file), `filename` (tên file phải khớp regex), `max_lines` (file không quá `max` dòng); ba kiểu sau xét cả file. Dòng có `convention-ignore: <ID> <lý do>` thì bỏ qua luật đó.
6. **Báo.** In cảnh báo (file:dòng, mã luật, lỗi, cách sửa) theo dạng công cụ cần (bảng dưới). Luôn thoát mã 0: có cảnh báo, `convention.md` sai cú pháp hay regex lỗi cũng chỉ thành một cảnh báo.
7. **Agent xử lý.** Sửa vi phạm trong phần code mình vừa sinh; vi phạm ở code cũ thì chỉ liệt kê cho dev; không tự thêm `convention-ignore` khi dev chưa đồng ý; số cảnh báo còn lại đưa vào hồ sơ changeset.

| Môi trường | Sensor chạy khi | Kiểm |
|---|---|---|
| Claude Code | Hook `PostToolUse` sau `Write`, `Edit`, `MultiEdit` gọi `xdlc sensor --hook claude` | File vừa sửa |
| Kiro | Hook `PostFileSave` (`.java`, `.ts`, `.py`, `.yml`, …): Kiro nhắc agent chạy `xdlc sensor` | Các file đang đổi trong git |
| Git pre-commit (dự án tự gắn) | Developer commit: `xdlc sensor --staged` | Phần đã staged |
| Git pre-push (`xdlc` cài) | Developer push: `xdlc sensor --pre-push`, rồi gửi về Xora DLC | File thay đổi so với remote |

Hook của Kiro là file JSON v1 (`PostFileSave`, action `agent`): scope local ở `.kiro/hooks/`, scope global ở `~/.kiro/hooks/` (chạy ở mọi workspace).

## Sensor cảnh báo dev qua đâu

Mỗi cảnh báo gồm file, dòng, mã luật, lỗi và cách sửa:

```
⚠ Convention: 1 cảnh báo trong 1 file (theo convention.md v1)
- src/payment/TransferService.java:42 [LOG-001] In ra console thay vì dùng logger. → dùng logger của dự án
```

Sensor in kết quả ra stdout. Dev thấy gì là do công cụ đang chạy sensor quyết định:

| Môi trường | Cơ chế | Dev thấy | Agent nhận |
|---|---|---|---|
| Claude Code | Hook `PostToolUse`. Sensor trả JSON gồm `systemMessage` và `additionalContext` | Dòng tóm tắt trong khung chat | Toàn bộ danh sách, vào ngữ cảnh |
| Kiro | Hook `PostFileSave` (action `agent`). Agent chạy `xdlc sensor` bằng shell và đọc stdout | Những gì agent báo lại trong chat | Toàn bộ stdout |
| Pre-commit / chạy tay | Chạy trực tiếp trong terminal | Toàn bộ danh sách | — |

Sensor chỉ cảnh báo, không chặn gì: luôn thoát với mã 0, kể cả khi có cảnh báo. Sau khi nhận cảnh báo, agent tự sửa vi phạm trong phần code mình vừa sinh, và liệt kê cho dev những vi phạm nằm ở code cũ.

**Hạn chế:**
- Trong IDE, dev chỉ nhận cảnh báo qua chat, chủ yếu qua lời agent.
- Chưa hiện cảnh báo trong editor hay trong panel Problems.
- Dev tự sửa code trong editor thì không có cảnh báo, trừ khi dự án đã gắn pre-commit.

## Gửi cảnh báo về Xora DLC

Hook git `pre-push` (`xdlc sensor --pre-push`) gửi cảnh báo về `POST /governance/sensor-events`; backend ghi mỗi cảnh báo thành một governance alert của vai Dev ở bước đã map.

1. Mỗi nhánh được push: tính các file thay đổi so với commit đã có trên remote (nhánh mới: so với nhánh mặc định của remote, không có thì commit trước đó), chạy sensor với `--base <commit đó> --json`.
2. Không có cảnh báo nào: không gửi (cũng không hỏi work order).
3. Tìm work order đã map với (URL `origin`, nhánh) bằng `xdlc link`. Chưa map: hỏi trên terminal (menu Việc của tôi), không có terminal thì bỏ qua lần gửi.
4. Gửi bằng token OAuth của `xdlc login`: work order, bước, máy, repo, commit, và mỗi cảnh báo (file, dòng, mã luật, mức, nội dung, thời điểm). Không gửi nội dung dòng code.
5. Backend chỉ nhận khi người đăng nhập là Dev của work order đó (403 nếu không: chọn lại bằng `xdlc link`).

Luôn thoát mã 0: hook không bao giờ chặn push. `git push --no-verify` bỏ qua hook (chấp nhận). Chưa có: gộp alert trùng giữa các lần push và tự đóng alert khi lỗi đã sửa.
