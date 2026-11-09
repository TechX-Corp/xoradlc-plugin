#!/usr/bin/env python3
"""Stdio MCP server shared by the six role plugins (xora-ba, xora-dev, ...).

Every role plugin runs this same file; only XORA_ROLE differs. All calls go to one AAP agent,
which owns KB/S3 access, per-user permission checks and audit.

Env: XORA_BASE_URL (https://<platform-host>/api/v1), XORA_API_KEY (ak_...),
     XORA_USER_ID (work email), XORA_AGENT (AAP agent id, default: developer),
     XORA_ROLE (ba | techlead | dev | qa | reviewer | delivery; set by the plugin's .mcp.json),
     XORA_DLC_URL + XORA_DLC_REFRESH_TOKEN (xoradlc-platform API and an Authentik refresh token; both, with
     XORA_USER_ID, are in the .env from the xoradlc-platform page) for steering/skills,
     XORA_DLC_SESSION (default ~/.config/xora/session.json: the rotated refresh token and the current access token).

Self-check: python3 xora_mcp.py --check "hello"
Steering (SessionStart hook): python3 xora_mcp.py --steering <role>
"""
import fcntl, json, os, sys, time, uuid, urllib.error, urllib.request

BASE = os.environ.get("XORA_BASE_URL", "").rstrip("/")
KEY = os.environ.get("XORA_API_KEY", "")
USER = os.environ.get("XORA_USER_ID", "")
AGENT = os.environ.get("XORA_AGENT", "developer")
ROLE = os.environ.get("XORA_ROLE", "")
DLC = os.environ.get("XORA_DLC_URL", "").rstrip("/")
DLC_REFRESH = os.environ.get("XORA_DLC_REFRESH_TOKEN", "")
SESSION = os.environ.get("XORA_DLC_SESSION") or os.path.expanduser("~/.config/xora/session.json")
HERE = os.path.dirname(os.path.abspath(__file__))  # <plugin>/mcp; symlink kept, so ../AGENTS.md is the role's

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
    {
        "name": "xora_get_skill",
        "description": "Fetch the full instructions of one of this role's skills from Xora-DLC. "
                       "Call it first whenever a skill tells you to, then follow the returned content.",
        "inputSchema": {
            "type": "object",
            "properties": {"name": {"type": "string", "pattern": "^[a-z0-9][a-z0-9-]{0,63}$"}},
            "required": ["name"],
        },
    },
]


def _post(path, body):
    """Return (json, error_text)."""
    req = urllib.request.Request(
        f"{BASE}{path}",
        data=json.dumps(body).encode(),
        method="POST",
        headers={
            "Authorization": f"Bearer {KEY}",
            "Content-Type": "application/json",
            "X-Correlation-Id": uuid.uuid4().hex,  # 32 hex chars, matches ^[A-Za-z0-9-]{8,32}$
        },
    )
    try:
        with urllib.request.urlopen(req, timeout=300) as r:
            return json.load(r), None
    except urllib.error.HTTPError as e:
        hint = f" (Retry-After: {e.headers.get('Retry-After')}s)" if e.code == 429 else ""
        return None, f"HTTP {e.code}{hint}: {e.read().decode(errors='replace')[:2000]}"
    except (urllib.error.URLError, TimeoutError, json.JSONDecodeError) as e:
        return None, f"Request failed: {e}"


def dlc_access_token():
    """A valid xoradlc-platform access token. Authentik may rotate the refresh token on every use, so the current one
    lives in SESSION (0600), not only in the env; a lock keeps the six role plugins from spending the same one twice.
    A new XORA_DLC_REFRESH_TOKEN (a fresh .env) starts the session over."""
    os.makedirs(os.path.dirname(SESSION), mode=0o700, exist_ok=True)
    fd = os.open(SESSION, os.O_RDWR | os.O_CREAT, 0o600)
    with os.fdopen(fd, "r+") as f:
        fcntl.flock(f, fcntl.LOCK_EX)
        s = json.loads(f.read() or "{}")
        if s.get("seed") != DLC_REFRESH:
            s = {"seed": DLC_REFRESH, "refresh_token": DLC_REFRESH}
        if s.get("access_token") and s.get("expires_at", 0) > time.time() + 30:
            return s["access_token"]
        req = urllib.request.Request(f"{DLC}/auth/refresh-token", method="POST",
                                     data=json.dumps({"refresh_token": s["refresh_token"]}).encode(),
                                     headers={"content-type": "application/json"})
        with urllib.request.urlopen(req, timeout=10) as r:
            t = json.load(r)
        s.update(access_token=t["access_token"], refresh_token=t.get("refresh_token") or s["refresh_token"],
                 expires_at=time.time() + int(t["expires_in"]))
        f.seek(0)
        f.truncate()
        json.dump(s, f)
        return s["access_token"]


def get_content(path):
    """GET a xoradlc-platform plugin content doc; return (markdown, error_text)."""
    if not (DLC and DLC_REFRESH):
        return None, "Missing XORA_DLC_URL / XORA_DLC_REFRESH_TOKEN env vars"
    try:
        token = dlc_access_token()
        req = urllib.request.Request(f"{DLC}/plugin{path}", headers={"Authorization": f"Bearer {token}"})
        with urllib.request.urlopen(req, timeout=10) as r:
            return json.load(r)["content"], None
    except urllib.error.HTTPError as e:
        hint = " (sign-in expired: download a new .env from the xoradlc-platform page)" if e.url.endswith("/auth/refresh-token") else ""
        return None, f"HTTP {e.code}{hint}: {e.read().decode(errors='replace')[:500]}"
    except (urllib.error.URLError, TimeoutError, json.JSONDecodeError, KeyError) as e:
        return None, f"Request failed: {e}"


def steering(role):
    """Print the role's AGENTS.md from the API; fall back to the bundled copy so a session never starts bare."""
    text, err = get_content(f"/content/{role}/steering")
    if text is None:
        with open(os.path.join(HERE, "..", "AGENTS.md"), encoding="utf-8") as f:
            text = f.read()
        print(f"<!-- xora: steering from local plugin copy ({err}) -->", file=sys.stderr)
    print(text)


def ask(text, session_id=None):
    """Return (text, is_error)."""
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
    # ponytail: AAP run endpoint not published yet; wire it here when it is.
    return f"xora_get_run is not wired yet (AAP run API pending); run_id={run_id}", True


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
        elif name == "xora_get_skill":
            text, e = get_content(f"/content/{ROLE}/skills/{a.get('name', '')}")
            text, err = (e, True) if text is None else (text, False)
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
    if len(sys.argv) > 2 and sys.argv[1] == "--steering":
        steering(sys.argv[2])
        sys.exit(0)
    if len(sys.argv) > 2 and sys.argv[1] == "--check":
        out, err = ask(sys.argv[2])
        print(out)
        sys.exit(1 if err else 0)
    main()
