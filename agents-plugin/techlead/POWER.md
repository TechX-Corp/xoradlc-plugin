---
name: xora-techlead
displayName: Xora Tech lead
description: Tech lead thiết kế giải pháp và lập plan.
keywords: ["techlead", "tech lead", "architect", "thiết kế", "design", "plan", "kiến trúc"]
mcpServers: ["xora-techlead"]
---

# Xora Tech lead

Tech lead thiết kế giải pháp và lập plan. Nối với agent AAP qua MCP server `xora-techlead`.

## Khi kích hoạt

1. Đọc và làm theo `AGENTS.md` trong thư mục power — quy tắc của vai.
2. Chọn skill theo việc:
   - `design` → `skills/design/SKILL.md` — Use when the user asks for a solution or architecture design based on an approved BRD.
   - `plan` → `skills/plan/SKILL.md` — Use when the user asks to break a design into an implementation plan, tasks or milestones.

## MCP server `xora-techlead`

Chạy cục bộ (stdio, Python 3.9+, không cần thư viện ngoài) từ `~/.agent-plugins/xora-techlead/mcp/xora_mcp.py`, đọc `XORA_BASE_URL`, `XORA_API_KEY`, `XORA_USER_ID`, `XORA_AGENT` từ `~/.xora/env`. Không bao giờ in hay ghi API key ra output.
