# Project Instructions

Root = **orchestration + cross-cutting policy** for **xoradlc-plugin**. Chi tiết skill → `.agents/skills/`. Không duplicate dài vào file này.

## 1. Project Overview

Repo ships:

- Go CLI `xdlc` (`cmd/xdlc`, `internal/*`) + embedded pack (`roles/`, `host-agents/`, `standards/`)
- Installers `install.sh` / `install.ps1` (GitHub Release assets only)
- Smoke `scripts/smoke_xdlc.py` + workflows under `.github/workflows/`
- Separate tree `aidlc-plugin/` (AI-DLC) — **not** used by `xdlc`

User docs: `USER_GUIDE.md`. Technical map: `README.md`.

Trước khi đổi code: hiểu layout và conventions hiện có. Không assume pattern — kiểm tra codebase.

## 2. Principles

- Minimal, focused changes; không refactor ngoài scope.
- Prefer pattern / helper có sẵn; tránh abstraction và dependency thừa.
- User-facing CLI/TUI: giữ parity trừ khi user yêu cầu đổi UX.
- Preserve backward compatibility (token/`~/.xdlc` schema, MCP name `xora-dlc`) trừ khi user yêu cầu khác.

## 3. Project Structure

```text
xoradlc-plugin/
├── AGENTS.md
├── USER_GUIDE.md
├── README.md
├── .agents/skills/     # git-*, release, …
├── cmd/xdlc/
├── internal/
├── roles/              # catalog + role AGENTS.md
├── host-agents/        # kiro / claude manifests (+ Claude hooks)
├── standards/          # shared guidance (e.g. MCP routing)
├── scripts/smoke_xdlc.py
├── install.sh / install.ps1
├── .github/workflows/
└── aidlc-plugin/
```

## 4. How to work

Một **Main Agent** owns toàn bộ task: hiểu request → đọc skills khớp → implement → verify → report.

- Break down khi phức tạp; parallel chỉ khi độc lập (không sửa cùng file cùng lúc).
- Verification cuối và báo cáo cho user luôn thuộc Main Agent.

## 5. Skills

Path: `.agents/skills/`

Trước khi implement: check skill khớp task → load và follow. Không load skill không liên quan.

Git / release: dùng skill khớp (`git-commit`, `git-create-pr`, `git-resolve-conflict`, `release`) khi commit / PR / conflict / cắt bản.

## 6. Git

- `BASE_BRANCH`: `main` (user có thể override khi gọi skill / tạo PR)
- Dùng skill `git-*` / `release` khớp (§5)
- Không discard uncommitted changes, destructive git, push thẳng protected branch, hay force push — trừ khi user authorize
- **Commit:** khi user đã yêu cầu commit → làm luôn, không hỏi confirm lại. Prompt không chỉ định file → mặc định `git add .` (vẫn loại secret / `.env` / binary `xdlc`/`xdlc.exe`)
- **Push / tạo PR / push tag release:** hỏi confirm trừ khi user đã authorize rõ

## 7. Build, test, release

Local verify (trước khi claim xong / trước khi tag):

```bash
go test ./...
go build -o xdlc ./cmd/xdlc   # Windows: -o xdlc.exe
python3 scripts/smoke_xdlc.py
```

- PR / `BASE_BRANCH` CI: `.github/workflows/xdlc-plugin-tests.yml` (smoke trên ubuntu, macOS, windows).
- Release: chỉ khi push tag `v*` → `.github/workflows/release.yml` cross-compile **6** GOOS/GOARCH assets (darwin/linux/windows × amd64/arm64). Job **fail sớm** nếu tagged commit không nằm trên lịch sử `BASE_BRANCH`. Chi tiết → skill `release`.
- `install` / `xdlc update` lấy binary từ GitHub Release — **không** fallback tip of `BASE_BRANCH`.
- Bump version: `internal/config/config.go` → `Version`; sync fence Commands trong `USER_GUIDE.md` với output `xdlc -h`.

## 8. Verification & communication

Sau khi đổi code: review diff; chạy test/smoke liên quan; report check nào chưa chạy được. Không claim test passed nếu chưa chạy thành công.

Task phức tạp: nói approach ngắn trước khi sửa.

Khi xong, report: changed / decisions / checks / limitations.

Hỏi lại khi thiếu requirement ảnh hưởng material tới implement.

## 9. Security

- Không expose secrets / tokens; không commit `.env` hay credential
- Không commit binary build (`/xdlc`, `/xdlc.exe` — đã gitignore)
- OAuth/token files dưới `~/.xdlc` là máy user — không copy vào repo
