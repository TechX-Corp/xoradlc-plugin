"""Smoke xdlc like a user on macOS / Linux / Windows: python3 scripts/smoke_xdlc.py (stdlib only).
Uses the built Go binary at repo root; no network except a fake local OAuth/MCP backend."""
import json, os, re, subprocess, sys, tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
ROLES = ["dev"]
WIN = os.name == "nt"
REPO_BIN = ROOT / ("xdlc.exe" if WIN else "xdlc")


def main():
    # Pack layout: roles / host-agents / standards
    assert (ROOT / "roles" / "catalog.yaml").is_file()
    assert (ROOT / "standards" / "GUIDANCE.md").is_file()
    assert (ROOT / "host-agents" / "claude" / "hooks" / "session-start.json").is_file()
    for host in ("kiro", "claude"):
        assert (ROOT / "host-agents" / host / "manifest.yaml").is_file(), host
    for role in ROLES:
        assert (ROOT / "roles" / role / "AGENTS.md").is_file(), role
    assert not (ROOT / "dev").exists(), "legacy agents-plugin/ and roles/dev/sensor should be absent"
    assert all((ROOT / "roles" / r).is_dir() for r in ("dev", "ba", "qa", "techlead", "reviewer", "delivery"))
    assert not (ROOT / ".claude-plugin").exists(), "claude marketplace dropped for now"

    check_sensor()
    check_pack()
    check_xdlc()
    check_menu()
    print("roles ok")


def check_sensor():
    """AGENTS describes AAP + commit/push sensor flow; no local convention_sensor."""
    assert not (ROOT / "roles" / "dev" / "sensor" / "convention_sensor.py").is_file()
    agents = (ROOT / "roles" / "dev" / "AGENTS.md").read_text(encoding="utf-8")
    assert "xdlc sensor" in agents and "xdlc sensor report" in agents
    assert "commit" in agents.lower() or "push" in agents.lower()
    assert "convention.md" in agents.lower() and ("không đọc" in agents.lower() or "not" in agents.lower() or "Không đọc" in agents)
    assert "xora-dlc" in agents


def _catalog_roles(path):
    """Minimal YAML parse for roles/*/enabled (stdlib only)."""
    roles, cur = [], {}
    for line in path.read_text(encoding="utf-8").splitlines():
        s = line.strip()
        if s.startswith("- id:"):
            if cur:
                roles.append(cur)
            cur = {"id": s.split(":", 1)[1].strip()}
        elif s.startswith("enabled:") and cur:
            cur["enabled"] = s.split(":", 1)[1].strip() == "true"
    if cur:
        roles.append(cur)
    return roles


def check_pack():
    """catalog.yaml maps whoami; manifests name MCP xora-dlc; GUIDANCE in standards."""
    roles = _catalog_roles(ROOT / "roles" / "catalog.yaml")
    assert any(r.get("id") == "dev" and r.get("enabled") for r in roles)
    assert any(r.get("id") == "ba" and not r.get("enabled") for r in roles)
    for host in ("kiro", "claude"):
        text = (ROOT / "host-agents" / host / "manifest.yaml").read_text(encoding="utf-8")
        assert re.search(r"(?m)^\s*name:\s*xora-dlc\s*$", text), host
    routing = (ROOT / "standards" / "GUIDANCE.md").read_text(encoding="utf-8")
    assert "xora-dlc" in routing and "xora_ask" in routing


