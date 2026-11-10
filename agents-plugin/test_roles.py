"""Self-check for the role plugins: python3 agents-plugin/test_roles.py (stdlib only, no network)."""
import json, os, re, subprocess, sys, tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parent
ROLES = ["dev"]  # các vai khác ở archived/, không kiểm


def rpc(*msgs, role="dev"):
    env = {k: v for k, v in os.environ.items() if not k.startswith("XORA_")} | {"XORA_ROLE": role}
    out = subprocess.run([sys.executable, str(ROOT / "mcp" / "xora_mcp.py")], env=env, check=True,
                         input="\n".join(json.dumps(m) for m in msgs) + "\n", capture_output=True, text=True).stdout
    return [json.loads(line) for line in out.splitlines()]


def main():
    market = json.loads((ROOT / ".claude-plugin" / "marketplace.json").read_text())
    assert sorted(p["name"] for p in market["plugins"]) == sorted(f"xora-{r}" for r in ROLES)
    for role in ROLES:
        d = ROOT / role
        manifest = d / ".claude-plugin" / "plugin.json"  # tuỳ chọn: thiếu thì Claude Code lấy tên từ marketplace
        assert not manifest.is_file() or json.loads(manifest.read_text())["name"] == f"xora-{role}"
        server = json.loads((d / ".mcp.json").read_text())["mcpServers"][f"xora-{role}"]
        assert server["env"]["XORA_ROLE"] == role
        assert (d / "mcp" / "xora_mcp.py").resolve() == ROOT / "mcp" / "xora_mcp.py", f"{role}: broken symlink"
        hooks = json.loads((d / "hooks" / "hooks.json").read_text())
        for cmd in (h["command"] for g in hooks["hooks"].values() for e in g for h in e["hooks"]):
            for f in re.findall(r'\$\{CLAUDE_PLUGIN_ROOT\}/([^"\s]+)', cmd):
                assert (d / f).is_file(), f"{role}: hook trỏ tới file không có: {f}"
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
        role="dev",
    )
    assert init["result"]["serverInfo"]["name"] == "xora-dev"
    assert [t["name"] for t in tools["result"]["tools"]] == ["xora_ask", "xora_get_run"]
    assert bad_tool["result"]["isError"] and "Missing XORA_BASE_URL" in bad_tool["result"]["content"][0]["text"]
    assert unknown["error"]["code"] == -32601
    check_sensor()
    check_kiro()
    check_runs()
    print("roles ok")


def check_sensor():
    """Sensor convention của xora-dev: bắt vi phạm, tôn trọng convention-ignore, im lặng khi chưa có convention.md."""
    sensor = ROOT / "dev" / "sensor" / "convention_sensor.py"
    with tempfile.TemporaryDirectory() as tmp:
        repo = Path(tmp)
        subprocess.run(["git", "init", "-q", str(repo)], check=True)
        bad = repo / "A.java"
        bad.write_text('class A { void a(){ System.out.println("x"); }\n'
                       '  String k = "abcdef123"; // convention-ignore: SEC-001 demo\n}\n')
        hook = lambda: subprocess.run([sys.executable, str(sensor), "--hook", "claude"], capture_output=True, text=True,
                                      input=json.dumps({"cwd": str(repo), "tool_input": {"file_path": str(bad)}})).stdout
        assert hook() == "", "chưa có convention.md thì sensor phải im lặng"
        (repo / "convention.md").write_text((ROOT / "dev" / "convention.sample.md").read_text())
        ctx = json.loads(hook())["hookSpecificOutput"]["additionalContext"]
        assert "[LOG-001]" in ctx and "SEC-001" not in ctx.split("\n- ", 1)[1], ctx

        # Đã commit: forbid chỉ báo trên dòng đổi; --full báo cả code cũ; --staged chỉ xét phần đã staged.
        g = lambda *a: subprocess.run(["git", "-C", str(repo), "-c", "user.name=t", "-c", "user.email=t@t", *a],
                                      check=True, capture_output=True)
        g("add", "."); g("commit", "-qm", "init")
        cli = lambda *a: subprocess.run([sys.executable, str(sensor), *a], cwd=repo, capture_output=True, text=True).stdout
        bad.write_text(bad.read_text()[:-2] + "} // end\n")  # chỉ sửa dòng 3
        assert hook() == "", "chỉ sửa dòng khác (dòng 1 đã có vi phạm cũ) thì không báo code cũ"
        assert "[LOG-001]" in cli("--full", str(bad))
        bad.write_text(bad.read_text() + 'class B { void b(){ System.out.println("y"); } }\n')
        assert "A.java:4 [LOG-001]" in cli(), "dòng mới thêm phải bị báo"
        assert "[LOG-001]" not in cli("--staged"), "chưa staged thì --staged không báo"
        g("add", "A.java")
        bad.write_text("\n\n" + bad.read_text())  # sửa thêm sau khi staged: số dòng working tree lệch với index
        assert "A.java:4 [LOG-001]" in cli("--staged"), "--staged phải đọc bản trong index"

        # Chỉ cảnh báo, không chặn: có vi phạm hay convention.md hỏng thì vẫn exit 0.
        code = lambda *a: subprocess.run([sys.executable, str(sensor), *a], cwd=repo, capture_output=True).returncode
        assert code() == 0, "có cảnh báo vẫn phải exit 0"
        (repo / "convention.md").write_text("```json\n{\"rules\": [\n```\n")
        assert code() == 0 and "lỗi" in json.loads(hook())["systemMessage"], "convention.md hỏng vẫn không chặn"



