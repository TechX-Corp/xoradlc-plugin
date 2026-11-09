# Danh mục chuẩn BRD

| # | Mục | Mức | Coi là thiếu khi |
|---|---|---|---|
| 1 | Bối cảnh & mục tiêu nghiệp vụ | 🟡 | Không nêu được đo thành công bằng chỉ số gì |
| 2 | Phạm vi | 🔴 | Mục "ngoài phạm vi" trống |
| 3 | Đối tượng sử dụng & quyền | 🟡 | Không nêu ai nhập, ai duyệt, hạn mức theo cấp |
| 4 | Yêu cầu chức năng | 🔴 | Có yêu cầu không mang ID `REQ-<epic>-<nnn>`, hoặc không có tiêu chí chấp nhận |
| 5 | Yêu cầu phi chức năng | 🔴 | Hiệu năng, khả dụng, audit không có số đo; hoặc không mang ID `REQ-xxx` |
| 6 | Yêu cầu tuân thủ & kiểm soát | 🔴 | Nghiệp vụ chạm tiền hoặc PII mà không nêu maker–checker, phân tách trách nhiệm, nghĩa vụ báo cáo |
| 7 | Luồng nghiệp vụ & ngoại lệ | 🔴 | Không mô tả khi hệ thống phụ thuộc không phản hồi |
| 8 | Quy tắc dữ liệu | 🔴 | Có trường dữ liệu mà không nêu kiểu, bắt buộc hay không, PII hay không, quy tắc che, thời hạn lưu |
| 9 | Ràng buộc | 🟡 | Không nêu hệ thống hiện có, mốc thời gian, cut-off |
| 10 | Thuật ngữ | 🟡 | Có khái niệm mới không được định nghĩa, hoặc một thứ gọi bằng hai từ |
| 11 | Giả định cần xác nhận | 🟡 | Có `[GIẢ ĐỊNH]` trong thân bài mà không có trong bảng |
| 12 | Câu hỏi còn mở | 🔴 | Có `[CẦN XÁC NHẬN]` trong thân bài mà không có trong bảng; hoặc bảng thiếu cột *câu hỏi \| đề xuất \| căn cứ \| ai trả lời \| chặn REQ nào* |
| 13 | Nhật ký thay đổi | 🟡 | `version ≥ 2` mà không có mục này |

Mục 6 hạ xuống 🟡 được nếu nghiệp vụ chỉ đọc dữ liệu, không chạm tiền — phải nêu lý do trong khối cảnh báo.

## Ba mức cảnh báo

| Mức | Nghĩa |
|---|---|
| 🔴 | Thiếu là có rủi ro thật về tuân thủ, kiểm toán hoặc vận hành |
| 🟡 | Tài liệu vẫn dùng được, người đọc sau phải đoán |
| ⚪ | Nội dung đủ, chỉ khác cấu trúc hoặc cách trình bày |

Hạ mức được nếu hậu quả thật không tồn tại — phải nêu lý do.

## Tự kiểm — 5 bước

1. Mở bảng danh mục ở trên.
2. Với từng dòng, đọc cột *Coi là thiếu khi* và chấm: **đủ** / **thiếu** / **không áp dụng**.
3. Mọi dòng đủ → in một dòng `Artifact đủ danh mục chuẩn.` rồi dừng.
4. Có dòng thiếu → in khối cảnh báo, đặt **sau** tài liệu, không thay cho tài liệu.
5. Người dùng trả lời → xử lý theo bảng *Xử lý câu trả lời*.

Dòng **không áp dụng** vẫn nêu một dòng lý do, mức ⚪.

## Khối cảnh báo — in đúng mẫu

```
⚠ Tài liệu lệch chuẩn ở 3 điểm — bạn vẫn phát hành được, đây là cảnh báo.

🔴 Thiếu mục "Yêu cầu tuân thủ & kiểm soát"
   Chuẩn cần vì: nghiệp vụ chạm tiền, kiểm toán nội bộ sẽ hỏi maker–checker.
   Để trống thì: cổng review nhiều khả năng trả lại, mất thêm một vòng.
   Điền được nếu bạn trả lời: ai nhập, ai duyệt, hạn mức theo cấp?

🟡 REQ-042-003 chưa có tiêu chí chấp nhận cho luồng sai
   Chuẩn cần vì: QA dựng test case từ tiêu chí chấp nhận.
   Để trống thì: QA tự nghĩ ca test, có thể lệch ý bạn.
   Điền được nếu bạn trả lời: dữ liệu sai định dạng thì hệ thống làm gì?

⚪ Gộp mục 5 và 6 theo yêu cầu của bạn ở vòng trước
   Giữ nguyên, chỉ ghi nhận.

Bạn chọn:
  [bổ sung]    — trả lời các câu hỏi trên, tôi viết tiếp
  [phát hành]  — giữ nguyên, tôi ghi phần lệch chuẩn vào tài liệu
  [sửa chuẩn]  — điểm này không hợp dự án, bỏ khỏi danh mục chuẩn
```

Không có điểm lệch thì không in bảng trống.

## Xử lý câu trả lời

| Người dùng trả lời | Làm gì |
|---|---|
| `bổ sung`, hoặc trả lời câu hỏi | Viết tiếp, tăng `version`, tự kiểm lại từ bước 1 |
| `phát hành` | Nộp nguyên trạng. Thêm mục *Sai khác so với chuẩn* vào cuối tài liệu |
| `sửa chuẩn` | Nộp nguyên trạng. Ghi: *điểm này bỏ khỏi danh mục chuẩn cho dự án này*. Dự án có file chuẩn riêng → đề xuất sửa, hỏi trước |
| Im lặng, lảng sang việc khác | Nộp nguyên trạng. Thêm mục *Sai khác so với chuẩn*, *Tình trạng* = `chưa phản hồi`, *Người chấp nhận* để trống |

Điểm đã được chấp nhận thì vòng sau không cảnh báo lại, trừ khi nội dung đổi làm rủi ro sống lại.

## Mục "Sai khác so với chuẩn"

```markdown
## Sai khác so với chuẩn

| Mục chuẩn | Mức | Tình trạng | Người chấp nhận | Thời điểm | Lý do |
|---|---|---|---|---|---|
| Yêu cầu tuân thủ & kiểm soát | 🔴 | Bỏ qua | <tên, vai> | <YYYY-MM-DD> | Nghiệp vụ chỉ đọc, không chạm tiền |
```

Không tự điền tên ai vào cột *Người chấp nhận*.

## Bốn điều cấm

1. **Không chặn** — không từ chối nộp vì chưa đủ chuẩn.
2. **Không im lặng bỏ qua** — thiếu mục thì in cảnh báo.
3. **Không lấp cho đủ** — không viết nội dung bịa vào mục bắt buộc; thà để trống và cảnh báo.
4. **Không nhắc lại** thứ người dùng đã chấp nhận bỏ.
