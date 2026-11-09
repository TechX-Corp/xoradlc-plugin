---
name: xora-qa
displayName: Xora QA
description: QA lập test plan, test case và ghi kết quả.
keywords: ["qa", "test plan", "test case", "kiểm thử", "test result", "kết quả test"]
mcpServers: ["xora-qa"]
---

# Xora QA

QA lập test plan, test case và ghi kết quả. Nối với agent AAP qua MCP server `xora-qa`.

## Khi kích hoạt

1. Đọc và làm theo `AGENTS.md` trong thư mục power — quy tắc của vai.
2. Chọn skill theo việc:
   - `test-cases` → `skills/test-cases/SKILL.md` — Use when the user asks to write test cases from a BRD, design or test plan.
   - `test-plan` → `skills/test-plan/SKILL.md` — Use when the user asks for a test plan (scope, approach, environments, entry/exit criteria) for a feature.
   - `test-results` → `skills/test-results/SKILL.md` — Use when the user asks to record or summarize test execution results and defects.

## MCP server `xora-qa`

Chạy cục bộ (stdio, Python 3.9+, không cần thư viện ngoài) từ `~/.agent-plugins/xora-qa/mcp/xora_mcp.py`, đọc kết nối từ `~/.xora/env` (xem `INSTALL-KIRO.md`). Không bao giờ in hay ghi API key ra output.
