---
name: handover
description: Use when the user asks to prepare a handover package for deploying to DEV or UAT.
---

Gọi tool `xora_get_skill` của MCP server `xora-delivery` với `name: handover` rồi làm theo đúng nội dung trả về. Nếu tool báo lỗi, báo lại lỗi cho người dùng và dừng; không tự đoán quy trình.
