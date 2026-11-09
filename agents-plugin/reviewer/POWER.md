---
name: xora-reviewer
displayName: Xora Reviewer, Owner
description: Reviewer gom bằng chứng và ra quyết định.
keywords: ["reviewer", "review", "phê duyệt", "bằng chứng", "evidence", "decision"]
mcpServers: ["xora-reviewer"]
---

# Xora Reviewer, Owner

Reviewer gom bằng chứng và ra quyết định. Nối với agent AAP qua MCP server `xora-reviewer`.

## Khi kích hoạt

1. Đọc và làm theo `AGENTS.md` trong thư mục power — quy tắc của vai.
2. Chọn skill theo việc:
   - `decision` → `skills/decision/SKILL.md` — Use when the user asks to record an approve/reject decision with rationale for a reviewed artifact.
   - `gather-evidence` → `skills/gather-evidence/SKILL.md` — Use when the user asks to collect the evidence (BRD, design, changeset, test results) needed to review a change.

## MCP server `xora-reviewer`

Chạy cục bộ (stdio, Python 3.9+, không cần thư viện ngoài) từ `~/.agent-plugins/xora-reviewer/mcp/xora_mcp.py`, đọc kết nối từ `~/.xora/env` (xem `INSTALL-KIRO.md`). Không bao giờ in hay ghi API key ra output.
