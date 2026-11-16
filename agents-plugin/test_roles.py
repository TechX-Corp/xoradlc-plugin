"""Self-check for the role plugins: python3 agents-plugin/test_roles.py (stdlib only, no network)."""
import json, os, re, subprocess, sys, tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parent
ROLES = ["dev"]  # các vai khác ở archived/, không kiểm


def main():
    market = json.loads((ROOT / ".claude-plugin" / "marketplace.json").read_text())
    assert sorted(p["name"] for p in market["plugins"]) == sorted(f"xora-{r}" for r in ROLES)
    for role in ROLES:
        d = ROOT / role
        manifest = d / ".claude-plugin" / "plugin.json"  # tuỳ chọn: thiếu thì Claude Code lấy tên từ marketplace
        assert not manifest.is_file() or json.loads(manifest.read_text())["name"] == f"xora-{role}"
        server = json.loads((d / ".mcp.json").read_text())["mcpServers"][f"xora-{role}"]
        # MCP từ xa của Xora DLC, đăng nhập OAuth (client public cố định): không key, không chạy gì trên máy.
        assert server["type"] == "http" and server["url"].endswith("/mcp/user"), server
        assert server["oauth"] == {"clientId": "xora-plugin"}, server
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
    """Bản Kiro: mcp.json + POWER.md mỗi vai, tên skill không trùng, cài thử vào workspace tạm (cả qua install.sh)."""
    names = [s.parent.name for r in ROLES for s in (ROOT / r / "skills").glob("*/SKILL.md")]
    dup = {n for n in names if names.count(n) > 1}
    assert not dup, f"Kiro gom mọi skill vào .kiro/skills, tên skill phải khác nhau giữa các vai: {dup}"
    for role in ROLES:
        d = ROOT / role
        srv = json.loads((d / "mcp.json").read_text())["mcpServers"][f"xora-{role}"]
        assert srv["url"] == "https://xoradlc-platform-api.techxcorp.com/mcp/user", role
        assert srv["oauth"] == {"clientId": "xora-plugin"} and "command" not in srv, role
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
        sensor = home / ".agent-plugins" / "xora-dev" / "sensor" / "convention_sensor.py"
        assert sensor.is_file() and not sensor.is_symlink(), "hook Kiro chạy sensor từ bản cài này"
        servers = json.loads((kiro / "settings" / "mcp.json").read_text())["mcpServers"]
        assert set(servers) == {"xora-dev"} and servers["xora-dev"]["oauth"]["clientId"] == "xora-plugin"

        # Lệnh một dòng: install.sh (curl | sh) với bản clone này, XORA_URL đổi backend (vd. local).
        ws2 = Path(tmp) / "ws2"; ws2.mkdir()
        repo = ROOT.parent
        subprocess.run(["sh", str(repo / "install.sh"), "kiro", str(ws2)], check=True, capture_output=True,
                       env=env | {"XORA_PLUGIN_DIR": str(repo), "XORA_URL": "http://localhost:8000/"})
        url = json.loads((ws2 / ".kiro" / "settings" / "mcp.json").read_text())["mcpServers"]["xora-dev"]["url"]
        assert url == "http://localhost:8000/mcp/user", url
        assert (ws2 / ".kiro" / "skills" / "implement" / "SKILL.md").is_file()
        usage = subprocess.run(["sh", str(repo / "install.sh"), "kiro"], env=env, capture_output=True, text=True)
        assert usage.returncode != 0 and "workspace" in usage.stderr, "thiếu đích cài thì phải báo cách dùng"
        bad = subprocess.run(["sh", str(ROOT / "install-kiro.sh"), "dev", str(ROOT / "dev")], env=env, capture_output=True, text=True)
        assert bad.returncode != 0, "phải từ chối cài vào chính thư mục plugin"



if __name__ == "__main__":
    main()
