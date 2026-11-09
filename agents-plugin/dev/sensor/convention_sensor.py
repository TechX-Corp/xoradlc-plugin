#!/usr/bin/env python3
"""Sensor convention — kiểm code theo convention.md của dự án, chỉ cảnh báo (không chặn).

convention.md có một khối ```json chứa luật:
  {"version": "1", "exclude": ["**/node_modules/**"], "rules": [
     {"id": "SEC-001", "type": "forbid", "files": ["**/*.java"], "exclude": ["**/test/**"],
      "pattern": "regex", "message": "…", "fix": "…"} ]}

Kiểu luật:
  forbid     regex không được xuất hiện trên dòng nào (báo từng dòng)
  require    regex phải xuất hiện ít nhất một lần trong file
  filename   tên file phải khớp regex
  max_lines  file không quá "max" dòng

Bỏ qua có chủ đích: thêm `convention-ignore: <ID> <lý do>` trên chính dòng đó.

Cách chạy:
  python3 convention_sensor.py [file ...]      # không có file → các file đổi trong git (staged + chưa staged + mới)
  python3 convention_sensor.py --hook claude   # hook PostToolUse của Claude Code: đọc JSON từ stdin
  python3 convention_sensor.py --strict ...    # exit 1 nếu có cảnh báo (dùng cho CI sau này)
Tìm convention.md: biến CONVENTION_FILE, hoặc đi ngược từ thư mục file lên tới gốc repo.
"""
import fnmatch
import json
import os
import re
import subprocess
import sys
from pathlib import Path

IGNORE = re.compile(r"convention-ignore:\s*([A-Z]+-\d+)")
RULES_BLOCK = re.compile(r"```json\s*\n(.*?)\n```", re.S)


def git_root(start):
    try:
        out = subprocess.run(["git", "-C", str(start), "rev-parse", "--show-toplevel"],
                             capture_output=True, text=True, timeout=10)
        return Path(out.stdout.strip()) if out.returncode == 0 and out.stdout.strip() else None
    except (OSError, subprocess.SubprocessError):
        return None


def find_convention(start):
    if os.environ.get("CONVENTION_FILE"):
        p = Path(os.environ["CONVENTION_FILE"]).expanduser()
        return p if p.is_file() else None
    start = Path(start).resolve()
    stop = git_root(start if start.is_dir() else start.parent)
    for d in [start if start.is_dir() else start.parent, *(start.parents)]:
        for name in ("convention.md", "docs/convention.md"):
            if (d / name).is_file():
                return d / name
        if stop and d == stop:
            break
    return None


def load_rules(conv_path):
    blocks = RULES_BLOCK.findall(conv_path.read_text(encoding="utf-8"))
    for b in blocks:
        try:
            cfg = json.loads(b)
        except json.JSONDecodeError as e:
            raise SystemExit(f"convention.md: khối luật JSON sai cú pháp ({e}).")
        if isinstance(cfg, dict) and "rules" in cfg:
            for r in cfg["rules"]:
                if r.get("pattern"):
                    r["_re"] = re.compile(r["pattern"])
            return cfg
    return {"rules": []}


def match(path, patterns):
    for pat in patterns or []:
        if fnmatch.fnmatch(path, pat) or (pat.startswith("**/") and fnmatch.fnmatch(path, pat[3:])):
            return True
    return False


def check_file(path, rel, cfg):
    if match(rel, cfg.get("exclude")):
        return []
    try:
        text = Path(path).read_text(encoding="utf-8")
    except (OSError, UnicodeDecodeError):
        return []
    lines = text.splitlines()
    out = []
    for r in cfg["rules"]:
        if not match(rel, r.get("files", ["**/*"])) or match(rel, r.get("exclude")):
            continue
        kind, hit = r.get("type"), []
        if kind == "forbid":
            hit = [i for i, ln in enumerate(lines, 1)
                   if r["_re"].search(ln) and r["id"] not in IGNORE.findall(ln)]
        elif kind == "require" and not r["_re"].search(text):
            hit = [0]
        elif kind == "filename" and not r["_re"].search(Path(rel).name):
            hit = [0]
        elif kind == "max_lines" and len(lines) > int(r.get("max", 500)):
            hit = [0]
        out += [(rel, ln, r) for ln in hit]
    return out


def changed_files(root):
    files = set()
    for args in (["diff", "--name-only"], ["diff", "--name-only", "--cached"],
                 ["ls-files", "--others", "--exclude-standard"]):
        res = subprocess.run(["git", "-C", str(root), *args], capture_output=True, text=True)
        files |= {f for f in res.stdout.splitlines() if f}
    return [root / f for f in sorted(files) if (root / f).is_file()]


def run(paths):
    """Trả (convention_path, version, danh sách cảnh báo)."""
    if not paths:
        return None, None, []
    conv = find_convention(paths[0])
    if not conv:
        return None, None, []
    cfg = load_rules(conv)
    base = conv.parent
    warns = []
    for p in paths:
        p = Path(p).resolve()
        try:
            rel = p.relative_to(base).as_posix()
        except ValueError:
            rel = p.name
        warns += check_file(p, rel, cfg)
    return conv, cfg.get("version"), warns


def report(conv, version, warns, max_items=30):
    files = len({w[0] for w in warns})
    head = f"⚠ Convention: {len(warns)} cảnh báo trong {files} file (theo {conv.name}" + (f" v{version}" if version else "") + ")"
    rows = []
    for rel, ln, r in warns[:max_items]:
        where = f"{rel}:{ln}" if ln else rel
        fix = f" → {r['fix']}" if r.get("fix") else ""
        rows.append(f"- {where} [{r['id']}] {r.get('message', '')}{fix}")
    if len(warns) > max_items:
        rows.append(f"- … và {len(warns) - max_items} cảnh báo khác")
    tail = ("Cảnh báo, không chặn. Sửa theo convention, hoặc nếu cố ý giữ: thêm `convention-ignore: <ID> <lý do>` "
            "trên dòng đó và ghi vào mục Nợ kỹ thuật của hồ sơ changeset.")
    return "\n".join([head, *rows, tail])


def hook_claude():
    try:
        data = json.load(sys.stdin)
    except json.JSONDecodeError:
        return 0
    ti = data.get("tool_input") or {}
    path = ti.get("file_path") or ti.get("notebook_path")
    if not path:
        return 0
    p = Path(path)
    if not p.is_absolute():
        p = Path(data.get("cwd") or ".") / p
    conv, version, warns = run([p])
    if warns:
        msg = report(conv, version, warns)
        print(json.dumps({
            "systemMessage": msg.splitlines()[0],
            "hookSpecificOutput": {"hookEventName": "PostToolUse", "additionalContext": msg},
        }, ensure_ascii=False))
    return 0


def main(argv):
    if argv[:2] == ["--hook", "claude"]:
        return hook_claude()
    strict = "--strict" in argv
    args = [a for a in argv if a != "--strict"]
    if args:
        paths = [Path(a) for a in args]
    else:
        root = git_root(Path.cwd())
        if not root:
            print("Không phải git repo và không có file nào được chỉ định.", file=sys.stderr)
            return 0
        paths = changed_files(root)
    conv, version, warns = run(paths)
    if paths and not conv:
        print("Chưa có convention.md — copy convention.sample.md ở gốc plugin xora-dev về gốc repo thành convention.md, rồi sửa theo dự án.", file=sys.stderr)
        return 0
    if warns:
        print(report(conv, version, warns, max_items=200))
        return 1 if strict else 0
    if conv:
        print(f"✓ Convention: {len(paths)} file đạt ({conv.name}" + (f" v{version}" if version else "") + ").")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
