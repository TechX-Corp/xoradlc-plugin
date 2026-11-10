---
name: draw-process-flow
description: Vẽ sơ đồ luồng nghiệp vụ ngân hàng bằng Mermaid — flowchart chia làn theo vai / hệ thống, có đủ nhánh ngoại lệ (từ chối, timeout hệ thống ngoài, trùng, đảo giao dịch, đường thủ công), mỗi bước gắn REQ; hoặc sequence diagram khi tích hợp nhiều hệ thống. Dùng cho mục "Luồng nghiệp vụ & ngoại lệ" của BRD, hoặc khi BA cần sơ đồ as-is / to-be để trao đổi với nghiệp vụ.
---

# Sơ đồ luồng nghiệp vụ (Mermaid)

Mermaid là văn bản, nằm ngay trong file `.md`: Claude Code, Kiro, GitHub, GitLab, Confluence (có app Mermaid) đều hiển thị được, và sửa được bằng diff như phần còn lại của BRD.

## Chọn loại sơ đồ

| Cần thể hiện | Dùng |
|---|---|
| Ai làm bước nào, nhánh quyết định, ngoại lệ | `flowchart TD` chia làn bằng `subgraph` theo vai / hệ thống |
| Thứ tự gọi giữa các hệ thống, timeout, phản hồi | `sequenceDiagram` |
| Vòng đời trạng thái hồ sơ / giao dịch | `stateDiagram-v2` |

Luồng dài → tách thành sơ đồ tổng quan (5–10 bước) và sơ đồ chi tiết cho từng bước phức tạp, thay vì một sơ đồ 40 nút.

## Mẫu flowchart

```mermaid
flowchart TD
  subgraph KH["Khách hàng"]
    A(["Bắt đầu: gửi lệnh chuyển khoản liên ngân hàng"])
  end
  subgraph NH["Ngân hàng số"]
    B["Kiểm tra hạn mức ngày<br/>REQ-042-002"]
    C{"Vượt hạn mức?"}
    D["Hạch toán tạm giữ tiền<br/>REQ-042-003"]
    E["Gửi lệnh sang NAPAS<br/>REQ-042-004"]
    F{"NAPAS phản hồi<br/>trong 30 giây?"}
    G{"Kết quả NAPAS"}
    H["Ghi nhận thành công"]
    I["Hoàn tiền tạm giữ<br/>REQ-042-006"]
    J["Đánh dấu chờ đối soát<br/>REQ-042-007"]
    Z1(["Kết thúc: thành công"])
    Z2(["Kết thúc: từ chối, mã LIMIT_EXCEEDED"])
    Z3(["Kết thúc: thất bại, đã hoàn tiền"])
  end
  subgraph VH["Vận hành"]
    K["Đối soát thủ công cuối ngày"]
    Q["[CẦN XÁC NHẬN] Ai duyệt hoàn tiền sau đối soát?"]
  end
  A --> B --> C
  C -- "Có" --> Z2
  C -- "Không" --> D --> E --> F
  F -- "Có" --> G
  F -- "Không, timeout" --> J --> K --> Q
  G -- "Thành công" --> H --> Z1
  G -- "Từ chối" --> I --> Z3
  classDef pending stroke-dasharray: 5 5
  class Q pending
```

## Mẫu sequence diagram

```mermaid
sequenceDiagram
  autonumber
  actor KH as Khách hàng
  participant NH as Ngân hàng số
  participant CORE as Core banking
  participant NP as NAPAS
  KH->>NH: Gửi lệnh chuyển khoản
  NH->>CORE: Tạm giữ số tiền (REQ-042-003)
  CORE-->>NH: Đã tạm giữ
  NH->>NP: Gửi lệnh (REQ-042-004)
  alt NAPAS phản hồi trong 30 giây
    NP-->>NH: Thành công / Từ chối
  else Timeout
    NH->>NH: Đánh dấu chờ đối soát (REQ-042-007)
  end
```

## Luật vẽ

| Luật | Kiểm thế nào |
|---|---|
| Không bịa bước | Mỗi bước có trong nguồn hoặc BRD. Bước suy ra mà chưa ai xác nhận → nút có nhãn `[CẦN XÁC NHẬN]` và class `pending` (viền nét đứt) |
| Gắn REQ | Bước ứng với yêu cầu thì ghi `REQ-xxx` trong nhãn (`<br/>` để xuống dòng) |
| Quyết định đủ nhánh | Mọi nút `{…}` có nhãn trên **mọi** nhánh ra, và các nhánh phủ hết trường hợp (Có / Không, Thành công / Từ chối / Timeout) |
| Mọi đường đều kết thúc | Mỗi nhánh đi tới một nút kết thúc `(["Kết thúc: …"])` nêu rõ kết quả, hoặc quay về một bước đã có. Không để nhánh treo |
| Hệ thống ngoài có nhánh lỗi | Mỗi lần gọi hệ thống ngoài / core / đối tác có nhánh timeout hoặc lỗi |
| Chạm tiền | Có nhánh hoàn tiền / đảo giao dịch và điểm đối soát |
| Phê duyệt | Bước duyệt (maker–checker) có nhánh *Từ chối* và nói đi đâu sau khi từ chối |
| Không quyết giải pháp | Tên hệ thống chỉ dùng khi nguồn đã nói (NAPAS, core). Không thêm queue, database, microservice |
| Cú pháp an toàn | Nhãn luôn đặt trong ngoặc kép `"…"`; không dùng ký tự `"` bên trong nhãn; id nút chỉ chữ / số không dấu |

## Tự kiểm và nộp

Trước khi nộp, đi lại từng nhánh từ nút *Bắt đầu* và chấm:

- [ ] Mọi nút quyết định có đủ nhánh và nhãn nhánh.
- [ ] Mọi nhánh kết thúc ở nút *Kết thúc* có kết quả rõ.
- [ ] Mỗi lần gọi hệ thống ngoài có nhánh lỗi / timeout.
- [ ] Luồng chạm tiền có hoàn tiền / đảo và đối soát.
- [ ] Bước duyệt có nhánh từ chối.
- [ ] Không có bước nào không truy được về nguồn mà thiếu nhãn `[CẦN XÁC NHẬN]`.

Ô nào chưa đạt → nêu trong khối cảnh báo sau sơ đồ (mẫu của skill `write-brd`), kèm câu hỏi để điền.

Báo người dùng 3 dòng: số bước và số nhánh ngoại lệ · số nút `[CẦN XÁC NHẬN]` · REQ trong BRD chưa xuất hiện trên sơ đồ (nếu vẽ từ BRD).

Vị trí: sơ đồ cho BRD đặt trong mục *Luồng nghiệp vụ & ngoại lệ* của chính BRD; sơ đồ độc lập (as-is / to-be) ghi `flows/<epic>-<slug>.md` hoặc theo quy ước dự án. Người dùng chỉ nhờ vẽ thì trả sơ đồ, không tự chèn vào BRD.
