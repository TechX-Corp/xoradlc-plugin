# AAP NL convention + agent sensor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace local/default `convention_sensor.py` checks with AAP natural-language convention fetch + in-session agent analysis on commit/push requests; post gov alerts via existing `POST /governance/sensor-events`; remove plugin pre-push convention logic.

**Architecture:** `xdlc sensor` fetches convention NL from AAP (OAuth → `POST /mcp/user` `xora_ask` with a fixed prompt) and prints it plus changed files for the agent. Agent instructions (AGENTS.md / hooks) require that check only when the Dev asks to commit/push. `xdlc sensor report` posts structured findings to `/governance/sensor-events`. Plugin no longer installs or runs a convention pre-push hook.

**Tech Stack:** Python 3.9+ stdlib (`xdlc`), existing OAuth/`api()` helpers, platform `POST /mcp/user` + `POST /governance/sensor-events`, Kiro/Claude agent markdown + hooks.

**Spec:** `docs/superpowers/specs/aap-convention-agent-sensor-design.md`

## Global Constraints

- English user-facing CLI strings (current `xdlc` surface).
- Do not read repo/plugin `convention.md` for automated checks.
- Manual `git push` must not run plugin convention logic; do not break foreign pre-push hooks.
- `POST /governance/sensor-events` body shape unchanged (`SensorFinding` fields).
- Keep `xdlc link` for work-order mapping used when reporting alerts.
- Tests: `python3 agents-plugin/test_roles.py` must stay green; README Commands block must match `xdlc` help if help text changes.
- Prefer surgical edits; no drive-by refactors.

## File map

| File | Responsibility |
|---|---|
| `xdlc` | `fetch_convention_nl`, rewrite `cmd_sensor`, add `sensor report`, remove `pre_push` convention path / stop installing plugin pre-push |
| `agents-plugin/dev/AGENTS.md` | Commit/push-only AAP convention workflow |
| `agents-plugin/dev/hooks/hooks.json` | Remove Claude PostToolUse → `convention_sensor.py` |
| `agents-plugin/dev/kiro-hooks/xora-dev-convention-sensor.json` | Remove or replace so Kiro does not auto-run local sensor on save |
| `agents-plugin/dev/sensor/convention_sensor.py` | Delete |
| `agents-plugin/dev/convention.md` | Keep only if still useful as KB sample docs elsewhere; stop using as sensor default (delete default path usage) |
| `agents-plugin/dev/README.md`, `agents-plugin/README.md`, `agents-plugin/MAINTAINERS.md` | Document new flow |
| `agents-plugin/test_roles.py` | Drop regex-sensor/pre-push alert tests; add fetch + report tests with mock backend |

---

### Task 1: `xdlc sensor` fetches convention NL (no local scan)

**Files:**
- Modify: `xdlc` (`cmd_sensor`, new helpers near `api()`)
- Test: `agents-plugin/test_roles.py` (extend fake backend used in `check_xdlc`)

**Interfaces:**
- Consumes: `api()` / `access_token()` / `base_url()`; `git_root` + git porcelain for changed files
- Produces:
  - `CONVENTION_ASK_PROMPT: str` — fixed English prompt asking AAP for the project coding convention text from KB
  - `ask_aap(question: str, timeout: int = 120) -> str` — POST JSON-RPC to `/mcp/user` (`initialize` optional if server allows bare `tools/call`; prefer single `tools/call` `xora_ask` with `{"text": question}` using Bearer from `access_token()`). Returns answer text or raises `Abort`
  - `changed_files(root: Path) -> list[str]` — paths from `git status --porcelain` (tracked changes + untracked), existing style
  - `cmd_sensor(argv)` without `--pre-push` / without `report`: print convention NL + changed file list; exit 0; if not logged in → message + exit 2

- [ ] **Step 1: Write the failing test**

In `check_xdlc`’s fake HTTP handler (same pattern as existing OAuth/MCP mocks), accept `POST /mcp/user` JSON-RPC `tools/call` / `xora_ask` and return MCP result text `"CONVENTION_NL_BODY"`. Assert:

```python
out = x("sensor", cwd=ws)
assert "CONVENTION_NL_BODY" in out
assert "implement" in out or any(p.name.endswith(".java") or "A.java" in out for p in [])  # adapt: create a dirty file in ws first
```

Setup: after Local setup + login against fake backend, create a modified file under `ws`, run `x("sensor", cwd=ws)`.

- [ ] **Step 2: Run test to verify it fails**

Run: `python3 agents-plugin/test_roles.py`  
Expected: FAIL — current `sensor` still invokes missing/old `convention_sensor.py` or does not hit `/mcp/user`.

- [ ] **Step 3: Implement fetch + new `cmd_sensor`**

Add near `api()`:

