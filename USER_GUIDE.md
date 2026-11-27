# Xora DLC — user guide

Use Xora DLC from **Kiro CLI** or **Claude Code** with the `xdlc` command: ask about work orders and KB via MCP, follow role instructions, and run convention checks when you commit/push.

## Prerequisites

| | Requirement |
|---|---|
| OS | macOS, Linux, Windows 10/11 (or WSL) |
| git | Any supported version |
| Agent | **Kiro CLI** (`kiro-cli`) or **Claude Code** (`claude`) |
| Account | Xora DLC account (SSO or local) with a role enabled in this plugin (e.g. DEV) |
| Terminal | macOS / Linux: `curl`. Windows: PowerShell; prefer Windows Terminal |

## Install

**macOS / Linux / WSL:**

```bash
curl -fsSL https://raw.githubusercontent.com/TechX-Corp/xoradlc-plugin/main/install.sh | sh
```

**Windows** (PowerShell):

```powershell
powershell -ExecutionPolicy ByPass -c "irm https://raw.githubusercontent.com/TechX-Corp/xoradlc-plugin/main/install.ps1 | iex"
```

Check:

```bash
xdlc --version
```

If something fails: see [Common install issues](#common-install-issues).

## Quick start

**1. Log in to Xora DLC**

```bash
xdlc login
```

**2. Install the pack for your agent**

Run it inside each project you work on (not in your home directory); git does not track `xdlc` files:

```bash
cd ~/your-project
xdlc setup --agent kiro
```

| Flag | Value | Meaning |
|---|---|---|
| `-a`, `--agent` | `kiro` | Kiro CLI |
| | `claude` | Claude Code |

Omit the flag and `xdlc` asks in the terminal (↑/↓, Enter). Open a **new** agent session after setup.

Global setups (`--scope global`, before 0.4.0) are no longer supported: `xdlc update` removes them, keeping any file you edited and your own MCP servers. Run `xdlc setup` in each project instead.

**Warning — agent routing:** setup installs MCP routing so the agent prefers MCP **`xora-dlc`** / tool **`xora_ask`** for Xora DLC intents. Other plugins may still win if their rules are stronger or you override.

**3. (Optional) Map the branch for convention alerts**

So the agent can post gov alerts with `xdlc sensor report` after a push-anyway confirm:

```bash
xdlc link
```

**4. Open a new agent session and ask**, for example:

> What is my work on Xora DLC?

## Commands

```text
xdlc 0.5.0: Xora DLC for Claude Code and Kiro CLI

Usage: xdlc <command> [options]

Commands:
  setup      Install the xora-dlc pack for Kiro CLI or Claude Code
  remove     Remove one or more setups
  link       Map the current branch to a work order (for agent convention alerts)
  login      Log in to Xora DLC
  logout     Log out of Xora DLC on this machine
  status     Signed-in account and setups (short id per row)
  view       Show one setup by id from status
  update     Update xdlc and the plugin to the latest release
  uninstall  Remove Xora DLC entirely from this machine

Options:
  -v, --version  Print installed version
  -h, --help     This command list

Shortcuts (skip terminal prompts):
  xdlc setup -a, --agent kiro|claude
  xdlc remove -a, --agent kiro|claude
  xdlc link [WO-…] [-t, --step STEP] [-b, --branch BRANCH] [-l, --list]

Setup installs into the current project; git does not track xdlc files.
For details on one command: xdlc <command> --help
```

## Common install issues

| Issue | Fix |
|---|---|
| macOS / Linux: `xdlc: command not found` | Add `~/.local/bin` to `PATH`, then open a new terminal |
| Windows: `'xdlc' is not recognized` | Open a new terminal; check user `PATH` includes `%LOCALAPPDATA%\Programs\xdlc` |
| PowerShell blocks the script | Use the install command above (`-ExecutionPolicy ByPass`) |
| `could not download xdlc` / no release yet | Need a GitHub Release (`vX.Y.Z`) with a binary for your OS; check access to `github.com` |
| `Current directory is your home … not a project` | `cd` into a project directory, then run `xdlc setup` again |
| Menu does not move with ↑/↓ | Use Windows Terminal, or select by number (see [Notes](#notes)) |

## Notes

- Errors (`forbidden`, MCP missing…): run `xdlc status`.
- Wrong account: sign out on the Xora DLC web app, then `xdlc logout` and `xdlc login`.
- After setup or update: open a new agent session.
- Older installs (e.g. MCP `xora-dev`): `xdlc update` or `remove` then `setup` again.
- If ↑/↓ does not move the menu: set `XDLC_PLAIN=1` to select by typing numbers.

  **macOS / Linux / WSL:**

  ```bash
  XDLC_PLAIN=1 xdlc setup
  ```

  **Windows** (PowerShell):

  ```powershell
  $env:XDLC_PLAIN = "1"; xdlc setup
  ```
