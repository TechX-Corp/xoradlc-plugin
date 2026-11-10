---
name: write-user-stories
description: Tách BRD hoặc yêu cầu nghiệp vụ ngân hàng thành user story (US-<epic>-<nnn>) truy vết về REQ, mỗi story có tiêu chí chấp nhận dạng Gherkin (Given / When / Then) gồm cả luồng sai và giá trị biên, kèm ma trận REQ ↔ US. Dùng khi dự án chạy Agile / Scrum, khi cần đưa yêu cầu vào backlog (Jira…), hoặc khi QA cần kịch bản kiểm thử từ yêu cầu.
---

# User story và tiêu chí chấp nhận Gherkin

Nguồn tốt nhất là **BRD đã có `REQ-xxx`**. Chưa có BRD mà nguồn còn mơ hồ → làm rõ trước (skill `clarify-requirements`), hoặc người dùng bảo làm luôn thì chỗ thiếu ghi `[GIẢ ĐỊNH]`.

## Bước 1 — Tách story

Mỗi story:

```
### US-042-003 — Chặn hồ sơ trùng CCCD
REQ: REQ-042-003 | Ưu tiên: Bắt buộc | Ước lượng: để trống (đội phát triển điền)

Là nhân viên kinh doanh,
tôi muốn hệ thống chặn việc tạo hồ sơ trùng số CCCD đang hoạt động,
để không mở hai hồ sơ cho cùng một khách hàng.
```

| Luật | Kiểm thế nào |
|---|---|
| Truy vết | Mỗi story chỉ về ít nhất một `REQ-xxx`. Không chỉ được → `[CẦN XÁC NHẬN]`, không tự bịa REQ |
| Đúng vai | *Là …* là vai nghiệp vụ thật trong BRD (NVKD, kiểm soát viên, khách hàng), không phải "người dùng" chung chung hay "hệ thống" |
| Có giá trị | *để …* nêu kết quả nghiệp vụ, không lặp lại hành động |
| Đủ nhỏ (INVEST) | Làm xong trong một sprint. Quá lớn → tách theo: bước trong luồng · quy tắc nghiệp vụ · biến thể dữ liệu · luồng đúng / luồng sai |
| Không quyết giải pháp | Không chọn màn hình, nút, API, bảng dữ liệu khi nguồn không nói |
| Phi chức năng | Hiệu năng, bảo mật, audit **không** viết thành story riêng; đưa vào AC của story liên quan hoặc giữ ở REQ phi chức năng và dẫn chiếu |

Maker–checker, phê duyệt theo cấp, đảo giao dịch, đối soát → luôn là story riêng, không gộp vào story luồng chính.

## Bước 2 — Tiêu chí chấp nhận Gherkin

Từ khoá `Feature / Scenario / Given / When / Then / And / But` giữ tiếng Anh để chạy được với Cucumber, Behave, SpecFlow, Jira Xray; nội dung viết tiếng Việt. Dự án đã dùng `# language: vi` thì theo dự án.

```gherkin
Feature: Chặn hồ sơ trùng CCCD (US-042-003 · REQ-042-003)

  Scenario: Từ chối khi CCCD đã thuộc hồ sơ đang hoạt động
    Given đã có hồ sơ "HS-0001" trạng thái "Active" với CCCD "001099000001"
    When nhân viên kinh doanh tạo hồ sơ mới với CCCD "001099000001"
    Then hệ thống từ chối lưu với mã lỗi "DUP_ID"
    And thông báo nêu mã hồ sơ đang giữ CCCD là "HS-0001"
    And ghi một dòng audit gồm người thao tác, thời điểm, CCCD đã băm

  Scenario: Cho lưu khi hồ sơ cũ đã đóng
    Given đã có hồ sơ "HS-0002" trạng thái "Closed" với CCCD "001099000002"
    When nhân viên kinh doanh tạo hồ sơ mới với CCCD "001099000002"
    Then hệ thống lưu hồ sơ mới

  Scenario Outline: CCCD sai định dạng
    When nhân viên kinh doanh tạo hồ sơ với CCCD "<cccd>"
    Then hệ thống từ chối lưu với mã lỗi "INVALID_ID"

    Examples:
      | cccd          | ghi chú         |
      | 00109900000   | thiếu 1 chữ số  |
      | 0010990000011 | thừa 1 chữ số   |
      | 00109900000A  | có chữ cái      |
```

| Luật | Kiểm thế nào |
|---|---|
| Có luồng sai | Mỗi story ít nhất một Scenario cho trường hợp sai: dữ liệu sai, trùng, vượt ngưỡng, hệ thống ngoài không phản hồi, bị từ chối duyệt |
| Có giá trị biên | Ngưỡng định lượng (hạn mức, số tiền, thời gian, cut-off) → `Scenario Outline` với giá trị ngay dưới, bằng, ngay trên ngưỡng |
| Kiểm được | `Then` là kết quả quan sát được: mã lỗi, trạng thái, số tiền, thông báo, dòng audit. Cấm "hệ thống xử lý đúng", "hiển thị phù hợp" |
| Một hành vi mỗi Scenario | Một `When` chính. Nhiều hành động → tách Scenario |
| Dữ liệu bịa | CCCD, số tài khoản, tên trong ví dụ là dữ liệu giả, không lấy từ nguồn thật |
| Khớp BRD | AC trong BRD nói gì thì Scenario nói đúng thế — cùng mã lỗi, cùng ngưỡng. Thấy BRD mâu thuẫn hoặc thiếu → nêu, không tự sửa BRD |

## Bước 3 — Ma trận truy vết

Cuối tài liệu:

```markdown
## Truy vết REQ ↔ US

| REQ | User story | Ghi chú |
|---|---|---|
| REQ-042-003 | US-042-003 | |
| REQ-042-005 | — | ⚠ chưa có story |
| — | US-042-009 | ⚠ [CẦN XÁC NHẬN] không gắn được REQ |
```

## Bước 4 — Tự kiểm và nộp

Chấm từng dòng, thiếu thì in khối cảnh báo **sau** tài liệu theo mẫu của skill `write-brd` (🔴 / 🟡 / ⚪, mỗi điểm 3 dòng *Chuẩn cần vì · Để trống thì · Điền được nếu*):

| Mục | Mức | Coi là thiếu khi |
|---|---|---|
| Truy vết | 🔴 | Có REQ chức năng không có story, hoặc story không gắn REQ |
| Luồng sai | 🔴 | Có story không có Scenario luồng sai |
| Giá trị biên | 🔴 | Có ngưỡng định lượng mà không có `Scenario Outline` biên |
| Chạm tiền / PII | 🔴 | Story chạm tiền hoặc PII mà AC không nói audit, maker–checker, che dữ liệu |
| Kích thước | 🟡 | Story có hơn 8 Scenario, hoặc *tôi muốn* chứa "và" nối hai việc |
| Vai | 🟡 | *Là người dùng* / *Là hệ thống* |

Báo người dùng 3 dòng: số story và số Scenario · REQ chưa có story · điểm 🔴 còn mở.

Ghi file theo quy ước dự án; mặc định `stories/US-<epic>-<slug>.md`, frontmatter như BRD (`id: <epic>/user-stories`, `inputs` trỏ về BRD và version đã dùng, `status: draft`). Một file `.feature` riêng chỉ tạo khi người dùng hoặc QA yêu cầu.
