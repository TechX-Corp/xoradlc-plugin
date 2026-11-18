# xoradlc-plugin (technical)

This repository ships **Xora DLC** integration for coding agents (`xdlc` + role/host packs) and a separate **AI-DLC** plugin.

| Path | What it is |
|---|---|
| `xdlc` | CLI: login, setup/remove, MCP proxy (`xora-dlc` / `xora_ask`), link, sensor, update |
| `roles/catalog.yaml` | Enable/disable role packs + map `whoami.role` → pack id |
| `roles/<id>/AGENTS.md` | Per-role instructions (installable only if `enabled: true`) |
| `host-agents/<host>/manifest.yaml` | Host setup/remove contract (MCP paths, GUIDANCE dest, skills dir, hooks) |
| `standards/GUIDANCE.md` | Shared MCP routing guidance |
| `standards/hooks/` | Hook templates (e.g. Claude SessionStart) |
| `install.sh` / `install.ps1` | Per-OS installer for the `xdlc` binary |
| `test_roles.py` | Stdlib self-check (no network) |
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
python3 test_roles.py
```

CI (`.github/workflows/xdlc-plugin-tests.yml`) runs the same script on Ubuntu, macOS, and Windows (Python 3.9 / 3.13). The Commands fence in `GUIDE.md` must match `xdlc` help output exactly. `install.ps1` must stay ASCII-only.

## Release

1. Bump `VERSION` in `xdlc` when ready.
2. Merge to `main`, tag `vX.Y.Z` when cutting a release (`install.sh` / `install.ps1` / `xdlc update` prefer the latest tag).

Dev overrides: `XDLC_REF=<branch>`, `XDLC_PLUGIN_DIR=<clone>`, `XDLC_API_URL` (tests only).
