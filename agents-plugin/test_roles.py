"""Self-check for the role plugins: python3 agents-plugin/test_roles.py (stdlib only, no network)."""
import json, os, re, subprocess, sys, tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parent
ROLES = ["ba", "techlead", "dev", "qa", "reviewer", "delivery"]


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
        assert json.loads((d / ".claude-plugin" / "plugin.json").read_text())["name"] == f"xora-{role}"
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
        role="qa",
    )
    assert init["result"]["serverInfo"]["name"] == "xora-qa"
    assert [t["name"] for t in tools["result"]["tools"]] == ["xora_ask", "xora_get_run"]
    assert bad_tool["result"]["isError"] and "Missing XORA_BASE_URL" in bad_tool["result"]["content"][0]["text"]
    assert unknown["error"]["code"] == -32601
    check_sensor()
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


if __name__ == "__main__":
    main()
