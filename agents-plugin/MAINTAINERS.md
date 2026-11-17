# Notes for maintainers

This file is not needed by plugin users; see [README.md](README.md) for usage.

## How the plugin runs

- `xdlc` (a single Python file at the repo root) installs the `xora-dev` plugin (`dev/`) for Kiro CLI or Claude Code, stores the OAuth token in `~/.xdlc/auth.json` (Windows: DPAPI), and runs a local MCP server `xdlc mcp` (stdio). `xdlc mcp` forwards every request to `<backend>/mcp/user` with the token.
- Tool `xora_ask`: the Xora DLC backend takes the question like web chat, attaches the role and MCP session, then hands it to that role’s AAP agent; AAP picks Xora DLC MCP functions itself.
- The convention sensor runs right after install with no extra steps: if the repo has `convention.md` (at the root or under `docs/`), it uses the repo rules; otherwise it uses the plugin default (currently no rules; more will be added later). Details: [dev/README.md](dev/README.md#sensor-chạy-thế-nào). Alerts on push: [dev/README.md](dev/README.md#gửi-cảnh-báo-về-xora-dlc).

## Checks before opening a PR

```bash
python3 agents-plugin/test_roles.py
claude plugin validate agents-plugin/dev
```

CI (`.github/workflows/xdlc-plugin-tests.yml`) runs `test_roles.py` on Ubuntu, macOS, and Windows with Python 3.9 and 3.13. The README “Commands” section must match `xdlc` output exactly (the test checks this). `install.ps1` must stay ASCII-only (PowerShell 5.1 reads it as ANSI).

## Releasing

1. Bump `VERSION` in `xdlc` and `version` in `.claude-plugin/marketplace.json` (must match; the test checks).
2. Merge to `main`, create tag `vX.Y.Z`. `install.sh` / `install.ps1` and `xdlc update` use the latest tag.

Try a pre-release build: `XDLC_REF=<branch> xdlc update`, or `XDLC_PLUGIN_DIR=<clone>` for `install.sh` / `install.ps1` / `xdlc setup` / `xdlc update`.

## Archived roles

`ba`, `techlead`, `qa`, `reviewer`, `delivery` live under [archived/](archived/), are not in the marketplace, and are not tested. To revive one: move the folder next to `dev/`, point `.mcp.json` and `mcp.json` at `xdlc mcp` like `dev/`, and add it to `kiro_files` in `xdlc`, `marketplace.json`, and `ROLES` in `test_roles.py`.

## Known limitations

- `dev/convention.md` is the default convention (used when the repo has no file of its own): `rules` is currently empty; more will be added later.

- `xora_ask` accepts at most 3,800 characters (backend `question` limit): `upload-tech-doc` only works for short docs or you must split them.
- Every push re-sends all remaining alerts: the backend does not dedupe alerts or auto-close them when fixed.
- Pre-push checks the working-tree version of changed files, not necessarily the exact bytes in the commit being pushed.
- The backend does not keep a token list: to fully block someone, lock the account.
