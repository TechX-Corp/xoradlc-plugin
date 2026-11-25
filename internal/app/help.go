package app

import (
	"fmt"

	"github.com/TechX-Corp/xoradlc-plugin/internal/config"
)

// MainHelp is the top-level xdlc help text (first line used by USER_GUIDE.md).
func MainHelp() string {
	return fmt.Sprintf(`xdlc %s: Xora DLC for Claude Code and Kiro CLI

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
For details on one command: xdlc <command> --help`, config.Version)
}

// CmdHelp returns per-command help.
func CmdHelp(cmd string) (string, bool) {
	m := map[string]string{
		"setup": fmt.Sprintf(`xdlc setup [-a, --agent kiro|claude]

Install the xora-dlc pack (AGENTS.md for the logged-in role, standards, MCP, optional AAP skills)
into the current project (cwd, not your home directory); git does not track xdlc files.
Requires xdlc login; account role must be enabled in roles/catalog.yaml.
If --agent is missing, prompts on the terminal.
Convention checks run in the agent on commit/push (xdlc sensor + AAP), not in git pre-push.`),
		"remove": `xdlc remove [-a, --agent kiro|claude]

Pick setups to remove: ↑/↓ move, Space select, a select all, Enter confirm
(no terminal: type 1,3 | 1-2 | all). Only deletes files xdlc installed;
asks first if you edited a file. In mcp.json, only removes the xora-dlc block;
in Claude settings, only removes the xdlc hooks.`,
		"link": `xdlc link [WO-…] [-t, --step STEP] [-b, --branch BRANCH] [-l, --list]

Map the repo's current branch (identified by origin URL) to a step you are working on in Xora DLC,
so agent convention alerts (xdlc sensor report) go to the right work order. No args: menu of in-progress steps
(from My Work). WO-…: pick directly (add --step if that WO has multiple steps). --list: print the current map
and available steps.`,
		"login": fmt.Sprintf(`xdlc login

Opens the browser to the Xora DLC login page (SSO or local account).
If the browser is already signed in on the web, that account is used: xdlc prints the account for you to check.`),
		"logout": `xdlc logout

Deletes the Xora DLC token from this machine.`,
		"status": `xdlc status

Prints version, signed-in account, and a short setups table (id, agent, project).`,
		"view": `xdlc view <id>

Shows repository, branch, work-order map, and last update for one setup id from xdlc status.`,
		"update": `xdlc update

Downloads the latest GitHub release (tag vX.Y.Z newer than the installed version), replaces the xdlc
binary for your platform, refreshes the embedded plugin pack, then reinstalls into every setup.
If no newer release is published yet, xdlc exits with an error (no main-branch fallback).`,
		"--version": `xdlc --version (or -v)

Prints the installed version.`,
		"uninstall": `xdlc uninstall

Removes every setup, logs out, deletes ~/.xdlc and the xdlc command itself. Asks for confirmation first.`,
	}
	text, ok := m[cmd]
	return text, ok
}