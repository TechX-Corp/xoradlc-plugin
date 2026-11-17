"""Self-check for the role plugins: python3 agents-plugin/test_roles.py (stdlib only, no network)."""
import json, os, re, subprocess, sys, tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parent
ROLES = ["dev"]
WIN = os.name == "nt"  # các vai khác ở archived/, không kiểm


def main():
    market = json.loads((ROOT / ".claude-plugin" / "marketplace.json").read_text())
    assert sorted(p["name"] for p in market["plugins"]) == sorted(f"xora-{r}" for r in ROLES)
    for role in ROLES:
        d = ROOT / role
        manifest = d / ".claude-plugin" / "plugin.json"  # tuỳ chọn: thiếu thì Claude Code lấy tên từ marketplace
        assert not manifest.is_file() or json.loads(manifest.read_text())["name"] == f"xora-{role}"
        server = json.loads((d / ".mcp.json").read_text())["mcpServers"][f"xora-{role}"]
        # MCP server cục bộ `xdlc mcp`: xdlc giữ token đăng nhập và chuyển tới <backend>/mcp/user.
        assert server["args"][-1].endswith("/xdlc\" mcp"), server
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

    check_sensor()
    check_kiro()
    check_xdlc()
    check_menu()
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
    """Bản Kiro: mcp.json + POWER.md mỗi vai, tên skill không trùng giữa các vai."""
    names = [s.parent.name for r in ROLES for s in (ROOT / r / "skills").glob("*/SKILL.md")]
    dup = {n for n in names if names.count(n) > 1}
    assert not dup, f"Kiro gom mọi skill vào .kiro/skills, tên skill phải khác nhau giữa các vai: {dup}"
    for role in ROLES:
        d = ROOT / role
        srv = json.loads((d / "mcp.json").read_text())["mcpServers"][f"xora-{role}"]
        assert srv["command"] == "xdlc" and srv["args"] == ["mcp"], role
        head = (d / "POWER.md").read_text().split("---")[1]
        assert f"name: xora-{role}\n" in head and f'mcpServers: ["xora-{role}"]' in head, role
    hooks = list((ROOT / "dev" / "kiro-hooks").glob("*"))
    assert [h.name for h in hooks] == ["xora-dev-convention-sensor.json"], hooks  # Kiro CLI V3 chỉ nạp JSON v1
    spec = json.loads(hooks[0].read_text(encoding="utf-8"))
    hook = spec["hooks"][0]
    assert spec["version"] == "v1" and hook["trigger"] == "PostFileSave" and hook["action"]["type"] == "agent", spec
    assert "`xdlc sensor`" in hook["action"]["prompt"] and "convention.md" in hook["action"]["prompt"]
    assert re.search(hook["matcher"], "src/A.java") and not re.search(hook["matcher"], "README.md")


