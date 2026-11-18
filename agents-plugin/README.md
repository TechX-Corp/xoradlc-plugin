# Xora DLC for Claude Code and Kiro CLI

The `xora-dev` plugin brings Xora DLC into the coding agent you already use (Kiro CLI or Claude Code). Install and manage it with `xdlc`:

- **Ask Xora DLC from the agent**: my work, work-order status, step BRD / design, open questions, conventions and internal precedents; answers include citations.
- **Ready-made working standards**: `implement` skill (code from approved BRD / design), `changeset` (PR dossier), `upload-tech-doc` (save docs to the KB), plus agent rules.
- **Convention checks**: when you ask the agent to commit/push, it runs `xdlc sensor` (convention from AAP/KB), reports mismatches, and can post governance alerts via `xdlc sensor report` if you push anyway.

## Prerequisites

| | Requirement |
|---|---|
| OS | macOS, Linux, Windows 10/11 (or WSL) |
| Python | 3.9+ (`python3` on macOS / Linux, `py` or `python` on Windows) |
| git | Any supported version |
| Agent | **Kiro CLI** (`kiro-cli`) or **Claude Code** (`claude`). Claude Code on Windows needs Git for Windows. Kiro CLI on native Windows is untested: if it only runs in WSL, install `xdlc` in WSL |
| Account | An Xora DLC account (SSO or local, same as the web) with a role the agent can query, e.g. DEV |
| Terminal | macOS / Linux: `curl`. Windows: PowerShell (built-in); prefer Windows Terminal |

## Install

**Warning — agent routing:** `xdlc setup` installs instruction files (including a short MCP routing guide) so the agent prefers MCP server `xora-dev` / tool `xora_ask` for Xora DLC intents (shared/work context, work orders, BRD/design, KB convention). If you already use other plugins with similar prompts, the agent can still follow *their* rules when those are stronger or the Developer overrides — expect some overlap. Open a **new** agent session after setup.

**macOS / Linux / WSL:**

```bash
curl -fsSL https://raw.githubusercontent.com/TechX-Corp/xoradlc-plugin/main/install.sh | sh
```

**Windows** (PowerShell or cmd):

```powershell
powershell -ExecutionPolicy ByPass -c "irm https://raw.githubusercontent.com/TechX-Corp/xoradlc-plugin/main/install.ps1 | iex"
```

Check:

```bash
xdlc --version
```

If something fails: see [Common install issues](#common-install-issues).

## Quick start

**1. Install the plugin for your agent.** Pick one:

For the project you are working on (`cd` into the project first):

```bash
cd ~/your-project
xdlc setup --agent kiro --scope local
```

For every repo on this machine:

```bash
xdlc setup --agent kiro --scope global
```

| Flag | Value | Meaning |
|---|---|---|
| `-a`, `--agent` | `kiro` | Kiro CLI |
| | `claude` | Claude Code |
| `-s`, `--scope` | `local` | Current project only; git does not track `xdlc` files |
| | `global` | Every repo on this machine |

Omit a flag and `xdlc` asks in the terminal (↑/↓ to move, Enter to select). If the team already committed plugin files into the repo, `xdlc` asks whether to untrack them from git or keep them tracked.

**2. Log in to Xora DLC:**

```bash
xdlc login
```

**3. (Optional) Map the branch for convention alerts:** needed so the agent can post gov alerts with `xdlc sensor report` after a push-anyway confirm:

```bash
xdlc link
```

**4. Open a new agent session in the project and ask**, for example:

> What is my work on Xora DLC?

## Commands

Run `xdlc` to see this list:

```text
xdlc 0.2.1: Xora DLC for Claude Code and Kiro CLI

Usage: xdlc <command> [options]

Commands:
  setup      Install the xora-dev plugin for Kiro CLI or Claude Code
  remove     Remove one or more setups
  link       Map the current branch to a work order (for agent convention alerts)
  login      Log in to Xora DLC
  logout     Log out of Xora DLC on this machine
  status     Backend, signed-in account and role, and where setup is installed
  use        Choose backend: deployed (prod) or local
  update     Update xdlc and the plugin to the latest release
  uninstall  Remove Xora DLC entirely from this machine

Options:
  -v, --version  Installed version and latest release
  -h, --help     This command list

Shortcuts (skip terminal prompts):
  xdlc setup -a, --agent kiro|claude  -s, --scope global|local
  xdlc remove -a, --agent kiro|claude
  xdlc link [WO-…] [-t, --step STEP] [-b, --branch BRANCH] [-l, --list]
  xdlc use prod | xdlc use local [-u, --url URL]

Scope: global = every repo on this machine; local = current project, git does not track xdlc files.
For details on one command: xdlc <command> --help
```

## Common install issues

| Issue | Fix |
|---|---|
| macOS / Linux: `xdlc: command not found` | Add `~/.local/bin` to `PATH`: `echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc` (bash: `~/.bashrc`), then open a new terminal |
| Windows: `'xdlc' is not recognized` | Open a new terminal window. Still failing: check the user `PATH` includes `%LOCALAPPDATA%\Programs\xdlc` |
| `python3 required` / `xdlc needs Python 3.9 or later` | Install Python 3.9+ (Windows: tick *Add python.exe to PATH*), open a new terminal, and run the install command again |
| Windows: `Python was not found; run without arguments to install from the Microsoft Store` | That is the Microsoft Store stub: install real Python, or turn off *App execution aliases* for `python.exe` in Settings |
| Windows: PowerShell blocks the script (ExecutionPolicy, AppLocker) | Use the install command above (includes `-ExecutionPolicy ByPass`). If a corporate machine still blocks it, ask IT |
| `could not download xdlc` | Check network access to `raw.githubusercontent.com` (corporate proxies may block it) |
| Menu does not move with ↑/↓ | Older consoles lack support: use Windows Terminal, or select by number (see [Notes](#notes)) |

## Notes

- Errors (agent reports `forbidden`, MCP `xora-dev` missing…): run `xdlc status` to see which account and role are in use and where setup is installed.
- Wrong account: sign out on the Xora DLC web app, then run `xdlc logout` and `xdlc login`.
- After `xdlc use`: open a new agent session.
- If ↑/↓ does not change the menu: set `XDLC_PLAIN=1` to select by typing numbers.

  **macOS / Linux / WSL:**

  ```bash
  XDLC_PLAIN=1 xdlc setup
  ```

  **Windows** (PowerShell):

  ```powershell
  $env:XDLC_PLAIN = "1"; xdlc setup
  ```
