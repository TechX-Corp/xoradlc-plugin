---
name: implement
description: Hiện thực code, cấu hình, migration và test cho một yêu cầu ngân hàng theo BRD và thiết kế đã duyệt — truy vết REQ, dùng lại helper và convention đang có, idempotency và audit log cho thao tác chạm tiền/PII, test luồng sai trước. Dùng khi Developer nhờ viết, sửa, hoặc refactor code.
---

# Hiện thực thay đổi

## Bước 1 — Trước khi gõ dòng nào

1. Liệt kê các `REQ-xxx` mà thay đổi này phục vụ, và thiết kế version nào. Không gắn được vào REQ nào → hỏi Developer: việc ngoài phạm vi, hay BRD thiếu REQ?
2. Đọc code hiện có ở vùng sắp sửa. Tìm sẵn helper, util, mẫu xử lý lỗi, cách đặt tên đang dùng — **dùng lại cái đang có** thay vì viết mới.
3. Thiết kế không khớp code thực tế hoặc thiếu chỗ cần → dừng, nêu với Developer, đề xuất sửa thiết kế. Không tự đi đường khác.

Cần convention hay tiền lệ nội bộ mà repo không có → hỏi agent AAP bằng `xora_ask` (xem AGENTS.md).

## Bước 2 — Trong lúc viết

| Luật | Áp dụng |
|---|---|
| Theo convention | Convention dự án; không có thì theo mẫu phổ biến nhất trong chính repo |
| Chạm tiền / hồ sơ khách hàng | Có idempotency key, có audit log (ai, khi nào, cái gì — PII đã che hoặc băm), xử lý được khi bị gọi lại hay thất bại giữa chừng |
| Lỗi | Trả theo mã lỗi chuẩn của dự án. Không nuốt exception. Không log dữ liệu nhạy cảm |
| Cấu hình | Tham số mới khai báo theo từng môi trường; secret lấy từ biến môi trường / kho secret, không hard-code |
| Migration | Mỗi migration có script lùi; chỉ ra điểm không lùi được |
| Test | Viết test cho nhánh sai **trước** nhánh thuận; dữ liệu test là dữ liệu bịa |
| Làm tắt cố ý | Comment nêu giới hạn và hướng nâng cấp, rồi đưa vào mục *Nợ kỹ thuật* của changeset |
| Phạm vi | Chỉ đụng phần cần cho REQ. Không refactor lan sang vùng khác khi không ai yêu cầu |

## Bước 3 — Kiểm tại chỗ

Chạy những gì dự án có: build, lint, **sensor convention** (`sensor/convention_sensor.py` của plugin — hook thường đã chạy, chạy lại một lần trên toàn bộ file đổi trước khi bàn giao), unit test, độ phủ, quét secret / phụ thuộc. Báo đúng kết quả thật:

- Chạy được → nêu lệnh đã chạy và kết quả (kể cả đỏ).
- Không chạy được (thiếu môi trường, thiếu quyền) → ghi *chưa chạy*, kèm lý do. **Không bao giờ báo xanh khi chưa chạy.**

## Bước 4 — Bàn giao cho Developer

Báo đúng 4 dòng, rồi đề nghị lập hồ sơ bằng skill `changeset`:

1. Phục vụ REQ nào.
2. Kiểm tại chỗ qua hay không qua, cái nào chưa chạy được.
3. Rủi ro lớn nhất của lần thay đổi này.
4. Phần nào Developer nên đọc kỹ trước khi commit.

Không push, không merge, không tạo PR vào nhánh chính khi Developer chưa bảo.