```python
CONVENTION_ASK_PROMPT = (
    "Return the full coding convention that Developers must follow for this organization / "
    "current work, from the knowledge base. Prefer the official coding convention document. "
    "Include any automated rule IDs (e.g. SEC-01) if present. Plain text only."
)

def ask_aap(question, timeout=120):
    token = access_token()
    if not token:
        raise Abort("Not logged in to Xora DLC: run xdlc login.", 2)
    # One JSON-RPC tools/call; mirror agents-plugin MCP clients
    body = {"jsonrpc": "2.0", "id": 1, "method": "tools/call",
            "params": {"name": "xora_ask", "arguments": {"text": question[:3800]}}}
    status, out = request("POST", base_url() + "/mcp/user", body=body, token=token, timeout=timeout)
    if status == 401 and (token := access_token(force_refresh=True)):
        status, out = request("POST", base_url() + "/mcp/user", body=body, token=token, timeout=timeout)
    if status == 0:
        raise Abort(f"Could not connect to {base_url()} ({out}).", 2)
    if status != 200 or not isinstance(out, dict):
        raise Abort(f"AAP ask failed (HTTP {status}): {out}", 2)
    result = out.get("result") or {}
    if result.get("isError"):
        raise Abort((result.get("content") or [{}])[0].get("text") or "AAP error", 2)
    texts = [c.get("text") for c in (result.get("content") or []) if c.get("type") == "text" and c.get("text")]
    if not texts:
        raise Abort("AAP returned no convention text.", 2)
    return texts[0]
```

Rewrite `cmd_sensor` (non-report, non-pre-push) to call `ask_aap(CONVENTION_ASK_PROMPT)`, print header + answer + `Changed files:` list from git. Do **not** invoke `convention_sensor.py`.

- [ ] **Step 4: Run tests**

Run: `python3 agents-plugin/test_roles.py`  
Expected: new sensor assertion passes; older sensor/pre-push asserts may still fail until later tasks.

- [ ] **Step 5: Commit**

```bash
git add xdlc agents-plugin/test_roles.py
git commit -m "feat(xdlc): sensor fetches convention NL from AAP via xora_ask"
```

---

### Task 2: `xdlc sensor report` posts `sensor-events`

**Files:**
- Modify: `xdlc` (`cmd_sensor` subcommand `report`)
- Test: `agents-plugin/test_roles.py`

**Interfaces:**
- Consumes: `api()`, `link_of`, `repo_identity`, `git_root`, `git_out`
- Produces: `cmd_sensor(["report", ...])` reads JSON findings from stdin or `--file PATH`, posts `POST /governance/sensor-events`, prints `sent N alert(s) to Xora DLC (...)` on 201

Findings JSON schema (stdin):

```json
{
  "findings": [
    {"file": "A.java", "line": 1, "rule_code": "LOG-001", "severity": "red", "message": "..."}
  ]
}
```

`at` filled by CLI (UTC ISO Z). `work_order_id` / `step_id` from `xdlc link` mapping; if missing → Abort telling Dev to run `xdlc link`. `commit` = `git rev-parse HEAD`. `session_id` = `agent-push:<branch>` (not `git-push:`).

- [ ] **Step 1: Write the failing test**

After login + `link` to WO-1 in the existing fake backend test:

```python
payload = json.dumps({"findings": [
    {"file": "A.java", "line": 1, "rule_code": "LOG-001", "severity": "red", "message": "println"}
]})
out = x("sensor", "report", cwd=pp, inp=payload + "\n")
assert "sent 1 alert(s) to Xora DLC (WO-1" in out
assert seen["sensor"]  # reuse existing capture list for POST /governance/sensor-events
```

- [ ] **Step 2: Run test — expect FAIL** (`sensor report` unknown)

- [ ] **Step 3: Implement `report` branch in `cmd_sensor`**

Reuse the body construction from today’s `pre_push` (lines ~1002–1008) without calling the regex sensor. Normalize severity to `red`/`yellow`. Truncate fields to API max lengths.

- [ ] **Step 4: Run tests — report assertion PASS**

- [ ] **Step 5: Commit**

```bash
git add xdlc agents-plugin/test_roles.py
git commit -m "feat(xdlc): sensor report posts governance sensor-events"
```

---

### Task 3: Remove plugin pre-push convention path

**Files:**
- Modify: `xdlc` — delete or gut `pre_push()`, stop `install_pre_push` from `cmd_setup` / `cmd_link` / `cmd_update` refresh; keep `remove_pre_push` so `remove`/`uninstall` still delete *our* hook if present
- Modify: HELP / CMD_HELP strings that mention pre-push sensor alerts
- Test: `agents-plugin/test_roles.py` — remove asserts that expect pre-push install + “sent … alert(s)” on `git push`; keep `xdlc link` mapping tests

**Interfaces:**
- Produces: `install_pre_push` either removed or always no-op returning False; `cmd_sensor --pre-push` prints one line that plugin no longer checks on push and returns

- [ ] **Step 1: Update tests first**

Delete/adjust the block that runs `git push` and expects convention alerts + pre-push file content `xdlc pre-push`. Keep link mapping assertions.

