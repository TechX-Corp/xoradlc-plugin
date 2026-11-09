# xoradlc-plugin

Repo chứa hai nhóm plugin độc lập, mỗi nhóm một thư mục:

| Thư mục | Là gì | Cần gì |
|---|---|---|
| [`aidlc-plugin/`](aidlc-plugin/README.md) | Plugin AI-DLC `techx-dlc`: 2 scope (`techx-dlc-onboard`, `techx-dlc-feature`) + 23 overlay | `aidlc` 2.11.2+; dùng `--plugin-root aidlc-plugin` |
| [`agents-plugin/`](agents-plugin/README.md) | 6 plugin theo vai (BA, Tech lead, Dev, QA, Reviewer, Delivery) cho Claude Code và Kiro, nối với agent AAP qua MCP stdio | Chỉ cần `python3`; không cần `aidlc` |

Hai nhóm không dùng chung file nào; sửa nhóm này không ảnh hưởng nhóm kia.
