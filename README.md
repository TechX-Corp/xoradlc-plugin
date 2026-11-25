# xoradlc-plugin (technical)

This repository ships **Xora DLC** integration for coding agents (`xdlc` + role/host packs) and a separate **AI-DLC** plugin.

| Path | What it is |
|---|---|
| `xdlc` / `cmd/xdlc` | Go CLI binary: login, setup/remove, MCP proxy (`xora-dlc` / `xora_ask`), link, sensor, update |
| `roles/catalog.yaml` | Enable/disable role packs + map `whoami.role` → pack id |
| `roles/<id>/AGENTS.md` | Per-role instructions (installable only if `enabled: true`) |
| `host-agents/<host>/` | Per-host pack: `manifest.yaml`, hook JSON, and Go adapters (parse transcript / agent email) |
| `standards/MCP_ROUTING.md` | Shared MCP intent → tool routing (agent-facing) |
| `host-agents/claude/hooks/` | Claude hook templates (SessionStart print; Stop / SubagentStop / SessionEnd report usage; SessionStart flushes the queue) |
| `host-agents/kiro/hooks/` | Kiro CLI hook templates (Stop / SessionEnd report usage; SessionStart flushes the queue) |
| `internal/usage/` | Shared queue + `POST /usage` pipeline (hook entrypoint; not listed in user help) |
| `install.sh` / `install.ps1` | Per-OS installer for the `xdlc` binary |
| `scripts/smoke_xdlc.py` | Stdlib smoke test against the built binary (fake local backend only) |
| `aidlc-plugin/` | Separate AI-DLC plugin `techx-dlc` — **not** used by `xdlc` |

User-facing install and commands: **[USER_GUIDE.md](USER_GUIDE.md)**.

## Architecture notes

- One MCP server name: **`xora-dlc`**. Role comes from the logged-in account (AAP), not from per-role MCP names.
- `xdlc setup` requires login; account role must be enabled in `roles/catalog.yaml`.
- Setups are per project only (cwd, never `$HOME`, where the agents keep machine-wide config); global setups from before 0.4.0 are removed by `xdlc update`.
- Host differences are declared in `manifest.yaml` (Kiro vs Claude paths). Claude does not need a marketplace for MCP / skills / hooks.
- IDE `SKILL.md` files may be hydrated at setup from an AAP allowlist (NL JSON); runtime calls go through `xora_ask`.
- Bundle on disk after install/update: `~/.xdlc/plugin/` (or `XDLC_PLUGIN_DIR` for local clones).
- **AI usage (metering):** after setup, host hooks report Claude/Kiro turns to Xora DLC `POST /usage` (tokens / credits, optional agent account email). Kiro V3 writes a turn's `usage_summary` only after its Stop hooks return, so the Kiro Stop hook hands the report to a detached `xdlc usage --deferred` child; SessionEnd also parses, so a session's last turn is sent. Failed sends wait in `~/.xdlc/usage-pending.jsonl` (30-day TTL). Convention gov alerts stay on `xdlc sensor report` → `POST /governance/sensor-events`. The hook entrypoint is not part of the user-facing command list in `USER_GUIDE.md`.

## Checks

```bash
go test ./...
go build -o xdlc ./cmd/xdlc   # Windows: -o xdlc.exe
python3 scripts/smoke_xdlc.py
```

CI (`.github/workflows/xdlc-plugin-tests.yml`) runs `go test`, builds `xdlc`, and runs the smoke script on Ubuntu, macOS, and Windows (Go 1.22; Python 3.12 for smoke only). The Commands fence in `USER_GUIDE.md` must match `xdlc` help output exactly. `install.ps1` must stay ASCII-only.

## Release

1. Bump `VERSION` in `internal/config/config.go` when ready.
2. Merge to `main` with `Version` bumped in `internal/config/config.go` — CI smoke on three OS, then auto-tags `vX.Y.Z` and publishes Release assets (`install.sh` / `install.ps1` / `xdlc update` use the latest release).

Dev overrides: `XDLC_REF=<branch>`, `XDLC_PLUGIN_DIR=<clone>`, `XDLC_API_URL` (tests only).
