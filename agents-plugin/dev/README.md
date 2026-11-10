# Plugin `xora-dev`

Plugin cho vai Developer, chạy trên Claude Code và Kiro.

## Plugin làm được gì

| Việc | Thành phần |
|---|---|
| Viết / sửa code, cấu hình, migration, test theo BRD và thiết kế đã duyệt | Skill `implement` |
| Lập hoặc rà hồ sơ changeset (mô tả PR), ghi rõ phần do AI sinh | Skill `changeset` |
| Lưu tài liệu kỹ thuật vào KB / S3 qua agent AAP, sau khi người dùng xác nhận | Skill `upload-tech-doc` |
| Tra cứu KB nội bộ (convention, tiền lệ, BRD / thiết kế) có trích dẫn | MCP `xora-dev`, tool `xora_ask` |
| Kiểm code theo `convention.md` của dự án và cảnh báo vi phạm | Sensor convention |

## Sensor lắng nghe gì

| Môi trường | Sensor chạy khi | Kiểm |
|---|---|---|
| Claude Code | Agent ghi / sửa file bằng `Write`, `Edit`, `MultiEdit` | File vừa sửa |
| Kiro | Một file code (`.java`, `.ts`, `.py`, `.yml`, …) được sửa | Các file đang đổi trong git |
| Git pre-commit (dự án tự gắn) | Developer commit | Phần đã staged |

Sensor chỉ báo trên **dòng code đã đổi**, không báo code cũ. Chưa có `convention.md` ở gốc repo thì sensor không chạy.

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
| Kiro | Hook `fileEdited` (`askAgent`). Agent chạy sensor bằng shell và đọc stdout | Những gì agent báo lại trong chat | Toàn bộ stdout |
| Pre-commit / chạy tay | Chạy trực tiếp trong terminal | Toàn bộ danh sách | — |

Sensor chỉ cảnh báo, không chặn gì: luôn thoát với mã 0, kể cả khi có cảnh báo. Sau khi nhận cảnh báo, agent tự sửa vi phạm trong phần code mình vừa sinh, và liệt kê cho dev những vi phạm nằm ở code cũ.

**Hạn chế:**
- Trong IDE, dev chỉ nhận cảnh báo qua chat, chủ yếu qua lời agent.
- Chưa hiện cảnh báo trong editor hay trong panel Problems.
- Dev tự sửa code trong editor thì không có cảnh báo, trừ khi dự án đã gắn pre-commit.

## Gửi cảnh báo về XoraDLC (đang build)

XoraDLC sẽ cung cấp MCP để sensor gọi vào báo nội dung cảnh báo. Mỗi lần sensor chạy sẽ gửi: người dùng, project, branch, commit, và danh sách cảnh báo (file, dòng, mã luật). Không gửi nội dung dòng code.
