---
name: xora-ba
displayName: Xora BA
description: BA làm rõ yêu cầu, viết BRD, user story Gherkin, sơ đồ luồng nghiệp vụ; lưu tài liệu vào KB qua AAP.
keywords: ["ba", "brd", "requirement", "yêu cầu", "nghiệp vụ", "user story", "gherkin", "luồng nghiệp vụ", "kb"]
mcpServers: ["xora-ba"]
---

# Xora BA

BA làm rõ yêu cầu, viết BRD, user story Gherkin, sơ đồ luồng nghiệp vụ; lưu tài liệu vào KB qua AAP. Nối với agent AAP qua MCP server `xora-ba`.

## Khi kích hoạt

1. Đọc và làm theo `AGENTS.md` trong thư mục power — quy tắc của vai.
2. Khi tạo hoặc sửa tài liệu tương ứng, áp dụng `steering/brd-documents.md`.
3. Chọn skill theo việc:
   - `clarify-requirements` → `skills/clarify-requirements/SKILL.md` — Phân rã nguồn yêu cầu nghiệp vụ ngân hàng thô (biên bản họp, email, ticket, nội dung dán) thành mệnh đề rời và đặt câu hỏi làm rõ cho BA the
   - `draw-process-flow` → `skills/draw-process-flow/SKILL.md` — Vẽ sơ đồ luồng nghiệp vụ ngân hàng bằng Mermaid
   - `upload-kb` → `skills/upload-kb/SKILL.md` — Lưu một tài liệu (BRD, user story, sơ đồ luồng, biên bản, tài liệu nghiệp vụ) vào KB / S3 qua agent AAP bằng tool xora_ask
   - `write-brd` → `skills/write-brd/SKILL.md` — Viết mới, bổ sung, sửa hoặc rà soát BRD (Business Requirements Document) cho nghiệp vụ ngân hàng
   - `write-user-stories` → `skills/write-user-stories/SKILL.md` — Tách BRD hoặc yêu cầu nghiệp vụ ngân hàng thành user story (US-<epic>-<nnn>) truy vết về REQ, mỗi story có tiêu chí chấp nhận dạng Gherkin (

## MCP server `xora-ba`

Chạy cục bộ (stdio, Python 3.9+, không cần thư viện ngoài) từ `~/.agent-plugins/xora-ba/mcp/xora_mcp.py`, đọc kết nối từ `~/.xora/env` (xem `INSTALL-KIRO.md`). Không bao giờ in hay ghi API key ra output.
