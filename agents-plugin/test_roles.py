"""Self-check for the role plugins: python3 agents-plugin/test_roles.py (stdlib only, no network)."""
import json, os, subprocess, sys
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
    assert [t["name"] for t in tools["result"]["tools"]] == ["xora_ask", "xora_get_run"]
    assert bad_tool["result"]["isError"] and "Missing XORA_BASE_URL" in bad_tool["result"]["content"][0]["text"]
    assert unknown["error"]["code"] == -32601
    print("roles ok")


if __name__ == "__main__":
    main()
