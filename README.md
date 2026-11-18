# xoradlc-plugin (technical)

This repository ships **Xora DLC** integration for coding agents (`xdlc` + role/host packs) and a separate **AI-DLC** plugin.

| Path | What it is |
|---|---|
| `xdlc` / `cmd/xdlc` | Go CLI binary: login, setup/remove, MCP proxy (`xora-dlc` / `xora_ask`), link, sensor, update |
| `roles/catalog.yaml` | Enable/disable role packs + map `whoami.role` → pack id |
| `roles/<id>/AGENTS.md` | Per-role instructions (installable only if `enabled: true`) |
| `host-agents/<host>/manifest.yaml` | Host setup/remove contract (MCP paths, GUIDANCE dest, skills dir, hooks) |
| `standards/GUIDANCE.md` | Shared MCP routing guidance |
| `host-agents/claude/hooks/` | Claude-only hook templates (SessionStart) |
| `install.sh` / `install.ps1` | Per-OS installer for the `xdlc` binary |
| `scripts/smoke_xdlc.py` | Stdlib smoke test against the built binary (fake local backend only) |
| `aidlc-plugin/` | Separate AI-DLC plugin `techx-dlc` — **not** used by `xdlc` |

User-facing install and usage: **[GUIDE.md](GUIDE.md)**.

## Architecture notes

- One MCP server name: **`xora-dlc`**. Role comes from the logged-in account (AAP), not from per-role MCP names.
- `xdlc setup` requires login; account role must be enabled in `roles/catalog.yaml`.
- Host differences are declared in `manifest.yaml` (Kiro vs Claude paths). Claude does not need a marketplace for MCP / skills / hooks.
- IDE `SKILL.md` files may be hydrated at setup from an AAP allowlist (NL JSON); runtime calls go through `xora_ask`.
- Bundle on disk after install/update: `~/.xdlc/plugin/` (or `XDLC_PLUGIN_DIR` for local clones).

## Checks

```bash
go test ./...
go build -o xdlc ./cmd/xdlc   # Windows: -o xdlc.exe
python3 scripts/smoke_xdlc.py
```

CI (`.github/workflows/xdlc-plugin-tests.yml`) runs `go test`, builds `xdlc`, and runs the smoke script on Ubuntu, macOS, and Windows (Go 1.22; Python 3.12 for smoke only). The Commands fence in `GUIDE.md` must match `xdlc` help output exactly. `install.ps1` must stay ASCII-only.

## Release

1. Bump `VERSION` in `internal/config/config.go` when ready.
2. Merge to `main`, tag `vX.Y.Z` when cutting a release (`install.sh` / `install.ps1` / `xdlc update` prefer the latest tag).

Dev overrides: `XDLC_REF=<branch>`, `XDLC_PLUGIN_DIR=<clone>`, `XDLC_API_URL` (tests only).
