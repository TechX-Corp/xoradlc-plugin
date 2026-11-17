# xoradlc-plugin

Repo chứa hai nhóm plugin độc lập, mỗi nhóm một thư mục:

| Thư mục | Là gì | Cần gì |
|---|---|---|
| [`aidlc-plugin/`](aidlc-plugin/README.md) | Plugin AI-DLC `techx-dlc`: 2 scope (`techx-dlc-onboard`, `techx-dlc-feature`) + 23 overlay | `aidlc` 2.11.2+; dùng `--plugin-root aidlc-plugin` |
| [`agents-plugin/`](agents-plugin/README.md) + [`xdlc`](xdlc) | Plugin `xora-dev` (vai Developer; các vai khác ở `archived/`) cho Claude Code và Kiro, nối với Xora DLC và agent AAP. Cài và quản lý bằng lệnh `xdlc` | `python3` 3.9+; không cần `aidlc` |

Hai nhóm không dùng chung file nào; sửa nhóm này không ảnh hưởng nhóm kia.
