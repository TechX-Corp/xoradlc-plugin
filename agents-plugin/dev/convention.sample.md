---
version: 1
owner: <tech lead của dự án>
updated: <YYYY-MM-DD>
---

# Convention — <tên dự án>

Chuẩn viết code chung của dự án. Agent và người đều theo file này; sensor của plugin `xora-dev` kiểm phần **Luật kiểm tự động** mỗi khi code được sinh hoặc sửa, và **cảnh báo** — không chặn.

Sửa file này như sửa code: qua review, tăng `version`.

## 1. Ngôn ngữ và cấu trúc

- Ngôn ngữ / framework: <ví dụ Java 17 + Spring Boot 3>
- Cấu trúc thư mục: <ví dụ `src/main/java/<package>/{controller,service,repository,domain}`>
- Dùng lại helper sẵn có trước khi viết mới: <liệt kê helper / util chung, ví dụ `common/error/ErrorCodes`>

## 2. Đặt tên

| Thứ | Quy ước | Ví dụ |
|---|---|---|
| File / class Java | PascalCase, trùng tên class | `TransferService.java` |
| File Python | snake_case | `limit_checker.py` |
| Hằng số | UPPER_SNAKE_CASE | `MAX_DAILY_LIMIT` |
| Mã lỗi | `<MIỀN>_<LÝ_DO>`, khai báo tập trung | `LIMIT_EXCEEDED` |

## 3. Xử lý lỗi

- Không nuốt exception: không `catch` rỗng, không `except: pass`.
- Lỗi trả ra ngoài dùng mã lỗi chuẩn của dự án, không trả stack trace.
- Thao tác chạm tiền / hồ sơ khách hàng: có idempotency key và xử lý bị gọi lại.

## 4. Logging và dữ liệu nhạy cảm

- Dùng logger của dự án, không `System.out`, `console.log`, `print` trong code chạy thật.
- Không log CCCD, số tài khoản, số điện thoại, tên khách hàng ở dạng rõ — che hoặc băm.
- Audit log cho thao tác chạm tiền / PII: ai, khi nào, cái gì.

## 5. Bảo mật

- Không viết cứng mật khẩu, token, API key, private key. Lấy từ biến môi trường / kho secret.
- Không commit file `.env`, key, certificate.

## 6. Test

- Test luồng sai trước luồng thuận. Dữ liệu test là dữ liệu bịa.
- <ngưỡng độ phủ của dự án, ví dụ ≥ 80% cho service mới>

## 7. Comment và nợ kỹ thuật

- `TODO` / `FIXME` phải kèm mã ticket (`TODO ABC-123: …`), và có trong mục *Nợ kỹ thuật* của hồ sơ changeset.
- Bỏ qua một luật có chủ đích: `convention-ignore: <ID> <lý do>` trên chính dòng đó.

## 8. Luật kiểm tự động

Sensor đọc khối JSON dưới đây. Thêm luật: chọn `type` (`forbid` / `require` / `filename` / `max_lines`), viết `pattern` (regex Python), giới hạn `files`, và **thử trên code thật** để tránh báo nhầm.

