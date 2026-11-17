# Plugin `xora-dev`

Plugin cho vai Developer, chạy trên Claude Code và Kiro.

## Plugin làm được gì

| Việc | Thành phần |
|---|---|
| Viết / sửa code, cấu hình, migration, test theo BRD và thiết kế đã duyệt | Skill `implement` |
| Lập hoặc rà hồ sơ changeset (mô tả PR), ghi rõ phần do AI sinh | Skill `changeset` |
| Lưu tài liệu kỹ thuật vào KB / S3 qua agent AAP, sau khi người dùng xác nhận | Skill `upload-tech-doc` |
| Hỏi Xora DLC (việc của tôi, work order, tài liệu, open question) và KB nội bộ (convention, tiền lệ), có trích dẫn; AAP tự chọn function MCP của Xora DLC. Đăng nhập bằng `xdlc login`, không key | MCP `xora-dev`, tool `xora_ask` |
| Kiểm code theo `convention.md` của dự án và cảnh báo vi phạm | Sensor convention |

## Sensor chạy thế nào

1. **Kích hoạt.** Agent ghi / sửa một file code → hook của agent chạy `xdlc sensor` (dev cũng chạy tay hoặc gắn vào pre-commit).
2. **Tìm luật.** Sensor tìm `convention.md` (biến `CONVENTION_FILE`, hoặc đi ngược từ file lên tới gốc repo; cả `docs/convention.md`), đọc khối ```` ```json ```` chứa luật. Không có file → in lời nhắc tạo `convention.md`, không kiểm gì.
3. **Chọn file.** Claude Code: đúng file vừa sửa (lấy từ dữ liệu hook). Kiro / chạy tay không truyền file: các file đang đổi trong git (staged, chưa staged, mới). `--staged`: chỉ phần đã staged, đọc bản trong index.
4. **Chọn dòng** (cho luật `forbid`). So với HEAD bằng `git diff`: chỉ xét dòng đã đổi, để code cũ không làm ngập cảnh báo. File mới, ngoài git, repo chưa có commit, hoặc `--full`: xét mọi dòng.
5. **Kiểm.** Mỗi luật áp cho file khớp `files` (trừ `exclude`). Bốn kiểu: `forbid` (regex không được có trên dòng nào; chỉ xét dòng đã đổi), `require` (regex phải có trong file), `filename` (tên file phải khớp regex), `max_lines` (file không quá `max` dòng); ba kiểu sau xét cả file. Dòng có `convention-ignore: <ID> <lý do>` thì bỏ qua luật đó.
6. **Báo.** In cảnh báo (file:dòng, mã luật, lỗi, cách sửa) theo dạng công cụ cần (bảng dưới). Luôn thoát mã 0: có cảnh báo, `convention.md` sai cú pháp hay regex lỗi cũng chỉ thành một cảnh báo.
7. **Agent xử lý.** Sửa vi phạm trong phần code mình vừa sinh; vi phạm ở code cũ thì chỉ liệt kê cho dev; không tự thêm `convention-ignore` khi dev chưa đồng ý; số cảnh báo còn lại đưa vào hồ sơ changeset.

| Môi trường | Sensor chạy khi | Kiểm |
|---|---|---|
| Claude Code | Hook `PostToolUse` sau `Write`, `Edit`, `MultiEdit` gọi `xdlc sensor --hook claude` | File vừa sửa |
| Kiro | Hook `fileEdited` (`.java`, `.ts`, `.py`, `.yml`, …): Kiro nhắc agent chạy `xdlc sensor` | Các file đang đổi trong git |
| Git pre-commit (dự án tự gắn) | Developer commit: `xdlc sensor --staged` | Phần đã staged |

Kiro chỉ chạy hook theo workspace, nên sensor của Kiro có ở scope `local`, không có ở `global`.

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
| Kiro | Hook `fileEdited` (`askAgent`). Agent chạy `xdlc sensor` bằng shell và đọc stdout | Những gì agent báo lại trong chat | Toàn bộ stdout |
| Pre-commit / chạy tay | Chạy trực tiếp trong terminal | Toàn bộ danh sách | — |

Sensor chỉ cảnh báo, không chặn gì: luôn thoát với mã 0, kể cả khi có cảnh báo. Sau khi nhận cảnh báo, agent tự sửa vi phạm trong phần code mình vừa sinh, và liệt kê cho dev những vi phạm nằm ở code cũ.

**Hạn chế:**
- Trong IDE, dev chỉ nhận cảnh báo qua chat, chủ yếu qua lời agent.
- Chưa hiện cảnh báo trong editor hay trong panel Problems.
- Dev tự sửa code trong editor thì không có cảnh báo, trừ khi dự án đã gắn pre-commit.

## Gửi cảnh báo về Xora DLC (chưa làm)

Hiện sensor chỉ in cảnh báo trên máy dev; Xora DLC (màn hình Governance) không thấy.

Backend đã có `POST /governance/sensor-events`: nhận một lần chạy sensor (`work_order_id`, `step_id` mặc định `impl`, máy, phiên, repo, commit, danh sách cảnh báo: file, dòng, mã luật, mức, nội dung, thời điểm; không có nội dung dòng code) và ghi thành governance alert của vai Dev ở bước đó. Plugin chưa gọi vì còn ba chỗ hở:

- **Xác thực:** endpoint nhận key `sensor` (`xmcp_…`, admin cấp riêng cho từng dev), không nhận token OAuth mà `xdlc login` giữ. Cần backend nhận thêm token đó (người gửi là người đăng nhập), để dev không phải xin thêm key.
- **Work order:** mỗi lần gửi phải có `work_order_id`, nhưng sensor không biết repo / nhánh đang làm thuộc work order nào. Cần quy ước (vd. ghi mã WO trong cấu hình repo hay tên nhánh) hoặc `xdlc` hỏi một lần rồi nhớ.
- **Quyền:** backend chỉ nhận khi người gửi là Dev được giao của work order đó; gửi sai work order sẽ bị từ chối.