def check_xdlc():

    """xdlc: cài qua install.sh, setup / remove Kiro (project, local), login OAuth + MCP proxy với backend giả,
    status, version, uninstall. HOME tạm, không ra mạng."""
    import http.server, threading, base64, hashlib, urllib.parse, urllib.request
    repo = ROOT
    assert REPO_BIN.is_file(), f"build first: go build -o {REPO_BIN.name} ./cmd/xdlc"
    version = strip_ansi(subprocess.run(
        [str(REPO_BIN), "--version"], capture_output=True, text=True, encoding="utf-8", check=True,
    ).stdout).strip().split()[-1]
    ps1 = (repo / "install.ps1").read_bytes()
    assert ps1.isascii(), "install.ps1 phải chỉ có ký tự ASCII (PowerShell 5.1 đọc .ps1 không BOM theo ANSI)"

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
                if msg.get("method") == "initialize":
                    return self._json(200, {"jsonrpc": "2.0", "id": msg["id"], "result": {
                        "protocolVersion": "2025-06-18", "capabilities": {"tools": {}},
                        "serverInfo": {"name": "xoradlc", "version": "0"}}})
                if msg.get("method") == "tools/list":
                    return self._json(200, {"jsonrpc": "2.0", "id": msg["id"], "result": {"tools": [{"name": "xora_ask"}]}})
                text = "Việc của bạn: WO-1"
                args = ((msg.get("params") or {}).get("arguments") or {})
                q = (args.get("text") or "")
                if "JSON array of skills" in q or "skills allowed" in q.lower():
                    text = json.dumps([
                        {"name": "implement", "description": "Implement code from approved design"},
                        {"name": "changeset", "description": "Write a changeset / PR dossier"},
                    ])
                elif "coding convention" in q.lower() or ("convention" in q.lower() and "Developers must follow" in q):
                    text = "CONVENTION_NL_BODY"
                return self._json(200, {"jsonrpc": "2.0", "id": msg["id"], "result": {"content": [{"type": "text", "text": text}], "isError": False}})
            self._json(404, {})

    srv = http.server.HTTPServer(("127.0.0.1", 0), Backend)
    threading.Thread(target=srv.serve_forever, daemon=True).start()
    with tempfile.TemporaryDirectory() as tmp:
        home = Path(tmp) / "home"; home.mkdir()
        env = {k: v for k, v in os.environ.items() if not k.startswith(("XDLC_", "XORA_"))} | {
            "HOME": str(home), "USERPROFILE": str(home), "XDLC_PLUGIN_DIR": str(repo), "XDLC_NO_BROWSER": "1",
            "XDLC_BIN_DIR": str(home / "bin"), "XDLC_NO_PATH": "1",
            "XDLC_API_URL": f"http://127.0.0.1:{srv.server_port}"}
        xdlc = home / "bin" / ("xdlc.exe" if WIN else "xdlc")

        def x(*args, cwd=None, inp="", ok=True):
            r = subprocess.run([str(xdlc), *args], cwd=cwd or home, env=env, input=inp,
                               capture_output=True, text=True, encoding="utf-8", timeout=60)
            assert (r.returncode == 0) == ok, (args, r.returncode, r.stdout, r.stderr)
            return strip_ansi(r.stdout + r.stderr)

        if WIN:  # install.ps1 (powershell -File), không sửa PATH thật
            r = subprocess.run(["powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", str(repo / "install.ps1")],
                               env=env, capture_output=True, text=True, encoding="utf-8", errors="replace")
            assert r.returncode == 0 and f"xdlc {version}" in r.stdout, r.stdout + r.stderr
            assert f"xdlc {version}" in strip_ansi(subprocess.run(
                [str(xdlc), "--version"], env=env, capture_output=True, text=True, encoding="utf-8",
            ).stdout)
        else:
            out = subprocess.run(["sh", str(repo / "install.sh")], env=env, capture_output=True, text=True, encoding="utf-8", errors="replace", check=True).stdout
            assert os.access(xdlc, os.X_OK) and f"xdlc {version}" in out, out
        assert xdlc.is_file()
        assert "setup" in x() and "uninstall" in x() and "--agent" in x("setup", "--help")
        readme = (ROOT / "GUIDE.md").read_text(encoding="utf-8").split("```text\n", 1)[1].split("\n```", 1)[0]
        assert x().rstrip("\n") == readme, "GUIDE.md Commands section must match `xdlc` output exactly"
        assert "unknown command" in x("nope", ok=False)
        assert "unknown command" in x("use", ok=False)

        # login first (setup requires logged-in enabled role). Use a non-home cwd so we don't
        # hit the "install in HOME?" confirm before the login gate (macOS /var vs /private/var).
        proj = Path(tmp) / "proj"; proj.mkdir()
        assert "login" in x("setup", "--agent", "kiro", "--scope", "local", cwd=proj, ok=False).lower()
        lp = subprocess.Popen([str(xdlc), "login"], env=env, stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
                              text=True, encoding="utf-8")
        url = next(l for l in lp.stdout if l.startswith("http"))
        urllib.request.urlopen(url.strip(), timeout=10).read()
        rest = lp.stdout.read(); assert lp.wait(timeout=30) == 0, rest
        assert "Signed in: Dat Dev (dat.dev) · role DEV" in rest, rest

        # setup Kiro Local ở ws (không phải git repo; đã có server và skill riêng của team).
        ws = Path(tmp) / "ws"; (ws / ".kiro" / "settings").mkdir(parents=True)
        (ws / ".kiro" / "settings" / "mcp.json").write_text(json.dumps({"mcpServers": {"postgres": {"command": "pg"}}}), encoding="utf-8")
        (ws / ".kiro" / "skills" / "deploy-guide").mkdir(parents=True)
        (ws / ".kiro" / "skills" / "deploy-guide" / "SKILL.md").write_text("team", encoding="utf-8")
        assert "global or local" in x("setup", "--agent", "kiro", "--scope", "project", cwd=ws, ok=False)
        out = x("setup", "--agent", "kiro", "--scope", "local", cwd=ws)
        assert "Installed" in out and "Warning:" in out and "xora-dlc" in out, out
        kiro = ws / ".kiro"
        assert (kiro / "skills" / "implement" / "SKILL.md").is_file() and (kiro / "steering" / "xora-dlc.md").is_file()
        assert (kiro / "steering" / "GUIDANCE.md").is_file()
        assert not (kiro / "hooks" / "xora-dev-convention-sensor.json").is_file()
        servers = json.loads((kiro / "settings" / "mcp.json").read_text(encoding="utf-8"))["mcpServers"]
        entry = servers["xora-dlc"]
        assert servers["postgres"] == {"command": "pg"} and entry["args"][-1] == "mcp"
        assert Path(entry["command"]).resolve() == xdlc.resolve() and entry["args"] == ["mcp"] and not entry.get("disabled")
        assert entry.get("autoApprove") == []
        # agent chạy đúng lệnh trong mcp.json: initialize trả lời được (chưa đăng nhập vẫn trả lời cục bộ)
        r = subprocess.run([entry["command"], *entry["args"]], env=env, capture_output=True, text=True, encoding="utf-8", timeout=60,
                           input=json.dumps({"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {}}) + "\n")
        assert json.loads(r.stdout)["result"]["serverInfo"]["name"] == "xoradlc", r.stdout + r.stderr
        sens0 = x("sensor", cwd=ws)
        assert "CONVENTION_NL_BODY" in sens0, sens0
        assert not (home / ".xdlc" / "plugin" / "roles" / "dev" / "sensor" / "convention_sensor.py").is_file()
        assert "Reinstalled" in x("setup", "--agent=kiro", "--scope=local", cwd=ws)
        assert "Reinstalled" in x("setup", "-a", "kiro", "-s", "local", cwd=ws)  # tuỳ chọn ngắn
        assert "-a, --agent" in x("setup", "--help")
        assert "unexpected" in x("setup", "-a", "kiro", "-x", "y", cwd=ws, ok=False).lower()
        assert len(json.loads((home / ".xdlc" / "installs.json").read_text(encoding="utf-8"))["installs"]) == 1, "setup lại không tạo mục trùng"

        g = lambda repo, *a: subprocess.run(["git", "-C", str(repo), "-c", "user.name=t", "-c", "user.email=t@t", *a],
                                           capture_output=True, text=True, encoding="utf-8", errors="replace", check=True).stdout
        # Local trong git repo, hỏi scope trên terminal: git không thấy file của xdlc.
        lr = Path(tmp) / "lr"; lr.mkdir(); g(lr, "init", "-q")
        assert "Global" in x("setup", "--agent", "kiro", cwd=lr, inp="2\n")
        exclude = (lr / ".git" / "info" / "exclude").read_text(encoding="utf-8")
        excl_lines = {ln.strip() for ln in exclude.splitlines() if ln.strip()}
        assert "/.kiro/skills/implement/SKILL.md" in excl_lines and "/.kiro/settings/mcp.json" in excl_lines, exclude
        porc = g(lr, "status", "--porcelain")
        assert porc == "", f"Local: git không thấy file; porcelain={porc!r}; exclude={exclude!r}"
        assert "terminal" in x("setup", "--agent", "kiro", cwd=lr, inp="", ok=False), "không có terminal thì không tự đoán"

        # mcp.json đã được team commit: chọn "gỡ khỏi git" hoặc "giữ như hiện tại".
        for name, answer in (("tr", "1"), ("tk", "2")):
            r = Path(tmp) / name; (r / ".kiro" / "settings").mkdir(parents=True); g(r, "init", "-q")
            (r / ".kiro" / "settings" / "mcp.json").write_text('{"mcpServers": {}}', encoding="utf-8")
            g(r, "add", "."); g(r, "commit", "-qm", "team")
            out = x("setup", "--agent", "kiro", "--scope", "local", cwd=r, inp=answer + "\n")
            assert ".kiro/settings/mcp.json" in out and "already tracked by git" in out, out
            tracked = g(r, "ls-files", ".kiro")
            if answer == "1":
                assert tracked == "" and "/.kiro/settings/mcp.json" in (r / ".git" / "info" / "exclude").read_text(encoding="utf-8")
                assert g(r, "status", "--porcelain").strip() == "D  .kiro/settings/mcp.json", "chỉ còn thay đổi gỡ khỏi git (chưa commit)"
            else:
                assert tracked.strip() == ".kiro/settings/mcp.json"
                assert g(r, "status", "--porcelain").strip() == "M .kiro/settings/mcp.json", "giữ: git thấy khối xora-dlc"
        assert "Reinstalled" in x("setup", "--agent", "kiro", "--scope", "local", cwd=Path(tmp) / "tk"), "lần sau nhớ lựa chọn, không hỏi lại"

        # MCP proxy + status (already logged in)
        mcp_in = "\n".join(json.dumps(m) for m in (
            {"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2025-06-18"}},
            {"jsonrpc": "2.0", "id": 2, "method": "tools/list"},
            {"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "xora_ask", "arguments": {"text": "Việc của tôi?"}}})) + "\n"
        st = x("status")
        assert "Dat Dev" in st and "Backend" not in st and "Agent" in st and "Scope" in st and st.count("Kiro") == 4, st
        resp = {r["id"]: r for r in map(json.loads, x("mcp", inp=mcp_in).splitlines())}
        assert resp[3]["result"]["content"][0]["text"] == "Việc của bạn: WO-1", resp[3]
        assert any(m.get("method") == "tools/call" and m["params"]["name"] == "xora_ask" for m in seen["mcp"]), seen

        # link + AAP sensor fetch + sensor report (no plugin pre-push)
        bare, pp = Path(tmp) / "pp.git", Path(tmp) / "pp"
        g(Path(tmp), "init", "-q", "--bare", str(bare)); pp.mkdir(); g(pp, "init", "-q", "-b", "main")
        g(pp, "remote", "add", "origin", str(bare))
        (pp / "README.md").write_text("demo\n", encoding="utf-8")
        g(pp, "add", "."); g(pp, "commit", "-qm", "init")
        out = x("setup", "--agent", "kiro", "--scope", "local", cwd=pp)
        assert "Installed" in out and not (pp / ".git" / "hooks" / "pre-push").exists(), out
        assert "WO-1 --step impl" in x("link", "-l", cwd=pp) and "no work order mapped" in x("status", cwd=pp)
        assert "Mapped branch main → WO-1" in x("link", cwd=pp, inp="1\n")
        assert "Mapped branch main → WO-2" in x("link", "WO-2", "-t", "impl", "-b", "main", cwd=pp)
        assert "is not among" in x("link", "WO-9", cwd=pp, ok=False)
        assert "(was: WO-2)" in x("link", "wo-1", cwd=pp) and "WO-1 · Tích hợp eKYC" in x("status", cwd=pp)
        (pp / "A.java").write_text('class A { void a(){ System.out.println("x"); } }\n', encoding="utf-8")
        sens = x("sensor", cwd=pp)
        assert "CONVENTION_NL_BODY" in sens and "A.java" in sens, sens
        payload = json.dumps({"findings": [
            {"file": "A.java", "line": 1, "rule_code": "LOG-001", "severity": "red", "message": "println"}]})
        assert "sent 1 alert(s) to Xora DLC (WO-1" in x("sensor", "report", cwd=pp, inp=payload + "\n")
        sent = seen["sensor"][-1]
        assert (sent["work_order_id"], sent["step_id"], sent["commit"]) == ("WO-1", "impl", g(pp, "rev-parse", "HEAD").strip())
        assert [(f["file"], f["line"], f["rule_code"], f["severity"]) for f in sent["findings"]] == [("A.java", 1, "LOG-001", "red")]
        assert sent["session_id"].startswith("agent-push:")
        out = x("setup", "--agent", "kiro", "--scope", "global")
        assert "Installed" in out and not (home / ".kiro" / "hooks" / "xora-dev-convention-sensor.json").is_file(), out

        # version, update (từ bản clone), remove, logout, uninstall
        assert f"xdlc {version}" in x("--version") and f"xdlc {version}" in x("-v")
        assert "  version " not in x() and "-v, --version" in x()
        check_claude_bundle(home, xdlc)
        assert "Updated" in x("update")
        assert "Removed" in x("remove", "--agent", "kiro", cwd=ws, inp="1\n2\n")  # 1 = item to remove, 2 = Yes
        assert not (kiro / "skills" / "implement").exists() and (kiro / "skills" / "deploy-guide" / "SKILL.md").is_file()
        assert json.loads((kiro / "settings" / "mcp.json").read_text(encoding="utf-8"))["mcpServers"] == {"postgres": {"command": "pg"}}
        assert "Logged out" in x("logout") and "not logged in" in x("status")
        assert "Removed Xora DLC" in x("uninstall", inp="2\n")  # 2 = Yes
        left_xdlc = (home / ".xdlc").exists()
        left_bin = xdlc.exists()
        # Windows may leave xdlc.exe.old while the process is still running; the PATH name must be gone.
        left_old = (xdlc.with_suffix(xdlc.suffix + ".old")).exists() if WIN else False
        assert not left_xdlc and not left_bin, f"uninstall leftovers: .xdlc={left_xdlc} bin={left_bin} old={left_old} path={xdlc}"
        assert "xdlc:begin" not in (lr / ".git" / "info" / "exclude").read_text(encoding="utf-8")
        assert not (pp / ".git" / "hooks" / "pre-push").exists() and not (home / ".kiro" / "hooks").exists()
    srv.shutdown()


def check_claude_bundle(home, xdlc):
    """host-agents/claude hooks + manifest; marketplace removed."""
    man = (ROOT / "host-agents" / "claude" / "manifest.yaml").read_text(encoding="utf-8")
    assert "xora-dlc" in man and "config_path_local" in man
    assert "hooks/session-start.json" in man
    hook = json.loads((ROOT / "host-agents" / "claude" / "hooks" / "session-start.json").read_text(encoding="utf-8"))
    assert "SessionStart" in hook["hooks"]


def strip_ansi(s):
    """Drop CSI sequences so asserts work when the CLI paints TTY output."""
    return re.sub(r"\x1b\[[0-9;?]*[A-Za-z]", "", s)


def pty_run(argv, steps, cwd, env, timeout=60):
    """Chạy xdlc trong terminal giả (pty): mỗi step (chờ chữ này, rồi gõ phím này). → (mã thoát, output)."""
    import pty, select, signal, time
    pid, fd = pty.fork()
    if pid == 0:
        os.chdir(cwd)
        os.execve(argv[0], argv, env)
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
        plain = strip_ansi(out.decode(errors="replace"))
        while sent < len(steps) and (i := plain.find(steps[sent][0], pos)) >= 0:
            pos = i + len(steps[sent][0])
            os.write(fd, steps[sent][1].encode())
            sent += 1
    else:
        os.kill(pid, signal.SIGKILL)
        raise AssertionError("pty: hết giờ\n" + out.decode(errors="replace"))
    _, status = os.waitpid(pid, 0)
    return os.waitstatus_to_exitcode(status), out.decode(errors="replace")


def check_menu():
    """Menu phím mũi tên: login + setup Local; remove multi-select. pty: macOS/Linux only."""
    assert REPO_BIN.is_file(), f"build first: go build -o {REPO_BIN.name} ./cmd/xdlc"
    if WIN:
        return
    import http.server, threading, base64, hashlib, urllib.parse, urllib.request
    seen = {}
    class Backend(http.server.BaseHTTPRequestHandler):
        def log_message(self, *a): pass
        def _json(self, code, body):
            raw = json.dumps(body).encode()
            self.send_response(code); self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(raw))); self.end_headers(); self.wfile.write(raw)
        def _auth(self):
            return (self.headers.get("Authorization") or "").endswith("at-1")
        def do_GET(self):
            if self.path.startswith("/.well-known/oauth-authorization-server"):
                return self._json(200, {"authorization_endpoint": f"http://127.0.0.1:{srv.server_port}/auth",
                                        "token_endpoint": f"http://127.0.0.1:{srv.server_port}/oauth/token"})
            if self.path.startswith("/auth"):
                q = urllib.parse.parse_qs(urllib.parse.urlparse(self.path).query)
                seen["challenge"] = q["code_challenge"][0]
                loc = q["redirect_uri"][0] + "?" + urllib.parse.urlencode({"code": "c-1", "state": q["state"][0]})
                self.send_response(302); self.send_header("Location", loc); self.end_headers(); return
            if self.path == "/oauth/userinfo":
                if not self._auth(): return self._json(401, {})
                return self._json(200, {"name": "Dat Dev", "username": "dat.dev", "role": "DEV"})
            if self.path == "/my-work":
                return self._json(200, {"steps_in_progress": []})
            self._json(404, {})
        def do_POST(self):
            body = self.rfile.read(int(self.headers.get("Content-Length") or 0))
            if self.path == "/oauth/token":
                return self._json(200, {"access_token": "at-1", "refresh_token": "rt-1", "expires_in": 3600, "token_type": "Bearer"})
            if self.path == "/mcp/user":
                if not self._auth(): return self._json(401, {})
                msg = json.loads(body)
                text = json.dumps([{"name": "implement", "description": "code"}])
                return self._json(200, {"jsonrpc": "2.0", "id": msg.get("id"), "result": {"content": [{"type": "text", "text": text}], "isError": False}})
            self._json(404, {})

    srv = http.server.HTTPServer(("127.0.0.1", 0), Backend)
    threading.Thread(target=srv.serve_forever, daemon=True).start()
    with tempfile.TemporaryDirectory() as tmp:
        home = Path(tmp) / "home"; home.mkdir()
        env = {k: v for k, v in os.environ.items() if not k.startswith(("XDLC_", "XORA_"))} | {
            "HOME": str(home), "XDLC_PLUGIN_DIR": str(ROOT), "TERM": "xterm-256color",
            "XDLC_BIN_DIR": str(home / "bin"), "XDLC_NO_BROWSER": "1",
            "XDLC_API_URL": f"http://127.0.0.1:{srv.server_port}"}
        subprocess.run(["sh", str(ROOT / "install.sh")], env=env, capture_output=True, text=True, check=True)
        xdlc = home / "bin" / "xdlc"
        lp = subprocess.Popen([str(xdlc), "login"], env=env, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True)
        url = next(l for l in lp.stdout if l.startswith("http"))
        urllib.request.urlopen(url.strip(), timeout=10).read()
        assert lp.wait(timeout=30) == 0
        down, enter = "\x1b[B", "\r"
        for name in ("a", "b"):
            (Path(tmp) / name).mkdir()
            code, out = pty_run([str(xdlc), "setup", "--agent", "kiro"], [("Enter", down + enter)], Path(tmp) / name, env)
            plain = strip_ansi(out)
            assert code == 0 and "Local" in plain and "Installed" in plain, out
            assert (Path(tmp) / name / ".kiro" / "steering" / "xora-dlc.md").is_file()
        code, out = pty_run([str(xdlc), "remove"], [("current directory", "q")], Path(tmp) / "a", env)
        assert code == 0 and "Cancelled" in strip_ansi(out) and (Path(tmp) / "a" / ".kiro").is_dir(), out
        code, out = pty_run([str(xdlc), "remove"], [("current directory", enter), ("Nothing selected", "a" + enter),
                                               ("Continue?", down + enter)], Path(tmp) / "a", env)
        assert code == 0 and strip_ansi(out).count("Removed") == 2, out
        for name in ("a", "b"):
            kiro = Path(tmp) / name / ".kiro"
            assert not (kiro / "steering" / "xora-dlc.md").exists(), kiro
            assert not (kiro / "skills" / "implement").exists(), kiro
        installs_path = home / ".xdlc" / "installs.json"
        if installs_path.is_file():
            installs = json.loads(installs_path.read_text(encoding="utf-8")).get("installs") or []
            assert installs == [], installs
    srv.shutdown()


if __name__ == "__main__":
    main()