- [ ] **Step 2: Run tests — expect FAIL/ERROR on leftover install_pre_push expectations or missing hooks**

- [ ] **Step 3: Implement removal**

- Remove calls to `install_pre_push` from setup/link/update.
- `cmd_sensor`: if `argv[:1] == ["--pre-push"]`, print `xdlc: convention checks run in the agent on commit/push, not in git pre-push.` and return (never block).
- `remove`/`uninstall`: still remove our hook file when `_ours`.

- [ ] **Step 4: Run full suite — green for this area**

- [ ] **Step 5: Commit**

```bash
git add xdlc agents-plugin/test_roles.py
git commit -m "feat(xdlc): drop plugin pre-push convention checks"
```

---

### Task 4: Delete `convention_sensor.py` and agent auto-hooks; rewrite agent instructions

**Files:**
- Delete: `agents-plugin/dev/sensor/convention_sensor.py` (and empty `sensor/` if unused)
- Modify: `agents-plugin/dev/hooks/hooks.json` — remove PostToolUse convention sensor command
- Modify: `agents-plugin/dev/kiro-hooks/xora-dev-convention-sensor.json` — delete file **or** replace prompt to only remind on commit/push (prefer **delete** + remove from `kiro_files` if that was the only hook)
- Modify: `agents-plugin/dev/AGENTS.md` — Sensor section: commit/push only; run `xdlc sensor`; analyze NL; confirm fix vs push-anyway; on push-anyway pipe findings JSON to `xdlc sensor report`
- Modify: `agents-plugin/dev/README.md`, `agents-plugin/README.md`, `agents-plugin/MAINTAINERS.md`
- Modify: `xdlc` `kiro_files()` if hook file removed
- Test: `check_sensor` / marketplace bundle asserts — rewrite or remove `check_sensor()` regex tests; assert sensor file **absent**; assert AGENTS.md contains commit/push + `xdlc sensor report`

- [ ] **Step 1: Rewrite `check_sensor` / hook tests to new expectations (fail on old files)**

- [ ] **Step 2: Run — FAIL while old files remain**

- [ ] **Step 3: Delete sensor + update hooks/AGENTS/docs/`kiro_files`**

AGENTS.md contract (English or keep Vietnamese if AGENTS still Vietnamese — **match existing AGENTS language**; do not mix). Include:

1. When Developer asks to commit or push: run `xdlc sensor`.
2. Compare pending changes to the printed convention; list mismatches with file, line if known, short rule id (≤32), severity red/yellow, message.
3. Ask: fix now, or push anyway.
4. Push anyway → `xdlc sensor report` with JSON findings on stdin, then perform git push.
5. Do not read local `convention.md` for this check. Manual terminal `git push` is out of scope for this plugin.

- [ ] **Step 4: Full test suite green**

Run: `python3 agents-plugin/test_roles.py`  
Expected: `roles ok`

- [ ] **Step 5: Commit**

```bash
git add -A agents-plugin xdlc
git commit -m "feat(agents-plugin): agent commit/push convention via AAP; remove local sensor"
```

---

### Task 5: Help text + README Commands sync

**Files:**
- Modify: `xdlc` HELP / CMD_HELP for `link`/`setup` (no “sensor sends alerts on git push” via hook)
- Modify: `agents-plugin/README.md` Commands fenced block to match `xdlc` output exactly
- Optional: expose `sensor` in help if product wants it visible (today it is HIDDEN — keep hidden unless product asks; AGENTS documents it)

- [ ] **Step 1: Align README ```text``` block with `python3 xdlc` stdout**

- [ ] **Step 2: Run `python3 agents-plugin/test_roles.py` — README equality assert passes**

- [ ] **Step 3: Commit**

```bash
git add xdlc agents-plugin/README.md
git commit -m "docs: sync README and help after AAP convention flow"
```

---

## Spec coverage check

| Spec requirement | Task |
|---|---|
| No local/default convention for checks | 1, 4 |
| Remove `convention_sensor.py` | 4 |
| `xdlc sensor` fetches AAP NL | 1 |
| Agent check only on commit/push request | 4 (AGENTS + hooks) |
| Push-anyway → `POST /governance/sensor-events` | 2 |
| Manual git push skips plugin convention | 3 |
| Do not break foreign pre-push | 3 |
| Docs + tests | 4, 5 |
| `xdlc link` still for WO map | 2, 3 |

## Placeholder / consistency review

- No TBD steps; `ask_aap` uses `/mcp/user` + `xora_ask` (same as agent MCP).
- `session_id` prefix `agent-push:` distinguishes from old `git-push:`.
- Severity and field truncation match existing pre_push body builder.

---

## Execution handoff

Plan complete and saved to `docs/superpowers/plans/aap-convention-agent-sensor.md`.

**Two execution options:**

**1. Subagent-Driven (recommended)** — fresh subagent per task, review between tasks  

**2. Inline Execution** — execute in this session with executing-plans checkpoints  

Which approach?
