# Design: AAP NL convention + agent check (no local sensor)

Date: 2026-11-18  
Status: approved  
Repo: `xoradlc-plugin` (plugin under `xora-dlc/plugin`)

## Goal

Convention checks no longer use local `convention.md`, plugin default rules, or `convention_sensor.py`. When the Developer asks the agent (Kiro / Claude) to **commit/push**, the agent fetches convention in natural language from AAP/KB, analyzes the pending changes, reports mismatches, and confirms: **fix** or **push anyway**. Push-anyway records governance alerts via existing `POST /governance/sensor-events`.

Manual `git push` on a terminal does **not** run this plugin’s convention logic (other tools’ pre-push hooks are untouched).

## Decisions (approved)

| Topic | Choice |
|---|---|
| Approach | Agent-only (Approach 1) |
| Local / default convention files | Ignored / removed from the check path |
| `convention_sensor.py` | Remove |
| Manual `git push` | Skip all plugin convention logic |
| Plugin pre-push hook | Do not install / do not own convention on push |
| When agent checks | Only when Dev asks agent to commit/push (not on every file save) |
| Gov alerts | Same as today: `POST /governance/sensor-events` with structured findings |

## Non-goals

- Deterministic regex scanning in CI/pre-push.
- Parsing NL into machine rules for a Python sensor.
- Changing husky / other tools’ pre-push hooks.
- Requiring a local `convention.md` in the repo.

## Current → target

| Today | Target |
|---|---|
| Find `convention.md` or plugin default → regex scan | Fetch convention NL from AAP (logged-in `xdlc`) |
| PostFileSave / PostToolUse runs `convention_sensor.py` | No automatic file-save sensor; guidance on commit/push only |
| `xdlc` installs pre-push → scan → `sensor-events` | No plugin pre-push convention; agent posts `sensor-events` when Dev chooses push-anyway |
| `xdlc sensor` wraps `convention_sensor.py` | `xdlc sensor` fetches NL convention (+ change list context) for the agent |

## Flow

```text
Dev: "commit / push"
        │
        ▼
Agent runs `xdlc sensor` (or equivalent helper)
        │  - requires xdlc login
        │  - calls AAP/backend for convention NL (KB)
        │  - prints convention text + changed files (git)
        ▼
Agent analyzes diff vs NL convention
        │
        ├─ no mismatches → commit/push as requested
        │
        └─ mismatches → show report → ask Dev
                ├─ fix → agent fixes → re-check (optional) → then commit/push
                └─ push anyway → agent submits POST /governance/sensor-events
                                   (structured findings) then commit/push
```

Manual `git push`: plugin hook absent/no-op → no AAP fetch, no agent analysis, no plugin gov alert from that path.

## Components

### 1. `xdlc sensor` (CLI)

- **Input:** optional flags later; v1 = current git changes (unstaged/staged/new as today for “what to look at”).
- **Behavior:**  
  - Ensure logged in; call backend/AAP path that returns coding convention NL for this context (work order / KB — exact API to confirm against platform MCP tools).  
  - Print: convention body (NL) + list of files/diff hints for the agent.  
  - Exit 0 on success; non-zero if not logged in / AAP unreachable (agent tells Dev to `xdlc login` or retry).
- **Does not:** parse rules, scan code, or POST alerts by itself (unless a thin `xdlc sensor report` subcommand is added for posting — see below).

### 2. Agent instructions (AGENTS.md / skills / hooks)

- Remove “run sensor after every save” and local `convention.md` default wording.
- Add: before commit/push requested by Dev, run `xdlc sensor`, analyze, report, confirm fix vs push-anyway.
- Remove or repurpose Kiro `PostFileSave` / Claude `PostToolUse` convention hooks so they no longer invoke `convention_sensor.py`.

### 3. Gov alert posting

- Keep request shape of `SensorEventsRequest` (file, line, rule_code, severity, message, work_order_id, step_id, …).
- Agent must map NL findings into that schema (`rule_code` may be a short stable slug the agent assigns when KB has no code).
- Prefer a small CLI helper, e.g. `xdlc sensor report --json <file|->`, that wraps the existing OAuth + `POST /governance/sensor-events` (same auth as today’s pre-push), so the agent does not hand-roll HTTP.
- Work order/step: from `xdlc link` mapping for current branch (unchanged). If unmapped, prompt Dev to `xdlc link` before posting alerts / clarify in agent copy.

### 4. Pre-push / setup cleanup

- Stop installing the xdlc pre-push hook for convention (setup local / `xdlc link` messages updated).
- Existing installs: `remove` / `update` should drop the plugin’s pre-push when no longer needed; document one-time cleanup for repos that already have the hook.
- Do not modify foreign pre-push scripts beyond removing *our* hook file when it is ours.

### 5. Docs & tests

- README / MAINTAINERS: convention check is agent+AAP on commit/push request; manual push skips plugin check.
- Delete or archive `convention_sensor.py`, default `convention.md` rules path, and related unit tests; replace with tests for `xdlc sensor` fetch + `sensor report` POST (mocked backend).

## Open points (resolve in implementation plan)

1. **Exact AAP/MCP call** for “get coding convention NL” (which tool/prompt vs dedicated HTTP). Prefer one stable backend entry if it exists; otherwise a fixed `xora_ask`-style prompt is acceptable for v1 with a clear agent contract.
2. **`rule_code`**: KB may not expose SEC-01 style codes — define agent fallback (e.g. `AAP-NL-001`) and max length 32.
3. Whether `xdlc link` remains required only for alert posting or also to scope which convention doc AAP returns.

## Success criteria

- No code path reads repo/plugin `convention.md` for automated checks.
- No `convention_sensor.py` in the shipped plugin.
- Asking the agent to push triggers NL convention fetch + analysis + confirm.
- Choosing push-anyway creates governance alerts via `POST /governance/sensor-events`.
- Manual `git push` does not run plugin convention logic; other tools’ hooks still run.
- `python3 agents-plugin/test_roles.py` updated and green.

## Out of scope / later

- Blocking push until agent has run a check.
- Deduping / auto-closing gov alerts (existing platform limitation).
- Structured convention API (Approach “JSON rules from backend”) — future upgrade path.
