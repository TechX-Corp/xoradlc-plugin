---
name: write-brd
description: Viết mới, bổ sung, sửa hoặc rà soát BRD (Business Requirements Document) cho nghiệp vụ ngân hàng — mỗi yêu cầu một ID REQ truy vết, tiêu chí chấp nhận định lượng, tự kiểm theo danh mục chuẩn và in khối cảnh báo lệch chuẩn. Dùng khi BA cần BRD, cần rà một BRD có sẵn, hoặc cập nhật BRD theo phản hồi.
---

# Viết và rà soát BRD

Đọc [references/brd-standard.md](references/brd-standard.md) trước khi viết hoặc rà: danh mục 13 mục chuẩn, mẫu khối cảnh báo, cách xử lý câu trả lời.

## Viết

Theo template dự án nếu có; không có thì theo danh mục chuẩn. Nguồn còn mơ hồ và BA chưa bảo làm luôn → dùng skill `clarify-requirements` trước.

Mỗi yêu cầu theo mẫu:

```
### REQ-042-003 — Chặn hồ sơ trùng số CCCD
Ưu tiên: Bắt buộc | Nguồn: yêu cầu nghiệp vụ §2 | Loại: Chức năng

Khi NVKD tạo hồ sơ với số CCCD đã tồn tại ở trạng thái Active,
hệ thống từ chối lưu và trả mã lỗi DUP_ID kèm mã hồ sơ đang giữ số đó.

Tiêu chí chấp nhận:
- AC1: CCCD trùng, hồ sơ cũ Active  → từ chối, trả DUP_ID, nêu mã hồ sơ cũ.
- AC2: CCCD trùng, hồ sơ cũ Closed  → cho phép lưu.
- AC3: Kiểm tra trùng xong trong ≤ 500ms ở p95.
- AC4: Mỗi lần bị chặn ghi một dòng audit log: người thao tác, thời điểm, CCCD đã băm.
```

| Luật | Kiểm thế nào |
|---|---|
| Cấm từ mơ hồ | Không chứa: nhanh, dễ dùng, thân thiện, ổn định, hợp lý, đầy đủ, linh hoạt, tối ưu. Gặp trong nguồn → đó là câu hỏi, không phải chữ để chép |
| Một yêu cầu một việc | "và" nối hai hành vi khác nhau → tách đôi |
| Có AC cho luồng sai | Mỗi REQ có ít nhất một AC cho trường hợp sai |
| Truy ngược được | Mỗi REQ chỉ về một câu trong nguồn hoặc câu trả lời của BA. Không truy được → bỏ, hoặc `[CẦN XÁC NHẬN]` |
| Giữ ID | Không đánh lại số đã cấp. Yêu cầu bị bỏ → `Đã huỷ ở v<n>` |

Mở đầu tài liệu bằng frontmatter (`id`, `version`, `status: draft`, `produced_by`, `inputs`, `approved_by: null`, `approved_at: null`). Ghi file thì theo quy ước đặt tên của dự án; mặc định `brd/BRD-<epic>-<slug>.md`.

## Nộp

Tự kiểm theo danh mục chuẩn (5 bước trong reference), in khối cảnh báo **sau** tài liệu nếu có điểm thiếu, rồi báo BA đúng 4 dòng:

1. Đã viết bao nhiêu `REQ-xxx`.
2. Còn bao nhiêu `[CẦN XÁC NHẬN]`, ở mục nào.
3. Điểm tiền / PII / tuân thủ chưa chốt — `[CẦN TUÂN THỦ XÁC NHẬN]` nêu riêng.
4. Rủi ro nếu duyệt luôn: *duyệt bây giờ là chấp nhận rủi ro ở REQ-xxx, REQ-yyy.*

Không viết "BRD đã hoàn chỉnh" khi bảng *Câu hỏi còn mở* chưa rỗng.

## Rà soát BRD có sẵn

Không viết lại. Chấm từng mục theo danh mục chuẩn, chấm từng REQ theo bảng luật ở trên, rồi in khối cảnh báo. Đề xuất sửa cụ thể từng chỗ; chỉ sửa file khi BA bảo.

## Khi BA phản hồi hoặc sửa tay

Tăng `version` · `status: draft` · ghi *Nhật ký thay đổi* (đổi gì, theo phản hồi nào, REQ nào bị ảnh hưởng) · giữ nguyên chữ BA đã sửa, chỉ chỉnh chỗ xung đột và nói rõ đã đụng vào đâu.
