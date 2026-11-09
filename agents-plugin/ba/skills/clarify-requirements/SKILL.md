---
name: clarify-requirements
description: Phân rã nguồn yêu cầu nghiệp vụ ngân hàng thô (biên bản họp, email, ticket, nội dung dán) thành mệnh đề rời và đặt câu hỏi làm rõ cho BA theo mẫu "hỏi kèm đề xuất". Dùng khi yêu cầu còn mơ hồ, thiếu ngưỡng định lượng, hoặc chạm tiền / PII / tín dụng / tuân thủ.
---

# Làm rõ yêu cầu

## Bước 1 — Phân rã

Tách nguồn thành mệnh đề rời. Mỗi mệnh đề trả lời: *ai · làm gì · khi nào · kết quả mong đợi · sai thì sao*. Nguồn có nhiều mảnh → liệt kê ra và hỏi BA đã đủ chưa.

Gắn mỗi mệnh đề vào một nhóm; nhóm quyết định câu phải hỏi kỹ:

| Nhóm | Dấu hiệu | Phải hỏi kỹ |
|---|---|---|
| **Tiền** | Hạch toán, chuyển khoản, phí, lãi, tỷ giá, hoàn tiền | Chống trùng, thất bại giữa chừng, đảo giao dịch, làm tròn, cut-off, đối soát |
| **Dữ liệu khách hàng** | CIF, CCCD, số tài khoản, liên hệ, ảnh eKYC | Phân loại PII (NĐ 13/2023), che hiển thị, audit truy cập, thời hạn lưu, chia sẻ ra ngoài |
| **Tín dụng / rủi ro** | Phê duyệt, hạn mức, xếp hạng, từ chối hồ sơ | Giải thích được quyết định, hạn mức theo cấp |
| **Tuân thủ** | AML/CFT, cấm vận, báo cáo NHNN, FATCA/CRS | Nghĩa vụ báo cáo, hạn nộp, hit thì dừng hay cảnh báo |
| **Nội bộ** | Báo cáo, tra cứu, màn hình quản trị | Phân quyền |

Với mọi nhóm, soi thêm: ngưỡng định lượng · luồng sai (dữ liệu sai, hệ thống ngoài timeout, trùng lặp) · phạm vi **không** làm · tác động lên nghiệp vụ đang chạy · maker–checker · đường thủ công khi hệ thống lõi hoặc đối tác chết.

Cần tiền lệ hoặc quy định nội bộ để đề xuất → hỏi agent AAP bằng `xora_ask` (xem AGENTS.md), dùng trích dẫn `[n]` làm *Căn cứ*.

## Bước 2 — Hỏi kèm đề xuất

Chỉ hỏi câu mà **câu trả lời khác nhau dẫn tới BRD khác nhau**. Tối đa 7 câu một vòng, xếp theo mức ảnh hưởng; số câu theo mức rủi ro — màn hình tra cứu nội bộ không cần đủ bảy câu.

Mỗi câu theo đúng mẫu, để BA chỉ cần gật, sửa, hoặc bác:

```
**H3. Giao dịch bị gửi hai lần trong 30 giây thì xử lý thế nào?**
- Vì sao cần biết: quyết định có cần khoá chống trùng — ảnh hưởng REQ-042-005, REQ-042-007.
- Đề xuất: chặn lần hai, trả mã DUP_TXN, khoá theo mã tham chiếu giao dịch, cửa sổ 24 giờ.
- Căn cứ: epic 038 §4 đã dùng cách này cho chuyển khoản nội bộ.
- Nếu chọn khác: cho qua cả hai thì phải thêm nghiệp vụ đảo giao dịch và đối soát thủ công.
- Không trả lời thì: ghi [GIẢ ĐỊNH] theo đề xuất trên.
```

- Đề xuất là **một giá trị cụ thể** (`≤ 500ms ở p95`), không phải danh sách để BA tự chọn.
- Căn cứ nêu nguồn thật. Không có nguồn → ghi *suy luận của agent, chưa có tiền lệ*.
- Câu hỏi pháp lý → không đề xuất, đánh `[CẦN TUÂN THỦ XÁC NHẬN]`, chỉ nêu điều nghi và vì sao.
- BA im lặng hoặc "tuỳ bạn" → vào BRD dưới dạng `[GIẢ ĐỊNH]`. Chỉ gỡ nhãn khi BA xác nhận bằng chữ rõ ràng.

Mặc định dừng chờ BA trả lời trước khi viết BRD. Bỏ qua bước hỏi khi BA bảo làm luôn — câu hỏi chuyển xuống khối cảnh báo sau bản thảo (xem skill `write-brd`).
