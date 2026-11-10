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

Luật `forbid` chỉ báo trên dòng đã đổi so với HEAD (file mới / ngoài git: mọi dòng), để code cũ
không làm ngập cảnh báo. Luật cấp file (require, filename, max_lines) vẫn xét cả file.

Bỏ qua có chủ đích: thêm `convention-ignore: <ID> <lý do>` trên chính dòng đó.

Cách chạy:
  python3 convention_sensor.py [file ...]      # không có file → các file đổi trong git (staged + chưa staged + mới)
  python3 convention_sensor.py --staged        # chỉ phần đã staged (dùng cho git pre-commit)
  python3 convention_sensor.py --full [file …] # kiểm mọi dòng, kể cả code cũ (rà toàn bộ)
  python3 convention_sensor.py --hook claude   # hook PostToolUse của Claude Code: đọc JSON từ stdin
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
HUNK = re.compile(r"^@@ -\S+ \+(\d+)(?:,(\d+))? @@", re.M)


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
            raise ValueError(f"convention.md: khối luật JSON sai cú pháp ({e})")
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


def read_text(path, staged):
    """Nội dung file; --staged thì đọc bản trong index để khớp số dòng của `git diff --cached`."""
    try:
        if staged and (root := git_root(path.parent)):
            res = subprocess.run(["git", "-C", str(root), "show", f":{path.relative_to(root.resolve()).as_posix()}"],
                                 capture_output=True)
            if res.returncode == 0:
                return res.stdout.decode("utf-8")
        return path.read_text(encoding="utf-8")
    except (OSError, UnicodeDecodeError, ValueError):
        return None


def check_file(path, rel, cfg, lines_changed=None, staged=False):
    if match(rel, cfg.get("exclude")):
        return []
    text = read_text(Path(path), staged)
    if text is None:
        return []
    lines = text.splitlines()
    out = []
    for r in cfg["rules"]:
        if not match(rel, r.get("files", ["**/*"])) or match(rel, r.get("exclude")):
            continue
        kind, hit = r.get("type"), []
        if kind == "forbid":
            hit = [i for i, ln in enumerate(lines, 1)
                   if (lines_changed is None or i in lines_changed)
                   and r["_re"].search(ln) and r["id"] not in IGNORE.findall(ln)]
        elif kind == "require" and not r["_re"].search(text):
            hit = [0]
        elif kind == "filename" and not r["_re"].search(Path(rel).name):
            hit = [0]
        elif kind == "max_lines" and len(lines) > int(r.get("max", 500)):
            hit = [0]
        out += [(rel, ln, r) for ln in hit]
    return out


def git(root, *args):
    return subprocess.run(["git", "-C", str(root), *args], capture_output=True, text=True)


def changed_lines(path, staged):
    """Số dòng đã đổi (set), hoặc None = xét mọi dòng (ngoài git, file mới, repo chưa có commit)."""
    root = git_root(path.parent)
    if not root or git(root, "ls-files", "--error-unmatch", str(path)).returncode != 0:
        return None
    base = ["--cached"] if staged else ["HEAD"]
    res = git(root, "diff", "-U0", "--no-color", "--no-ext-diff", *base, "--", str(path))
    if res.returncode != 0:  # chưa có HEAD
        return None
    out = set()
    for start, count in HUNK.findall(res.stdout):
        n = int(count) if count else 1
        out.update(range(int(start), int(start) + n))
    return out


def changed_files(root, staged=False):
    files = set()
    cmds = ([["diff", "--name-only", "--cached", "--diff-filter=d"]] if staged else
            [["diff", "--name-only"], ["diff", "--name-only", "--cached"],
             ["ls-files", "--others", "--exclude-standard"]])
    for args in cmds:
        res = subprocess.run(["git", "-C", str(root), *args], capture_output=True, text=True)
        files |= {f for f in res.stdout.splitlines() if f}
    return [root / f for f in sorted(files) if (root / f).is_file()]


def run(paths, full=False, staged=False):
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
        warns += check_file(p, rel, cfg, None if full else changed_lines(p, staged), staged)
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
    tail = ("Chỉ cảnh báo, không chặn. Agent sửa vi phạm trong phần code vừa sinh / sửa; vi phạm ở code cũ thì liệt kê "
            "cho Developer. Cố ý giữ: Developer đồng ý thì thêm `convention-ignore: <ID> <lý do>` trên dòng đó "
            "và ghi vào mục Nợ kỹ thuật của hồ sơ changeset.")
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
    full, staged = ("--full" in argv), ("--staged" in argv)
    args = [a for a in argv if a not in ("--full", "--staged")]
    if args:
        paths = [Path(a) for a in args]
    else:
        root = git_root(Path.cwd())
        if not root:
            print("Không phải git repo và không có file nào được chỉ định.", file=sys.stderr)
            return 0
        paths = changed_files(root, staged)
    conv, version, warns = run(paths, full, staged)
    if paths and not conv:
        print("Chưa có convention.md — copy convention.sample.md ở gốc plugin xora-dev về gốc repo thành convention.md, rồi sửa theo dự án.", file=sys.stderr)
        return 0
    if warns:
        print(report(conv, version, warns, max_items=200))
        return 0
    if conv:
        print(f"✓ Convention: {len(paths)} file đạt ({conv.name}" + (f" v{version}" if version else "") + ").")
    return 0


if __name__ == "__main__":
    # Chỉ cảnh báo, không chặn: mọi lỗi (convention.md sai, regex sai, …) thành một cảnh báo, luôn exit 0.
    try:
        main(sys.argv[1:])
    except Exception as e:
        msg = f"⚠ Sensor convention lỗi, bỏ qua lần kiểm này: {e}"
        if sys.argv[1:3] == ["--hook", "claude"]:
            print(json.dumps({"systemMessage": msg}, ensure_ascii=False))
        else:
            print(msg, file=sys.stderr)
    sys.exit(0)
