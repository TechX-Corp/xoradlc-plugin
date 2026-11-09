#!/usr/bin/env python3
"""Stdio MCP server shared by the six role plugins (xora-ba, xora-dev, ...).

Every role plugin runs this same file; only XORA_ROLE differs. All calls go to one AAP agent,
which owns KB/S3 access, per-user permission checks and audit.

Env: XORA_BASE_URL (https://<platform-host>/api/v1), XORA_API_KEY (ak_...),
     XORA_USER_ID (work email), XORA_AGENT (AAP agent id, default: developer),
     XORA_ROLE (ba | techlead | dev | qa | reviewer | delivery; set by the plugin's .mcp.json),
     XORA_API (invoke | runs; default invoke). "runs" = Xora /v1 runs contract:
       POST {BASE}/agents/{AGENT}/runs {"input":{"question"}} -> run_id, then poll GET .../runs/{run_id}.
       Example: XORA_BASE_URL=https://<host>/v1 XORA_API=runs XORA_AGENT=agent_xxx XORA_API_KEY=xap_...
     XORA_TIMEOUT (seconds to wait for a run, default 180).
Self-check: python3 xora_mcp.py --check "hello"
"""
import json, os, sys, time, uuid, urllib.error, urllib.request

BASE = os.environ.get("XORA_BASE_URL", "").rstrip("/")
KEY = os.environ.get("XORA_API_KEY", "")
USER = os.environ.get("XORA_USER_ID", "")
AGENT = os.environ.get("XORA_AGENT", "developer")
ROLE = os.environ.get("XORA_ROLE", "")
API = os.environ.get("XORA_API", "invoke")
TIMEOUT = float(os.environ.get("XORA_TIMEOUT", "180"))

TOOLS = [
    {
        "name": "xora_ask",
        "description": "Ask the AAP agent (company knowledge base, cited answers; it also reads/writes KB "
                       "and S3 artifacts on your behalf). Pass session_id from a previous answer to continue.",
        "inputSchema": {
            "type": "object",
            "properties": {
                "text": {"type": "string", "minLength": 1, "maxLength": 20000},
                "session_id": {"type": "string"},
            },
            "required": ["text"],
        },
    },
    {
        "name": "xora_get_run",
        "description": "Get the status/result of a long-running AAP run by run_id.",
        "inputSchema": {
            "type": "object",
            "properties": {"run_id": {"type": "string", "minLength": 1}},
            "required": ["run_id"],
        },
    },
]


def _post(path, body, method="POST", extra=None):
    """Return (json, error_text). body=None sends no payload (GET)."""
    headers = {
        "Authorization": f"Bearer {KEY}",
        "Content-Type": "application/json",
        "X-Correlation-Id": uuid.uuid4().hex,  # 32 hex chars, matches ^[A-Za-z0-9-]{8,32}$
    }
    headers.update(extra or {})
    req = urllib.request.Request(
        f"{BASE}{path}",
        data=json.dumps(body).encode() if body is not None else None,
        method=method,
        headers=headers,
    )
    try:
        with urllib.request.urlopen(req, timeout=300) as r:
            return json.load(r), None
    except urllib.error.HTTPError as e:
        hint = f" (Retry-After: {e.headers.get('Retry-After')}s)" if e.code == 429 else ""
        return None, f"HTTP {e.code}{hint}: {e.read().decode(errors='replace')[:2000]}"
    except (urllib.error.URLError, TimeoutError, json.JSONDecodeError) as e:
        return None, f"Request failed: {e}"


def _format_run(d):
    """Format a /v1 run. Return (text, is_error)."""
    status = d.get("status")
    if status in ("queued", "running"):
        return f"Run {d.get('run_id')} còn {status}; gọi xora_get_run với run_id này để lấy kết quả.", False
    if status != "succeeded":
        return f"Run {d.get('run_id')} {status}: {d.get('error')}", True
    out = d.get("output") or {}
    lines = [out.get("answer", ""), "", f"[run_id: {d.get('run_id')} | evidence: {out.get('evidence_status')}]"]
    for c in out.get("citations") or []:
        lines.append(f"[{c.get('n')}] {c.get('document')} {c.get('section') or ''} {c.get('version') or ''}".rstrip())
    return "\n".join(lines), False


