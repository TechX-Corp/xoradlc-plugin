---
name: changeset
description: Lập, bổ sung hoặc rà soát hồ sơ changeset (mô tả PR / merge request) theo chuẩn ngân hàng — truy vết REQ, phạm vi thay đổi, phần do AI sinh theo từng file, cấu hình, migration và rollback, kết quả kiểm tại chỗ, quét bảo mật, giới hạn đã biết — rồi in khối cảnh báo lệch chuẩn. Dùng sau khi code xong hoặc khi cần rà một PR.
---

# Hồ sơ changeset

Đọc [references/changeset-standard.md](references/changeset-standard.md) trước khi viết hoặc rà: danh mục 13 mục chuẩn, mẫu khối cảnh báo, cách xử lý câu trả lời.

## Thu thập trước khi viết

Lấy từ thực tế, không từ trí nhớ:

- Danh sách file đổi: `git diff --stat` (hoặc tương đương) so với nhánh gốc.
- SHA commit thật, nếu đã commit. Chưa commit → để trống, ghi *chưa commit*.
- Kết quả build / lint / test / quét đã thực sự chạy trong phiên, kèm lệnh. Cái chưa chạy → ghi *chưa chạy*, lý do.
- Kết quả sensor convention trên các file đổi: số cảnh báo còn lại theo mã luật, và các dòng `convention-ignore` mới thêm (mỗi dòng phải có trong mục *Nợ kỹ thuật cố ý*).
- REQ và version thiết kế mà thay đổi phục vụ.

## Mục "Phần do AI sinh" — bắt buộc và phải thật

Người tiếp quản cần phân biệt phần nào do người viết, phần nào do AI. Ghi theo từng file hoặc từng vùng:

```markdown
## Phần do AI sinh

| File / vùng | Nguồn gốc | Developer đã đọc kỹ? | Ghi chú |
|---|---|---|---|
| src/payment/DuplicateGuard.java | AI sinh | Chưa | Logic khoá theo mã tham chiếu, cần đọc kỹ cửa sổ 24h |
| src/payment/TransferService.java#L40-L88 | AI sinh, người sửa nhiều | Rồi | |
| db/migration/V42__dup_guard.sql | người viết | — | |
```

Ba giá trị nguồn gốc: `AI sinh` / `AI sinh, người sửa nhiều` / `người viết`. Không biết Developer đã đọc chưa → hỏi, đừng tự điền "Rồi". Phần AI sinh mà chưa ai đọc kỹ là thông tin có ích nhất trong hồ sơ — đừng làm đẹp nó.

## Nộp

Tự kiểm theo danh mục chuẩn (5 bước trong reference), in khối cảnh báo **sau** hồ sơ nếu có điểm thiếu, rồi báo đúng 4 dòng:

1. Phục vụ REQ nào.
2. Kiểm tại chỗ qua hay không qua, cái nào chưa chạy được.
3. Rủi ro lớn nhất của lần thay đổi này.
4. Phần nào reviewer nên đọc kỹ trước khi duyệt.

Mục 3 (AI sinh), 5 (migration & rollback), 13 (giới hạn đã biết) là ba mục reviewer đọc đầu tiên — đừng để chúng chung chung.

## Rà soát PR / changeset có sẵn

Không viết lại. Chấm từng mục theo danh mục chuẩn, đối chiếu với diff thật (file trong diff mà không có trong *Phạm vi thay đổi*, secret lộ trong diff, migration không có script lùi), rồi in khối cảnh báo. Đề xuất sửa cụ thể; chỉ sửa khi người dùng bảo.
