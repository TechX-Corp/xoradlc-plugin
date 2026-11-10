---
inclusion: fileMatch
fileMatchPattern: "**/brd/**/*.md"
---

# Quy ước file BRD

Áp dụng khi tạo hoặc sửa file BRD trong workspace.

## Đặt tên và vị trí

- Mặc định: `brd/BRD-<epic>-<slug>.md` (ví dụ `brd/BRD-042-chuyen-khoan-lien-ngan-hang.md`). Dự án có quy ước riêng thì theo dự án.
- ID yêu cầu: `REQ-<epic>-<nnn>`, cấp tăng dần, không đánh lại số đã cấp.

## Frontmatter bắt buộc

```yaml
---
id: <epic-id>/brd
version: <tăng 1 mỗi lần sửa sau khi đã nộp>
status: draft | awaiting_review | approved | rejected
produced_by: xora-ba
inputs:
  - path: <nguồn>
    version: <version đã dùng>
approved_by: null
approved_at: null
---
```

## Khi sửa

- Tăng `version`, đặt `status: draft`, ghi *Nhật ký thay đổi*: đổi gì, theo phản hồi nào, REQ nào bị ảnh hưởng.
- File đang `status: approved` → không sửa đè; tạo version mới và ghi rõ bản đã duyệt bị thay thế.
- Chỉ người duyệt được đặt `approved_by`, `approved_at`, `status: approved`.