```json
{
  "version": "1",
  "exclude": ["**/node_modules/**", "**/build/**", "**/dist/**", "**/target/**", "**/vendor/**", "**/.git/**", "**/*.min.js"],
  "rules": [
    {
      "id": "SEC-001", "type": "forbid",
      "files": ["**/*.java", "**/*.kt", "**/*.ts", "**/*.js", "**/*.py", "**/*.cs", "**/*.go", "**/*.yml", "**/*.yaml", "**/*.properties"],
      "pattern": "(?i)(password|passwd|pwd|secret|api[_-]?key|access[_-]?token)\\s*[:=]\\s*[\"'][^\"'\\s$]{6,}[\"']",
      "message": "Giá trị bí mật viết cứng.",
      "fix": "lấy từ biến môi trường / kho secret"
    },
    {
      "id": "SEC-002", "type": "forbid", "files": ["**/*"],
      "pattern": "-----BEGIN (RSA |EC |OPENSSH )?PRIVATE KEY-----",
      "message": "Private key nằm trong repo.",
      "fix": "xoá khỏi repo, thu hồi key, dùng kho secret"
    },
    {
      "id": "LOG-001", "type": "forbid",
      "files": ["**/*.java"], "exclude": ["**/test/**"],
      "pattern": "System\\.(out|err)\\.print",
      "message": "In ra console thay vì dùng logger.",
      "fix": "dùng logger của dự án"
    },
    {
      "id": "LOG-002", "type": "forbid",
      "files": ["**/*.ts", "**/*.js"], "exclude": ["**/*.test.*", "**/*.spec.*", "**/test/**"],
      "pattern": "console\\.(log|debug)\\(",
      "message": "console.log trong code chạy thật.",
      "fix": "dùng logger của dự án"
    },
    {
      "id": "LOG-003", "type": "forbid",
      "files": ["**/*.py"], "exclude": ["**/test_*.py", "**/*_test.py", "**/tests/**", "**/scripts/**"],
      "pattern": "^\\s*print\\(",
      "message": "print() trong code chạy thật.",
      "fix": "dùng logging"
    },
    {
      "id": "PII-001", "type": "forbid",
      "files": ["**/*.java", "**/*.kt", "**/*.ts", "**/*.js", "**/*.py", "**/*.cs", "**/*.go"],
      "pattern": "(?i)(log(ger)?\\.(info|debug|warn|error|trace)|logging\\.\\w+)\\(.*(?<!mask\\()(?<!hash\\()\\b(cccd|cmnd|national_?id|id_?number|account_?(no|number)|so_?tai_?khoan|phone|so_?dien_?thoai|customer_?name)\\b",
      "message": "Có thể đang log dữ liệu cá nhân ở dạng rõ.",
      "fix": "che hoặc băm trước khi log — giá trị đã qua mask(…) / hash(…) không bị báo"
    },
    {
      "id": "ERR-001", "type": "forbid",
      "files": ["**/*.java", "**/*.kt", "**/*.ts", "**/*.js", "**/*.cs"],
      "pattern": "catch\\s*(\\([^)]*\\))?\\s*\\{\\s*\\}",
      "message": "catch rỗng — exception bị nuốt.",
      "fix": "log và trả mã lỗi chuẩn, hoặc ném tiếp"
    },
    {
      "id": "ERR-002", "type": "forbid", "files": ["**/*.py"],
      "pattern": "except(\\s+\\w+(\\s+as\\s+\\w+)?)?\\s*:\\s*pass\\b",
      "message": "except: pass — exception bị nuốt.",
      "fix": "log và xử lý, hoặc raise tiếp"
    },
    {
      "id": "ERR-003", "type": "forbid", "files": ["**/*.java"],
      "pattern": "\\.printStackTrace\\(\\)",
      "message": "printStackTrace thay vì log.",
      "fix": "logger.error(\"…\", e)"
    },
    {
      "id": "TODO-001", "type": "forbid",
      "files": ["**/*.java", "**/*.kt", "**/*.ts", "**/*.js", "**/*.py", "**/*.cs", "**/*.go"],
      "pattern": "\\b(TODO|FIXME)\\b(?!.*\\b[A-Z][A-Z0-9]+-\\d+)",
      "message": "TODO/FIXME không kèm mã ticket.",
      "fix": "ghi `TODO ABC-123: …` và đưa vào mục Nợ kỹ thuật"
    },
    {
      "id": "NAME-001", "type": "filename", "files": ["**/*.java"],
      "pattern": "^[A-Z][A-Za-z0-9]*\\.java$",
      "message": "Tên file Java không theo PascalCase."
    },
    {
      "id": "NAME-002", "type": "filename", "files": ["**/*.py"],
      "pattern": "^[a-z_][a-z0-9_]*\\.py$",
      "message": "Tên file Python không theo snake_case."
    },
    {
      "id": "SIZE-001", "type": "max_lines", "max": 500,
      "files": ["**/*.java", "**/*.kt", "**/*.ts", "**/*.js", "**/*.py", "**/*.cs", "**/*.go"],
      "message": "File dài hơn 500 dòng.",
      "fix": "tách theo trách nhiệm"
    }
  ]
}
```
