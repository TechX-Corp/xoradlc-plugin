---
name: upload-kb
description: Lưu một tài liệu (BRD, user story, sơ đồ luồng, biên bản, tài liệu nghiệp vụ) vào KB / S3 qua agent AAP bằng tool xora_ask — rà trạng thái duyệt và dữ liệu nhạy cảm, hỏi xác nhận rồi mới gửi yêu cầu lưu, báo đúng kết quả AAP trả về. Dùng khi người dùng nói "đẩy / đưa / upload / lưu tài liệu này vào KB".
---

# Lưu tài liệu vào KB / S3

KB là dữ liệu dùng chung của tổ chức: tài liệu đã vào thì agent và người khác sẽ dùng làm căn cứ. **Không bao giờ gửi yêu cầu lưu khi người dùng chưa xác nhận trong lượt này.**

Ghi đi qua cùng tool `xora_ask` của MCP server `xora-ba`: bạn gửi toàn văn tài liệu kèm yêu cầu lưu rõ ràng, **agent AAP** thực hiện ghi, kiểm quyền theo người dùng và vai, chỉ cho artifact đã phê duyệt lên S3, và ghi audit.

## Bước 1 — Chuẩn bị

1. Xác định file cần lưu. Người dùng nói chung chung ("lưu BRD lên KB") mà có nhiều file khớp → liệt kê và hỏi file nào.
2. Đọc **toàn văn** file. Lấy `version`, `status` từ frontmatter nếu có.
3. Đặt tên tài liệu: theo `id` hoặc tiêu đề H1, kèm version — ví dụ `BRD-042 Chuyển khoản liên ngân hàng v3`.

## Bước 2 — Rà trước khi lưu

| Kiểm | Gặp thì |
|---|---|
| `status` khác `approved` (draft, awaiting_review) hoặc không có frontmatter | Cảnh báo: AAP chỉ đưa artifact đã phê duyệt lên S3, và tài liệu chưa duyệt trong KB sẽ thành căn cứ cho người khác. Chỉ gửi khi người dùng xác nhận rõ là vẫn lưu bản chưa duyệt |
| Dữ liệu thật của khách hàng (CCCD, số tài khoản, tên, số điện thoại), secret, token | **Dừng.** Chỉ ra vị trí, đề nghị thay bằng dữ liệu bịa. Không gửi khi chưa xử lý |
| Còn `[CẦN XÁC NHẬN]` / `[CẦN TUÂN THỦ XÁC NHẬN]` | Nêu số lượng và mục; người dùng quyết có lưu không |
| Dài hơn ~18.000 ký tự (giới hạn `text` của `xora_ask` là 20.000, chừa chỗ cho yêu cầu) | Đề xuất tách theo mục lớn, lưu từng phần với tên `<tên> — phần i/n` |

## Bước 3 — Xác nhận

Luôn hỏi đúng mẫu, rồi **dừng chờ** trả lời:

```
Sắp gửi agent AAP lưu vào KB / S3:
- Tài liệu: BRD-042 Chuyển khoản liên ngân hàng v3
- File: brd/BRD-042-chuyen-khoan-lien-ngan-hang.md · status: approved · 12.480 ký tự
- Lưu ý: còn 2 [CẦN XÁC NHẬN] ở mục 12
Gõ "lưu" để xác nhận, hoặc nói cần sửa gì.
```

Chỉ coi là đồng ý khi người dùng trả lời rõ (*lưu*, *đẩy*, *đồng ý*). Im lặng, "để sau", hay câu mơ hồ → không gửi. Xác nhận cho tài liệu này không áp dụng cho tài liệu khác.

## Bước 4 — Gửi và báo kết quả

Gọi `xora_ask` **không** truyền `session_id` (yêu cầu ghi tách khỏi hội thoại tra cứu), với `text` theo mẫu:

```
Yêu cầu LƯU tài liệu vào KB / S3. Người dùng đã xem và xác nhận.
Tên tài liệu: <tên>
Nguồn: <đường dẫn file> · version: <v> · status: <status>
Sau khi xử lý, trả lời rõ: đã lưu hay chưa; mã / đường dẫn tài liệu trong KB hoặc key trên S3 nếu có; nếu không lưu được thì nêu lý do. Không tóm tắt hay sửa nội dung.
<<<TÀI LIỆU
<toàn văn>
TÀI LIỆU>>>
```

Đọc câu trả lời của AAP và báo đúng như nó nói:

| AAP trả lời | Báo người dùng |
|---|---|
| Đã lưu, có mã / đường dẫn / key | Đã lưu, kèm mã và `session_id`. Tra lại ngay có thể chưa thấy nếu KB còn đang index |
| Báo đã lưu nhưng không có mã | AAP báo đã lưu nhưng chưa có mã để kiểm; đề nghị lát nữa tìm lại bằng `xora_ask` theo tên tài liệu |
| Không lưu được, thiếu quyền, chưa phê duyệt, hoặc không chắc | Chưa lưu, nêu lý do AAP đưa ra. **Không nói là đã lưu** |
| Lỗi HTTP / hết thời gian chờ | Chưa chắc đã lưu. **Không gửi lại** khi chưa biết lần trước ra sao — tránh trùng tài liệu; hỏi AAP theo tên tài liệu trước |

Không tự ghi "đã đưa vào KB" vào tài liệu hay nhật ký thay đổi khi AAP chưa xác nhận đã lưu.
