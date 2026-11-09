---
inclusion: fileMatch
fileMatchPattern: "**/changesets/**/*.md"
---

# Quy ước file hồ sơ changeset

Áp dụng khi tạo hoặc sửa hồ sơ changeset trong workspace. Dự án dùng mô tả PR thay cho file riêng thì vẫn theo nội dung này.

## Đặt tên và vị trí

- Mặc định: `changesets/CS-<epic>-<nnn>-<slug>.md` (ví dụ `changesets/CS-042-001-chong-trung-giao-dich.md`). Dự án có quy ước riêng thì theo dự án.

## Frontmatter bắt buộc

```yaml
---
id: <epic-id>/changeset-<nnn>
version: <tăng 1 mỗi lần sửa sau khi đã nộp>
status: draft | awaiting_review | approved | rejected
produced_by: xora-dev
inputs:
  - path: <BRD / thiết kế>
    version: <version đã dùng>
commits: [<SHA thật, để trống nếu chưa commit>]
approved_by: null
approved_at: null
---
```

## Khi sửa

- Tăng `version`, đặt `status: draft`, ghi *Nhật ký thay đổi*: đổi gì, theo phản hồi nào.
- File đang `status: approved` → không sửa đè; tạo version mới.
- Chỉ người duyệt được đặt `approved_by`, `approved_at`, `status: approved`. Không bịa SHA hay kết quả chạy.
