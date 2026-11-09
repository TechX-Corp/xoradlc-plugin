---
name: gather-evidence
description: Use when the user asks to collect the evidence (BRD, design, changeset, test results) needed to review a change.
---

Gọi tool `xora_get_skill` của MCP server `xora-reviewer` với `name: gather-evidence` rồi làm theo đúng nội dung trả về. Nếu tool báo lỗi, báo lại lỗi cho người dùng và dừng; không tự đoán quy trình.
