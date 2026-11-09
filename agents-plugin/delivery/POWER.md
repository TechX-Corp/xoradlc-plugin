---
name: xora-delivery
displayName: Xora Delivery, DevOps
description: Delivery lập hồ sơ bàn giao DEV / UAT.
keywords: ["delivery", "devops", "bàn giao", "handover", "UAT", "release"]
mcpServers: ["xora-delivery"]
---

# Xora Delivery, DevOps

Delivery lập hồ sơ bàn giao DEV / UAT. Nối với agent AAP qua MCP server `xora-delivery`.

## Khi kích hoạt

1. Đọc và làm theo `AGENTS.md` trong thư mục power — quy tắc của vai.
2. Chọn skill theo việc:
   - `handover` → `skills/handover/SKILL.md` — Use when the user asks to prepare a handover package for deploying to DEV or UAT.

## MCP server `xora-delivery`

Chạy cục bộ (stdio, Python 3.9+, không cần thư viện ngoài) từ `~/.agent-plugins/xora-delivery/mcp/xora_mcp.py`, đọc kết nối từ `~/.xora/env` (xem `INSTALL-KIRO.md`). Không bao giờ in hay ghi API key ra output.