def check_xdlc():
    """xdlc: cài qua install.sh, setup / remove Kiro (project, local), login OAuth + MCP proxy với backend giả,
    status, use, version, uninstall. HOME tạm, không ra mạng."""
    import http.server, threading, base64, hashlib, urllib.parse, urllib.request
    repo = ROOT.parent
    cli = repo / "xdlc"
    market = json.loads((ROOT / ".claude-plugin" / "marketplace.json").read_text())
    version = re.search(r'^VERSION = "([^"]+)"', cli.read_text(), re.M).group(1)
    assert market["plugins"][0]["version"] == version, "VERSION trong xdlc phải khớp marketplace.json"

    seen = {"mcp": [], "sensor": []}
    steps = [{"work_order_id": "WO-1", "work_order_name": "Tích hợp eKYC", "step_id": "impl", "step_name": "Thực hiện"},
             {"work_order_id": "WO-2", "work_order_name": "Chuyển tiền", "step_id": "impl", "step_name": "Thực hiện"}]

    class Backend(http.server.BaseHTTPRequestHandler):
        def log_message(self, *a): pass
        def _json(self, code, body):
            data = json.dumps(body).encode()
            self.send_response(code); self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(data))); self.end_headers(); self.wfile.write(data)
        def _auth(self):
            return self.headers.get("Authorization") == "Bearer at-1"
        def do_GET(self):
            u = urllib.parse.urlsplit(self.path); q = dict(urllib.parse.parse_qsl(u.query))
            base = f"http://127.0.0.1:{self.server.server_port}"
            if u.path == "/.well-known/oauth-authorization-server":
                return self._json(200, {"authorization_endpoint": base + "/oauth/authorize", "token_endpoint": base + "/oauth/token"})
            if u.path == "/oauth/authorize":
                assert q["client_id"] == "xora-plugin" and q["code_challenge_method"] == "S256"
                seen["challenge"] = q["code_challenge"]
                self.send_response(302)
                self.send_header("Location", q["redirect_uri"] + "?" + urllib.parse.urlencode({"code": "c-1", "state": q["state"]}))
                self.end_headers(); return
            if u.path == "/oauth/userinfo":
                if not self._auth(): return self._json(401, {"detail": "no"})
                return self._json(200, {"sub": "u-1", "username": "dat.dev", "name": "Dat Dev", "role": "dev"})
            if u.path == "/my-work":
                if not self._auth(): return self._json(401, {"detail": "no"})
                return self._json(200, {"steps_in_progress": steps, "questions_to_answer": [], "badge": 0})
            self._json(404, {})
        def do_POST(self):
            body = self.rfile.read(int(self.headers.get("Content-Length") or 0))
            if self.path == "/oauth/token":
                f = dict(urllib.parse.parse_qsl(body.decode()))
                if f["grant_type"] == "authorization_code":
                    digest = base64.urlsafe_b64encode(hashlib.sha256(f["code_verifier"].encode()).digest()).rstrip(b"=").decode()
                    assert f["code"] == "c-1" and digest == seen["challenge"], f
                return self._json(200, {"access_token": "at-1", "refresh_token": "rt-1", "expires_in": 3600, "token_type": "Bearer"})
            if self.path == "/governance/sensor-events":
                if not self._auth(): return self._json(401, {"detail": "no"})
                seen["sensor"].append(json.loads(body))
                return self._json(201, {"received": len(seen["sensor"][-1]["findings"]), "alerts": []})
            if self.path == "/mcp/user":
                if not self._auth(): return self._json(401, {"detail": "no"})
                msg = json.loads(body); seen["mcp"].append(msg)
                if "id" not in msg:
                    self.send_response(202); self.end_headers(); return
                return self._json(200, {"jsonrpc": "2.0", "id": msg["id"], "result": {"content": [{"type": "text", "text": "Việc của bạn: WO-1"}], "isError": False}})
            self._json(404, {})

    srv = http.server.HTTPServer(("127.0.0.1", 0), Backend)
    threading.Thread(target=srv.serve_forever, daemon=True).start()
    with tempfile.TemporaryDirectory() as tmp:
        home = Path(tmp) / "home"; home.mkdir()
        env = {k: v for k, v in os.environ.items() if not k.startswith(("XDLC_", "XORA_"))} | {
            "HOME": str(home), "USERPROFILE": str(home), "XDLC_PLUGIN_DIR": str(repo), "XDLC_NO_BROWSER": "1",
            "XDLC_BIN_DIR": str(home / "bin"), "XDLC_NO_PATH": "1"}
        xdlc = home / "bin" / ("xdlc.py" if WIN else "xdlc")

        def x(*args, cwd=None, inp="", ok=True):
            r = subprocess.run([sys.executable, str(xdlc), *args], cwd=cwd or home, env=env, input=inp,
                               capture_output=True, text=True, encoding="utf-8", timeout=60)
            assert (r.returncode == 0) == ok, (args, r.returncode, r.stdout, r.stderr)
            return r.stdout + r.stderr

        if WIN:  # install.ps1 (powershell -File), không sửa PATH thật
            r = subprocess.run(["powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", str(repo / "install.ps1")],
                               env=env, capture_output=True, text=True, encoding="utf-8", errors="replace")
            assert r.returncode == 0 and f"xdlc {version}" in r.stdout, r.stdout + r.stderr
            shim = (home / "bin" / "xdlc.cmd").read_text()
            assert shim.startswith("@echo off") and '"%~dp0xdlc.py" %*' in shim, shim
            assert f"xdlc {version}" in subprocess.run([str(home / "bin" / "xdlc.cmd"), "--version"], env=env,
                                                       capture_output=True, text=True, encoding="utf-8").stdout
        else:
            out = subprocess.run(["sh", str(repo / "install.sh")], env=env, capture_output=True, text=True, check=True).stdout
            assert os.access(xdlc, os.X_OK) and f"xdlc {version}" in out, out
        assert xdlc.is_file()
        assert "setup" in x() and "uninstall" in x() and "--agent" in x("setup", "--help")
        readme = (ROOT / "README.md").read_text().split("```text\n", 1)[1].split("\n```", 1)[0]
        assert x().rstrip("\n") == readme, "README (mục Các lệnh) phải khớp đúng output của `xdlc`"
        assert "không có lệnh" in x("nope", ok=False)

        # setup Kiro Local ở ws (không phải git repo; đã có server và skill riêng của team).
        ws = Path(tmp) / "ws"; (ws / ".kiro" / "settings").mkdir(parents=True)
        (ws / ".kiro" / "settings" / "mcp.json").write_text(json.dumps({"mcpServers": {"postgres": {"command": "pg"}}}))
        (ws / ".kiro" / "skills" / "deploy-guide").mkdir(parents=True)
        (ws / ".kiro" / "skills" / "deploy-guide" / "SKILL.md").write_text("team")
        assert "global hoặc local" in x("setup", "--agent", "kiro", "--scope", "project", cwd=ws, ok=False)
        assert "Đã cài" in x("setup", "--agent", "kiro", "--scope", "local", cwd=ws)
        kiro = ws / ".kiro"
        assert (kiro / "skills" / "implement" / "SKILL.md").is_file() and (kiro / "steering" / "xora-dev.md").is_file()
        assert (kiro / "hooks" / "xora-dev-convention-sensor.json").is_file()
        servers = json.loads((kiro / "settings" / "mcp.json").read_text())["mcpServers"]
        entry = servers["xora-dev"]
        assert servers["postgres"] == {"command": "pg"} and entry["args"][-1] == "mcp"
        if WIN:  # Windows không chạy thẳng file .py: py -3 / python <xdlc.py> mcp
            assert Path(entry["command"]).name.lower() in ("py.exe", "python.exe") and entry["args"][-2] == str(xdlc), entry
        else:
            assert entry == {"command": str(xdlc), "args": ["mcp"], "disabled": False, "autoApprove": []}, entry
        # agent chạy đúng lệnh trong mcp.json: initialize trả lời được (chưa đăng nhập vẫn trả lời cục bộ)
        r = subprocess.run([entry["command"], *entry["args"]], env=env, capture_output=True, text=True, encoding="utf-8", timeout=60,
                           input=json.dumps({"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {}}) + "\n")
        assert json.loads(r.stdout)["result"]["serverInfo"]["name"] == "xoradlc", r.stdout + r.stderr
        # xdlc sensor: chạy sensor của plugin (chưa có convention.md: nhắc và không chặn)
        r = subprocess.run([sys.executable, str(xdlc), "sensor"], cwd=ws, env=env, capture_output=True, text=True, encoding="utf-8")
        assert r.returncode == 0, r.stdout + r.stderr
        assert (home / ".xdlc" / "plugin" / "agents-plugin" / "dev" / "sensor" / "convention_sensor.py").is_file()
        assert "Đã cài lại" in x("setup", "--agent=kiro", "--scope=local", cwd=ws)
        assert len(json.loads((home / ".xdlc" / "installs.json").read_text())["installs"]) == 1, "setup lại không tạo mục trùng"

        g = lambda repo, *a: subprocess.run(["git", "-C", str(repo), "-c", "user.name=t", "-c", "user.email=t@t", *a],
                                           capture_output=True, text=True, check=True).stdout
        # Local trong git repo, hỏi scope trên terminal: git không thấy file của xdlc.
        lr = Path(tmp) / "lr"; lr.mkdir(); g(lr, "init", "-q")
        assert "Global" in x("setup", "--agent", "kiro", cwd=lr, inp="2\n")
        exclude = (lr / ".git" / "info" / "exclude").read_text()
        assert "/.kiro/skills/implement/SKILL.md" in exclude and "/.kiro/settings/mcp.json" in exclude, exclude
        assert g(lr, "status", "--porcelain") == "", "Local: git không thấy file"
        assert "terminal" in x("setup", "--agent", "kiro", cwd=lr, inp="", ok=False), "không có terminal thì không tự đoán"

        # mcp.json đã được team commit: chọn "gỡ khỏi git" hoặc "giữ như hiện tại".
        for name, answer in (("tr", "1"), ("tk", "2")):
            r = Path(tmp) / name; (r / ".kiro" / "settings").mkdir(parents=True); g(r, "init", "-q")
            (r / ".kiro" / "settings" / "mcp.json").write_text('{"mcpServers": {}}')
            g(r, "add", "."); g(r, "commit", "-qm", "team")
            out = x("setup", "--agent", "kiro", "--scope", "local", cwd=r, inp=answer + "\n")
            assert ".kiro/settings/mcp.json" in out and "đã được git theo dõi" in out, out
            tracked = g(r, "ls-files", ".kiro")
            if answer == "1":
                assert tracked == "" and "/.kiro/settings/mcp.json" in (r / ".git" / "info" / "exclude").read_text()
                assert g(r, "status", "--porcelain").strip() == "D  .kiro/settings/mcp.json", "chỉ còn thay đổi gỡ khỏi git (chưa commit)"
            else:
                assert tracked.strip() == ".kiro/settings/mcp.json"
                assert g(r, "status", "--porcelain").strip() == "M .kiro/settings/mcp.json", "giữ: git thấy khối xora-dev"
        assert "Đã cài lại" in x("setup", "--agent", "kiro", "--scope", "local", cwd=Path(tmp) / "tk"), "lần sau nhớ lựa chọn, không hỏi lại"

        # login OAuth (backend giả), status, MCP proxy
        x("use", "local", "--url", f"http://127.0.0.1:{srv.server_port}")
        mcp_in = "\n".join(json.dumps(m) for m in (
            {"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2025-06-18"}},
            {"jsonrpc": "2.0", "id": 2, "method": "tools/list"},
            {"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "xora_ask", "arguments": {"text": "Việc của tôi?"}}})) + "\n"
        resp = {r["id"]: r for r in map(json.loads, x("mcp", inp=mcp_in).splitlines())}
        assert resp[1]["result"]["serverInfo"]["name"] == "xoradlc" and resp[2]["result"]["tools"][0]["name"] == "xora_ask"
        assert resp[3]["result"]["isError"] and "xdlc login" in resp[3]["result"]["content"][0]["text"], resp[3]
        assert "chưa đăng nhập" in x("status")

        p = subprocess.Popen([sys.executable, str(xdlc), "login"], env=env, stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
                             text=True, encoding="utf-8")
        url = next(l for l in p.stdout if l.startswith("http"))
        urllib.request.urlopen(url.strip(), timeout=10).read()  # "trình duyệt": authorize → loopback callback
        rest = p.stdout.read(); assert p.wait(timeout=30) == 0, rest
        assert "Đã đăng nhập: Dat Dev (dat.dev) · vai DEV" in rest, rest
        auth = json.loads((home / ".xdlc" / "auth.json").read_text(encoding="utf-8"))
        if WIN:  # DPAPI: file không chứa token dạng rõ
            assert list(auth) == ["dpapi"] and "at-1" not in json.dumps(auth), auth
        else:
            assert oct((home / ".xdlc" / "auth.json").stat().st_mode & 0o777) == "0o600" and auth["local"]["access_token"] == "at-1"
        st = x("status")
        assert "Dat Dev" in st and "local" in st and st.count("Kiro    Local") == 4, st
        resp = {r["id"]: r for r in map(json.loads, x("mcp", inp=mcp_in).splitlines())}
        assert resp[3]["result"]["content"][0]["text"] == "Việc của bạn: WO-1", resp[3]
        assert any(m.get("method") == "tools/call" and m["params"]["name"] == "xora_ask" for m in seen["mcp"]), seen

        # pre-push: map work order cho nhánh (xdlc link), push thật lên remote bare, sensor gửi cảnh báo về backend
        bare, pp = Path(tmp) / "pp.git", Path(tmp) / "pp"
        g(Path(tmp), "init", "-q", "--bare", str(bare)); pp.mkdir(); g(pp, "init", "-q", "-b", "main")
        g(pp, "remote", "add", "origin", str(bare))
        rules = {"version": "1", "rules": [{"id": "LOG-001", "type": "forbid", "files": ["**/*.java"], "severity": "red",
                                            "pattern": "System\\.out\\.println", "message": "In ra console", "fix": "dùng logger"}]}
        (pp / "convention.md").write_text("```json\n" + json.dumps(rules) + "\n```\n", encoding="utf-8")
        g(pp, "add", "."); g(pp, "commit", "-qm", "init")
        out = x("setup", "--agent", "kiro", "--scope", "local", cwd=pp)
        assert "pre-push" in out and "xdlc pre-push" in (pp / ".git" / "hooks" / "pre-push").read_text(), out
        assert "WO-1 --step impl" in x("link", "--list", cwd=pp) and "chưa map" in x("status", cwd=pp)
        assert "Đã map nhánh main → WO-1" in x("link", cwd=pp, inp="1\n")  # menu (gõ số khi không có terminal)
        assert "Đã map nhánh main → WO-2" in x("link", "WO-2", cwd=pp)  # viết tắt, cho agent
        assert "không có trong" in x("link", "WO-9", cwd=pp, ok=False)
        assert "(trước: WO-2)" in x("link", "wo-1", cwd=pp) and "WO-1 · Tích hợp eKYC" in x("status", cwd=pp)
        (pp / "A.java").write_text('class A { void a(){ System.out.println("x"); } }\n', encoding="utf-8")
        g(pp, "add", "."); g(pp, "commit", "-qm", "a")
        push = lambda *a: subprocess.run(["git", "-C", str(pp), "push", "-q", *a], env=env | {"XDLC_NO_TTY": "1"},
                                         capture_output=True, text=True, encoding="utf-8", errors="replace", timeout=120)
        r = push("origin", "main")
        assert r.returncode == 0 and "đã gửi 1 cảnh báo về Xora DLC (WO-1" in r.stdout + r.stderr, (r.stdout, r.stderr)
        sent = seen["sensor"][-1]
        assert (sent["work_order_id"], sent["step_id"], sent["commit"]) == ("WO-1", "impl", g(pp, "rev-parse", "HEAD").strip())
        assert [(f["file"], f["line"], f["rule_code"], f["severity"]) for f in sent["findings"]] == [("A.java", 1, "LOG-001", "red")]
        g(pp, "checkout", "-qb", "feature/x")  # nhánh chưa map, không có terminal: không gửi, vẫn push được
        (pp / "B.java").write_text('class B { void b(){ System.out.println("y"); } }\n', encoding="utf-8")
        g(pp, "add", "."); g(pp, "commit", "-qm", "b")
        n = len(seen["sensor"]); r = push("origin", "feature/x")
        assert r.returncode == 0 and "chưa map work order" in r.stdout + r.stderr and len(seen["sensor"]) == n, r.stdout + r.stderr
        # Kiro Global: hook sensor vào ~/.kiro/hooks; nhắc xdlc link cho pre-push
        out = x("setup", "--agent", "kiro", "--scope", "global")
        assert "xdlc link" in out and (home / ".kiro" / "hooks" / "xora-dev-convention-sensor.json").is_file(), out

        # version, update (từ bản clone), remove, logout, uninstall
        assert f"xdlc {version}" in x("version")
        check_claude_bundle(home, xdlc)
        assert "Đã cập nhật" in x("update")
        assert "Đã gỡ" in x("remove", "--agent", "kiro", cwd=ws, inp="1\ny\n")
        assert not (kiro / "skills" / "implement").exists() and (kiro / "skills" / "deploy-guide" / "SKILL.md").is_file()
        assert json.loads((kiro / "settings" / "mcp.json").read_text())["mcpServers"] == {"postgres": {"command": "pg"}}
        assert "Đã đăng xuất" in x("logout") and "chưa đăng nhập" in x("status")
        assert "Đã gỡ Xora DLC" in x("uninstall", inp="y\n")
        assert not (home / ".xdlc").exists() and not xdlc.exists()
        assert not (lr / ".kiro").exists() and "xdlc" not in (lr / ".git" / "info" / "exclude").read_text()
        assert not (pp / ".git" / "hooks" / "pre-push").exists() and not (home / ".kiro" / "hooks").exists()
    srv.shutdown()


def check_claude_bundle(home, xdlc):
    """Bản plugin Claude trong ~/.xdlc: .mcp.json và hook đều gọi xdlc (chạy được cả trên Windows)."""
    import importlib.machinery, importlib.util
    loader = importlib.machinery.SourceFileLoader("xdlc_mod", str(xdlc))
    spec = importlib.util.spec_from_loader("xdlc_mod", loader)
    mod = importlib.util.module_from_spec(spec); loader.exec_module(mod)
    with tempfile.TemporaryDirectory() as tmp:
        bundle = Path(tmp) / "agents-plugin"
        import shutil
        shutil.copytree(ROOT, bundle, ignore=shutil.ignore_patterns("archived"))
        mod.claude_bundle(bundle)
        mcp = json.loads((bundle / "dev" / ".mcp.json").read_text(encoding="utf-8"))["mcpServers"]["xora-dev"]
        assert mcp["args"][-1] == "mcp", mcp
        cmds = [h["command"] for g in json.loads((bundle / "dev" / "hooks" / "hooks.json").read_text(encoding="utf-8"))["hooks"].values()
                for e in g for h in e["hooks"]]
        assert any(" _print " in c and "${CLAUDE_PLUGIN_ROOT}/AGENTS.md" in c for c in cmds), cmds
        assert any(c.endswith(" sensor --hook claude") for c in cmds), cmds
        assert not any(c.startswith(("cat ", "python3 ")) for c in cmds), cmds


def pty_run(argv, steps, cwd, env, timeout=60):
    """Chạy xdlc trong terminal giả (pty): mỗi step (chờ chữ này, rồi gõ phím này). → (mã thoát, output)."""
    import pty, select, signal, time
    pid, fd = pty.fork()
    if pid == 0:
        os.chdir(cwd)
        os.execve(sys.executable, [sys.executable, *argv], env)
    import fcntl, struct, termios
    fcntl.ioctl(fd, termios.TIOCSWINSZ, struct.pack("HHHH", 40, 160, 0, 0))  # terminal 160 cột
    out, pos, sent, deadline = b"", 0, 0, time.time() + timeout
    while time.time() < deadline:
        if select.select([fd], [], [], 0.2)[0]:
            try:
                chunk = os.read(fd, 4096)
            except OSError:
                break
            if not chunk:
                break
            out += chunk
        while sent < len(steps) and (i := out.find(steps[sent][0].encode(), pos)) >= 0:
            pos = i + len(steps[sent][0].encode())
            os.write(fd, steps[sent][1].encode())
            sent += 1
    else:
        os.kill(pid, signal.SIGKILL)
        raise AssertionError("pty: hết giờ\n" + out.decode(errors="replace"))
    _, status = os.waitpid(pid, 0)
    return os.waitstatus_to_exitcode(status), out.decode(errors="replace")


def check_menu():
    """Menu phím mũi tên trong terminal thật: ↓ + Enter chọn scope; Space chọn nhiều mục để gỡ; Esc huỷ.
    pty chỉ có trên macOS / Linux; trên Windows menu dùng msvcrt, kiểm tay."""
    if WIN:
        return
    cli = str(ROOT.parent / "xdlc")
    with tempfile.TemporaryDirectory() as tmp:
        home = Path(tmp) / "home"; home.mkdir()
        env = {k: v for k, v in os.environ.items() if not k.startswith(("XDLC_", "XORA_"))} | {
            "HOME": str(home), "XDLC_PLUGIN_DIR": str(ROOT.parent), "TERM": "xterm-256color", "XDLC_BIN_DIR": str(home / "bin")}
        down, enter = "\x1b[B", "\r"
        for name in ("a", "b"):
            (Path(tmp) / name).mkdir()
            code, out = pty_run([cli, "setup", "--agent", "kiro"], [("Enter chọn", down + enter)], Path(tmp) / name, env)
            assert code == 0 and "› Local" in out and "Đã cài" in out, out
            assert (Path(tmp) / name / ".kiro" / "skills" / "implement" / "SKILL.md").is_file()
        code, out = pty_run([cli, "remove"], [("Enter xác nhận", "\x1b")], Path(tmp) / "a", env)
        assert code == 0 and "Đã huỷ" in out and (Path(tmp) / "a" / ".kiro").is_dir(), out
        code, out = pty_run([cli, "remove"], [("Enter xác nhận", enter), ("Chưa chọn mục nào", " " + down + " " + enter),
                                               ("[y/N]", "y" + enter)], Path(tmp) / "a", env)
        assert code == 0 and out.count("Đã gỡ") == 2, out
        assert not (Path(tmp) / "a" / ".kiro").exists() and not (Path(tmp) / "b" / ".kiro").exists()
        assert json.loads((home / ".xdlc" / "installs.json").read_text())["installs"] == []


if __name__ == "__main__":
    main()
