---
name: xora-dev
displayName: Xora Developer
description: Developer hiện thực theo thiết kế đã duyệt, lập hồ sơ changeset; sensor kiểm convention khi sinh code.
keywords: ["developer", "dev", "code", "changeset", "pull request", "convention", "migration", "kb"]
mcpServers: ["xora-dev"]
---

# Xora Developer

Developer hiện thực theo thiết kế đã duyệt, lập hồ sơ changeset; sensor kiểm convention khi sinh code. Nối với agent AAP qua MCP server `xora-dev`.

## Khi kích hoạt

1. Đọc và làm theo `AGENTS.md` trong thư mục power — quy tắc của vai.
2. Khi tạo hoặc sửa tài liệu tương ứng, áp dụng `steering/changeset-documents.md`.
3. Chọn skill theo việc:
   - `changeset` → `skills/changeset/SKILL.md` — Lập, bổ sung hoặc rà soát hồ sơ changeset (mô tả PR / merge request) theo chuẩn ngân hàng
   - `implement` → `skills/implement/SKILL.md` — Hiện thực code, cấu hình, migration và test cho một yêu cầu ngân hàng theo BRD và thiết kế đã duyệt
   - `upload-tech-doc` → `skills/upload-tech-doc/SKILL.md` — Lưu một tài liệu (hồ sơ changeset, ghi chú thiết kế, ADR, runbook, convention) vào KB / S3 qua agent AAP bằng tool xora_ask

## MCP server `xora-dev`

Chạy cục bộ (stdio, Python 3.9+, không cần thư viện ngoài) từ `~/.agent-plugins/xora-dev/mcp/xora_mcp.py`, đọc kết nối từ `~/.xora/env` (xem `INSTALL-KIRO.md`). Không bao giờ in hay ghi API key ra output.
