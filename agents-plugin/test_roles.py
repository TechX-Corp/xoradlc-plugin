"""Self-check for the role plugins: python3 agents-plugin/test_roles.py (stdlib only, no network)."""
import json, os, subprocess, sys, tempfile, threading
from http.server import BaseHTTPRequestHandler, HTTPServer
from pathlib import Path

ROOT = Path(__file__).resolve().parent
ROLES = ["ba", "techlead", "dev", "qa", "reviewer", "delivery"]


def rpc(*msgs, role="dev", **extra):
    env = {k: v for k, v in os.environ.items() if not k.startswith("XORA_")} | {"XORA_ROLE": role} | extra
    out = subprocess.run([sys.executable, str(ROOT / "mcp" / "xora_mcp.py")], env=env, check=True,
                         input="\n".join(json.dumps(m) for m in msgs) + "\n", capture_output=True, text=True).stdout
    return [json.loads(line) for line in out.splitlines()]


def main():
    market = json.loads((ROOT / ".claude-plugin" / "marketplace.json").read_text())
    assert sorted(p["name"] for p in market["plugins"]) == sorted(f"xora-{r}" for r in ROLES)
    for role in ROLES:
        d = ROOT / role
        assert json.loads((d / ".claude-plugin" / "plugin.json").read_text())["name"] == f"xora-{role}"
        server = json.loads((d / ".mcp.json").read_text())["mcpServers"][f"xora-{role}"]
        assert server["env"]["XORA_ROLE"] == role
        assert (d / "mcp" / "xora_mcp.py").resolve() == ROOT / "mcp" / "xora_mcp.py", f"{role}: broken symlink"
        json.loads((d / "hooks" / "hooks.json").read_text())
        assert (d / "AGENTS.md").is_file()
        skills = list((d / "skills").glob("*/SKILL.md"))
        assert skills, f"{role}: no skills"
        for s in skills:
            head = s.read_text().split("---")[1]
            assert f"name: {s.parent.name}\n" in head and "description: " in head, s

    init, tools, bad_tool, unknown, _ = rpc(
        {"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {}},
        {"jsonrpc": "2.0", "id": 2, "method": "tools/list"},
        {"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "xora_ask", "arguments": {"text": "hi"}}},
        {"jsonrpc": "2.0", "id": 4, "method": "nope"},
        {"jsonrpc": "2.0", "method": "notifications/initialized"},  # no reply expected
        {"jsonrpc": "2.0", "id": 5, "method": "ping"},
        role="qa",
    )
    assert init["result"]["serverInfo"]["name"] == "xora-qa"
    assert [t["name"] for t in tools["result"]["tools"]] == ["xora_ask", "xora_get_run", "xora_get_skill"]
    assert bad_tool["result"]["isError"] and "Missing XORA_BASE_URL" in bad_tool["result"]["content"][0]["text"]
    assert unknown["error"]["code"] == -32601

    # Steering hook without API env falls back to the plugin's own AGENTS.md (run via the role's symlink).
    clean = {k: v for k, v in os.environ.items() if not k.startswith("XORA_")}
    out = subprocess.run([sys.executable, str(ROOT / "ba" / "mcp" / "xora_mcp.py"), "--steering", "ba"], env=clean,
                         check=True, capture_output=True, text=True).stdout
    assert out.strip() == (ROOT / "ba" / "AGENTS.md").read_text().strip()

    # xora_get_skill against a local stand-in for xoradlc-platform: POST /auth/refresh-token (Authentik rotates the
    # refresh token) then GET /plugin/content/...; the second call reuses the access token kept in the session file.
    refreshes = []
    session = str(Path(tempfile.mkdtemp()) / "xora" / "session.json")

    class Fake(BaseHTTPRequestHandler):
        def reply(self, ok, body):
            self.send_response(200 if ok else 404)
            self.end_headers()
            self.wfile.write(json.dumps(body if ok else {"message": "Content not found"}).encode())

        def do_POST(self):
            body = json.loads(self.rfile.read(int(self.headers["Content-Length"])))
            refreshes.append(body)
            self.reply(self.path == "/api/auth/refresh-token" and body == {"refresh_token": "r"},
                       {"access_token": "t", "refresh_token": "r2", "expires_in": 300})

        def do_GET(self):
            ok = self.path == "/api/plugin/content/qa/skills/test-plan" and self.headers["Authorization"] == "Bearer t"
            self.reply(ok, {"content": "PLAN"})
        log_message = lambda *a: None
    srv = HTTPServer(("127.0.0.1", 0), Fake)
    threading.Thread(target=srv.serve_forever, daemon=True).start()
    url = f"http://127.0.0.1:{srv.server_port}/api"
    good, missing = rpc(
        {"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": {"name": "xora_get_skill", "arguments": {"name": "test-plan"}}},
        {"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "xora_get_skill", "arguments": {"name": "nope"}}},
        role="qa", XORA_DLC_URL=url, XORA_DLC_REFRESH_TOKEN="r", XORA_DLC_SESSION=session,
    )
    srv.shutdown()
    assert refreshes == [{"refresh_token": "r"}]  # one refresh for both calls
    kept = json.loads(Path(session).read_text())
    assert kept["refresh_token"] == "r2" and kept["seed"] == "r" and Path(session).stat().st_mode & 0o777 == 0o600
    assert good["result"] == {"content": [{"type": "text", "text": "PLAN"}], "isError": False}
    assert missing["result"]["isError"] and "HTTP 404" in missing["result"]["content"][0]["text"]
    print("roles ok")


if __name__ == "__main__":
    main()
