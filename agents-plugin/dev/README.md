# Plugin `xora-dev`

Plugin cho vai Developer, chạy trên Claude Code và Kiro.

## Plugin làm được gì

| Việc | Thành phần |
|---|---|
| Viết / sửa code, cấu hình, migration, test theo BRD và thiết kế đã duyệt | Skill `implement` |
| Lập hoặc rà hồ sơ changeset (mô tả PR), ghi rõ phần do AI sinh | Skill `changeset` |
| Lưu tài liệu kỹ thuật vào KB / S3 qua agent AAP, sau khi người dùng xác nhận | Skill `upload-tech-doc` |
| Hỏi Xora DLC (việc của tôi, work order, tài liệu, open question) và KB nội bộ (convention, tiền lệ), có trích dẫn; AAP tự chọn function MCP của Xora DLC. Đăng nhập bằng `xdlc login`, không key | MCP `xora-dev`, tool `xora_ask` |
| Khi Dev nhờ agent commit/push: lấy convention từ AAP/KB, agent soi lệch, hỏi fix hoặc push + gov alert | `xdlc sensor` / `xdlc sensor report` + AGENTS.md |

## Sensor chạy thế nào

1. **Kích hoạt.** Chỉ khi Developer yêu cầu agent **commit hoặc push** (xem `AGENTS.md`). Không chạy sau mỗi lần lưu file; không có hook PostFileSave / PostToolUse cho convention.
2. **Lấy convention.** `xdlc sensor` gọi AAP (`xora_ask` qua `/mcp/user`) lấy convention dạng ngôn ngữ tự nhiên từ KB; in kèm danh sách file đang đổi trong git. Không đọc `convention.md` local.
3. **Agent phân tích.** So diff với convention; liệt kê lệch (file, dòng, rule_code, severity, message); hỏi **sửa** hay **vẫn push**.
4. **Gov alert.** Vẫn push dù còn lệch → `xdlc sensor report` (JSON findings trên stdin) → `POST /governance/sensor-events` (cần `xdlc link` map nhánh → work order).
5. **Tự `git push` trên terminal.** Plugin không kiểm convention (pre-push của tool khác vẫn chạy bình thường).