def ask_runs(text, session_id=None):
    if not (BASE and KEY):
        return "Missing XORA_BASE_URL / XORA_API_KEY env vars", True
    # ponytail: /v1 runs rejects extra input fields (no role, no user, no session); role travels as a text prefix,
    # and every run is independent. Drop this when the AAP /invoke contract is reachable.
    question = (f"[Vai người hỏi: {ROLE}]\n" if ROLE else "") + text
    d, err = _post(f"/agents/{AGENT}/runs", {"input": {"question": question}},
                   extra={"Idempotency-Key": f"mcp-{uuid.uuid4().hex}"})
    if d is None:
        return err, True
    deadline = time.time() + TIMEOUT
    while d.get("status") in ("queued", "running", None) and time.time() < deadline:
        time.sleep(2)
        d, err = _post(f"/agents/{AGENT}/runs/{d.get('run_id') or d.get('id')}", None, method="GET")
        if d is None:
            return err, True
    text, is_err = _format_run(d)
    if session_id:
        text += "\n(API runs không giữ session_id: mỗi lượt độc lập, hãy gửi kèm ngữ cảnh cần thiết.)"
    return text, is_err


def ask(text, session_id=None):
    """Return (text, is_error)."""
    if API == "runs":
        return ask_runs(text, session_id)
    if not (BASE and KEY and USER):
        return "Missing XORA_BASE_URL / XORA_API_KEY / XORA_USER_ID env vars", True
    # ponytail: "role" field name not confirmed with AAP yet; adjust when the AAP contract lands.
    # ponytail: identity is self-declared (XORA_USER_ID) under a shared API key; switch to the user's SSO token
    # once AAP publishes its auth flow, so per-user permission checks are real.
    body = {"text": text, "user": {"id": USER}, "role": ROLE, "stream": False}  # no SSE, one JSON reply
    if session_id:
        body["session_id"] = session_id
    d, err = _post(f"/platform/public/agents/{AGENT}/invoke", body)
    if d is None:
        return err, True
    if d.get("error"):
        return f"Agent error: {d['error']}", True
    lines = [d.get("output", ""), "", f"[session_id: {d.get('session_id')} | evidence: {d.get('evidence_status')}]"]
    for c in d.get("citations") or []:
        if c.get("cited"):
            lines.append(f"[{c.get('n')}] {c.get('document')} {c.get('section') or ''} {c.get('link') or ''}".rstrip())
    return "\n".join(lines), False


def get_run(run_id):
    if API == "runs":
        d, err = _post(f"/agents/{AGENT}/runs/{run_id}", None, method="GET")
        return (err, True) if d is None else _format_run(d)
    # ponytail: AAP /invoke has no run endpoint published yet; wire it here when it is.
    return f"xora_get_run is not wired yet for XORA_API=invoke (AAP run API pending); run_id={run_id}", True


def handle(msg):
    m, p = msg.get("method"), msg.get("params") or {}
    if m == "initialize":
        return {"protocolVersion": p.get("protocolVersion", "2024-11-05"),
                "capabilities": {"tools": {}},
                "serverInfo": {"name": f"xora-{ROLE or 'role'}", "version": "0.1.0"}}
    if m == "tools/list":
        return {"tools": TOOLS}
    if m == "tools/call":
        a = p.get("arguments") or {}
        name = p.get("name")
        if name == "xora_ask":
            text, err = ask(a.get("text", ""), a.get("session_id"))
        elif name == "xora_get_run":
            text, err = get_run(a.get("run_id", ""))
        else:
            text, err = f"Unknown tool: {name}", True
        return {"content": [{"type": "text", "text": text}], "isError": err}
    if m == "ping":
        return {}
    raise LookupError(m)


def main():
    for line in sys.stdin:
        if not line.strip():
            continue
        msg = json.loads(line)
        if "id" not in msg:  # notification, no reply
            continue
        try:
            resp = {"jsonrpc": "2.0", "id": msg["id"], "result": handle(msg)}
        except LookupError as e:
            resp = {"jsonrpc": "2.0", "id": msg["id"], "error": {"code": -32601, "message": f"Method not found: {e}"}}
        print(json.dumps(resp), flush=True)


if __name__ == "__main__":
    if len(sys.argv) > 2 and sys.argv[1] == "--check":
        out, err = ask(sys.argv[2])
        print(out)
        sys.exit(1 if err else 0)
    main()
