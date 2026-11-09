---
name: decision
description: Use when the user asks to record an approve/reject decision with rationale for a reviewed artifact.
---

Gọi tool `xora_get_skill` của MCP server `xora-reviewer` với `name: decision` rồi làm theo đúng nội dung trả về. Nếu tool báo lỗi, báo lại lỗi cho người dùng và dừng; không tự đoán quy trình.
