# Danh mục chuẩn — hồ sơ changeset

| # | Mục | Mức | Coi là thiếu khi |
|---|---|---|---|
| 1 | Truy vết | 🔴 | Không nêu `REQ-xxx` phục vụ, thiết kế version nào, commit / PR kèm SHA |
| 2 | Phạm vi thay đổi | 🔴 | Không liệt kê file, module, cấu hình, migration CSDL bị đụng |
| 3 | **Phần do AI sinh** | 🔴 | Không có; hoặc ghi chung chung không theo file / vùng; hoặc không nói phần nào Developer chưa đọc kỹ |
| 4 | Thay đổi cấu hình | 🔴 | Có tham số mới mà không nêu theo từng môi trường. **Có giá trị secret trong hồ sơ cũng tính là lỗi** |
| 5 | Migration & rollback | 🔴 | Có migration mà không có script lùi, hoặc không nói đã chạy thử chưa, hoặc không chỉ ra điểm không lùi được |
| 6 | Kết quả kiểm tại chỗ | 🔴 | Thiếu kết quả build, lint/convention, unit test, độ phủ; hoặc kết quả đỏ bị bỏ qua không nhắc |
| 7 | Quét bảo mật | 🔴 | Không có kết quả SAST, quét phụ thuộc, quét secret; hoặc có phát hiện mà không nêu |
| 8 | Tương thích ngược | 🔴 | Có đổi API hoặc schema mà không nêu ai bị ảnh hưởng |
| 9 | Audit log & giám sát | 🔴 | Thao tác chạm tiền hoặc PII mà không nêu đã log ở đâu, trường gì, đã che PII chưa |
| 10 | Feature flag | 🟡 | Có cờ mà không nêu tên, mặc định, cách bật tắt, khi nào gỡ |
| 11 | Hướng dẫn kiểm thử thủ công | 🟡 | Không có bước, dữ liệu cần, kết quả mong đợi cho QA |
| 12 | Nợ kỹ thuật cố ý | 🟡 | Có comment đánh dấu làm tắt trong code mà không có trong mục này |
| 13 | Giới hạn đã biết | 🔴 | Không nêu cái gì chưa làm, cái gì chưa test được và vì sao |

Mục không liên quan (không có migration, không có feature flag) → chấm **không áp dụng**, ghi một dòng lý do.

## Ba mức cảnh báo

| Mức | Nghĩa |
|---|---|
| 🔴 | Thiếu là có rủi ro thật về tuân thủ, kiểm toán hoặc vận hành |
| 🟡 | Hồ sơ vẫn dùng được, người đọc sau phải đoán |
| ⚪ | Nội dung đủ, chỉ khác cấu trúc hoặc cách trình bày |

Hạ mức được nếu hậu quả thật không tồn tại — phải nêu lý do.

## Tự kiểm — 5 bước

1. Mở bảng danh mục ở trên.
2. Với từng dòng, đọc cột *Coi là thiếu khi* và chấm: **đủ** / **thiếu** / **không áp dụng**.
3. Mọi dòng đủ → in một dòng `Artifact đủ danh mục chuẩn.` rồi dừng.
4. Có dòng thiếu → in khối cảnh báo, đặt **sau** hồ sơ, không thay cho hồ sơ.
5. Người dùng trả lời → xử lý theo bảng *Xử lý câu trả lời*.

## Khối cảnh báo — in đúng mẫu

```
⚠ Hồ sơ lệch chuẩn ở 2 điểm — bạn vẫn nộp được, đây là cảnh báo.

🔴 Migration V42 chưa có script lùi
   Chuẩn cần vì: deploy lỗi trên UAT/PROD phải lùi được trong cửa sổ bảo trì.
   Để trống thì: reviewer nhiều khả năng trả lại, hoặc sự cố phải sửa nóng.
   Điền được nếu bạn trả lời: bảng mới có được drop an toàn khi lùi không?

🟡 Chưa có hướng dẫn kiểm thử thủ công
   Chuẩn cần vì: QA cần bước, dữ liệu và kết quả mong đợi để kiểm.
   Để trống thì: QA tự đoán ca kiểm, có thể bỏ sót luồng sai.
   Điền được nếu bạn trả lời: màn hình / API nào, dữ liệu mẫu nào?

Bạn chọn:
  [bổ sung]    — trả lời các câu hỏi trên, tôi viết tiếp
  [phát hành]  — giữ nguyên, tôi ghi phần lệch chuẩn vào hồ sơ
  [sửa chuẩn]  — điểm này không hợp dự án, bỏ khỏi danh mục chuẩn
```

Không có điểm lệch thì không in bảng trống.

## Xử lý câu trả lời

| Người dùng trả lời | Làm gì |
|---|---|
| `bổ sung`, hoặc trả lời câu hỏi | Viết tiếp, tăng `version`, tự kiểm lại từ bước 1 |
| `phát hành` | Nộp nguyên trạng. Thêm mục *Sai khác so với chuẩn* vào cuối hồ sơ |
| `sửa chuẩn` | Nộp nguyên trạng. Ghi: *điểm này bỏ khỏi danh mục chuẩn cho dự án này*. Dự án có file chuẩn riêng → đề xuất sửa, hỏi trước |
| Im lặng, lảng sang việc khác | Nộp nguyên trạng. Thêm mục *Sai khác so với chuẩn*, *Tình trạng* = `chưa phản hồi`, *Người chấp nhận* để trống |

Điểm đã được chấp nhận thì vòng sau không cảnh báo lại, trừ khi nội dung đổi làm rủi ro sống lại.

## Mục "Sai khác so với chuẩn"

```markdown
## Sai khác so với chuẩn

| Mục chuẩn | Mức | Tình trạng | Người chấp nhận | Thời điểm | Lý do |
|---|---|---|---|---|---|
| Quét bảo mật | 🔴 | Để sau | <tên, vai> | <YYYY-MM-DD> | Pipeline SAST chưa có trên nhánh feature |
```

Không tự điền tên ai vào cột *Người chấp nhận*.

## Bốn điều cấm

1. **Không chặn** — không từ chối nộp vì chưa đủ chuẩn.
2. **Không im lặng bỏ qua** — thiếu mục thì in cảnh báo.
3. **Không lấp cho đủ** — không bịa kết quả test, quét, SHA để mục trông như đã xong; thà để trống và cảnh báo.
4. **Không nhắc lại** thứ người dùng đã chấp nhận bỏ.