def check_kiro():
    """Bản Kiro: mcp.json + POWER.md mỗi vai, tên skill không trùng, cài thử vào workspace tạm rồi bắt tay MCP."""
    names = [s.parent.name for r in ROLES for s in (ROOT / r / "skills").glob("*/SKILL.md")]
    dup = {n for n in names if names.count(n) > 1}
    assert not dup, f"Kiro gom mọi skill vào .kiro/skills, tên skill phải khác nhau giữa các vai: {dup}"
    for role in ROLES:
        d = ROOT / role
        srv = json.loads((d / "mcp.json").read_text())["mcpServers"][f"xora-{role}"]
        assert srv["env"]["XORA_ROLE"] == role and f".agent-plugins/xora-{role}/mcp/xora_mcp.py" in srv["args"][-1], role
        head = (d / "POWER.md").read_text().split("---")[1]
        assert f"name: xora-{role}\n" in head and f'mcpServers: ["xora-{role}"]' in head, role
    for hook in (ROOT / "dev" / "kiro-hooks").glob("*.kiro.hook"):
        json.loads(hook.read_text())

    with tempfile.TemporaryDirectory() as tmp:
        home, ws = Path(tmp) / "home", Path(tmp) / "ws"
        ws.mkdir(parents=True); home.mkdir()
        env = {k: v for k, v in os.environ.items() if not k.startswith("XORA_")} | {"HOME": str(home)}
        for role in ("dev",):
            subprocess.run(["sh", str(ROOT / "install-kiro.sh"), role, str(ws)], env=env, check=True, capture_output=True)
        kiro = ws / ".kiro"
        assert (kiro / "steering" / "xora-dev.md").is_file() and (kiro / "hooks" / "convention-sensor.kiro.hook").is_file()
        assert (kiro / "skills" / "upload-tech-doc").is_dir()
        installed = home / ".agent-plugins" / "xora-dev" / "mcp" / "xora_mcp.py"
        assert installed.is_file() and not installed.is_symlink(), "phải chép file thật, không chép symlink"
        servers = json.loads((kiro / "settings" / "mcp.json").read_text())["mcpServers"]
        assert set(servers) == {"xora-dev"}
        srv = servers["xora-dev"]
        out = subprocess.run([srv["command"], *srv["args"]], env=env | srv["env"], capture_output=True, text=True, check=True,
                             input=json.dumps({"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {}}) + "\n").stdout
        assert json.loads(out)["result"]["serverInfo"]["name"] == "xora-dev", out
        bad = subprocess.run(["sh", str(ROOT / "install-kiro.sh"), "dev", str(ROOT / "dev")], env=env, capture_output=True, text=True)
        assert bad.returncode != 0, "phải từ chối cài vào chính thư mục plugin"



def check_runs():
    """XORA_API=runs: tạo run, chờ, trả câu trả lời; vai đi kèm câu hỏi. Server giả cục bộ, không ra mạng."""
    import http.server, threading
    seen = {}

    class Stub(http.server.BaseHTTPRequestHandler):
        def log_message(self, *a): pass
        def _send(self, code, body):
            data = json.dumps(body).encode()
            self.send_response(code); self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(data))); self.end_headers(); self.wfile.write(data)
        def do_POST(self):
            seen["body"] = json.loads(self.rfile.read(int(self.headers["Content-Length"])))
            seen["idem"] = self.headers.get("Idempotency-Key")
            self._send(202, {"run_id": "arun_1", "status": "queued"})
        def do_GET(self):
            self._send(200, {"run_id": "arun_1", "status": "succeeded",
                             "output": {"answer": "Không có BRD.", "evidence_status": "insufficient", "citations": []}})

    srv = http.server.HTTPServer(("127.0.0.1", 0), Stub)
    threading.Thread(target=srv.serve_forever, daemon=True).start()
    try:
        env = {k: v for k, v in os.environ.items() if not k.startswith("XORA_")} | {
            "XORA_API": "runs", "XORA_BASE_URL": f"http://127.0.0.1:{srv.server_port}/v1",
            "XORA_API_KEY": "xap_test", "XORA_AGENT": "agent_x", "XORA_ROLE": "dev"}
        out = subprocess.run([sys.executable, str(ROOT / "mcp" / "xora_mcp.py"), "--check", "Có BRD nào?"],
                             env=env, capture_output=True, text=True, timeout=60)
        assert out.returncode == 0 and "Không có BRD." in out.stdout and "arun_1" in out.stdout, out.stdout + out.stderr
        assert seen["body"] == {"input": {"question": "[Vai người hỏi: dev]\nCó BRD nào?"}} and seen["idem"], seen
    finally:
        srv.shutdown()


if __name__ == "__main__":
    main()
