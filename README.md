# xoradlc-plugin

This repo holds two independent plugin groups, each in its own directory:

| Directory | What it is | Requirements |
|---|---|---|
| [`aidlc-plugin/`](aidlc-plugin/README.md) | AI-DLC plugin `techx-dlc`: 2 scopes (`techx-dlc-onboard`, `techx-dlc-feature`) + 23 overlays | `aidlc` 2.11.2+; use `--plugin-root aidlc-plugin` |
| [`agents-plugin/`](agents-plugin/README.md) + [`xdlc`](xdlc) | `xora-dev` plugin (Developer role; other roles under `archived/`) for Claude Code and Kiro, connected to Xora DLC and the AAP agent. Install and manage with `xdlc` | Python 3.9+ (macOS, Linux, Windows); `aidlc` not required |

The two groups share no files; changes in one do not affect the other.
