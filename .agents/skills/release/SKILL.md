---
name: release
description: >-
  Cut an xoradlc-plugin GitHub Release for the Go xdlc CLI: bump version,
  sync docs, verify CI, tag vX.Y.Z, and confirm multi-OS release assets.
  Use when the user asks to release, cut a version, tag a release, publish
  xdlc, or ship install/update binaries. Do not auto-tag on every BASE_BRANCH merge.
---

# Release (xoradlc-plugin)

Ship **user-facing** `xdlc` binaries via **GitHub Release** attached to tag `vX.Y.Z`.
`install.sh` / `install.ps1` / `xdlc update` download Release assets only — **no tip-of-branch fallback**.

Follow root `AGENTS.md` (Git safety + `BASE_BRANCH`). Prefer invoking `git-commit` / `git-create-pr` for commit/PR steps.

**Resolve base branch** (same priority as `git-create-pr`):

1. Base branch explicitly specified by the user.
2. `BASE_BRANCH` in root `AGENTS.md`.
3. Default `main` if neither is set.

Below, **BASE_BRANCH** means that resolved name.

## When to use

- User asks to release / tag / publish a new `xdlc` version.
- After features on a branch are merged (or ready to merge) to **BASE_BRANCH** and CI is green.

## Non-goals

- Do **not** create a tag on every merge to **BASE_BRANCH**.
- Do **not** delete a tag/release that users may already have installed.
- Do **not** publish by uploading binaries manually unless CI is broken and the user explicitly asks.
- Dev-only installs (`XDLC_PLUGIN_DIR`, local `go build`) are **not** a release.

## Current pipeline (source of truth)

| Piece | Role |
|---|---|
| `internal/config/config.go` → `Version` | Semver string embedded in the binary (`xdlc --version`, help) |
| `USER_GUIDE.md` Commands fence | Must match `xdlc` / `xdlc -h` output **exactly** (includes version on line 1) |
| `.github/workflows/xdlc-plugin-tests.yml` | PR / **BASE_BRANCH**: `go test` → build → `scripts/smoke_xdlc.py` (**3** runners: ubuntu, macOS, windows) |
| `.github/workflows/release.yml` | On push tag `v*`: cross-compile **6** GOOS/GOARCH binaries + upload Release |
| Release matrix | **3 OS × 2 arch = 6 assets**: `darwin`/`linux`/`windows` × `amd64`/`arm64` |
| Asset names | `xdlc-darwin-amd64`, `xdlc-darwin-arm64`, `xdlc-linux-amd64`, `xdlc-linux-arm64`, `xdlc-windows-amd64.exe`, `xdlc-windows-arm64.exe` |
| `install.sh` / `install.ps1` | Fetch `releases/latest` (or `XDLC_REF`) + asset for OS/arch |
| `xdlc update` | Needs a Release **newer than** installed `Version`, then replace binary + reapply setups |

Do not confuse **smoke CI (3 OS runners)** with **release binaries (6 GOOS/GOARCH)**.

## Workflow

### 0. Confirm intent + proposed version

1. Read current `Version` in `internal/config/config.go`.
2. Propose the next version using **Version bump rules** below (patch / minor / major) from the change set since the last release tag.
3. **Stop and ask the user** to chốt (accept / change) the proposed `Version` and tag `vX.Y.Z` before bumping files, committing, or tagging. Do not proceed past this gate without an explicit OK.
4. Confirm **BASE_BRANCH** (from `AGENTS.md` / user override).
5. Stop if working tree has unrelated dirty changes the user did not ask to include.

## Version bump rules (SemVer)

Tag = `v` + `Version` (e.g. code `0.2.2` ↔ tag `v0.2.2`). Choose the **highest** applicable bump for the changes since the previous release:

| Bump | Example | When |
|---|---|---|
| **patch** | `0.2.1` → `0.2.2` | Bugfix, CI/docs/chore only; no user-facing contract change |
| **minor** | `0.2.x` → `0.3.0` | New commands/flags/UX or capabilities; **backward compatible** for existing installs |
| **major** | `0.x` → `1.0.0` or `1.x` → `2.0.0` | **Breaking**: MCP server/tool rename, `~/.xdlc` schema incompatible, removed commands, install/asset layout that breaks `install` / `xdlc update` |

Notes:

- If the user already stated an exact version, use that (still show it and get OK before applying).
- Prefer **patch** when unsure and changes are clearly non-breaking; ask if ambiguous between minor and major.
- Do not skip numbers casually (e.g. jump `0.2.1` → `0.2.5`) unless the user asks.

### 1. Preconditions

1. `git fetch origin` and ensure release work is based on up-to-date **BASE_BRANCH** (merge or rebase as the user prefers; for a PR use `git-create-pr`).
2. Run locally (must pass):

```bash
go test ./...
go build -o xdlc ./cmd/xdlc   # Windows: -o xdlc.exe
python3 scripts/smoke_xdlc.py
```

3. If a PR is open, wait until **ubuntu / macOS / Windows** smoke jobs are green before tagging.

### 2. Bump version

1. Set `Version` in `internal/config/config.go` to the new semver **without** the `v` prefix (e.g. `"0.2.2"`).
2. Sync `USER_GUIDE.md` Commands fence first line to:

```text
xdlc <Version>: Xora DLC for Claude Code and Kiro CLI
```

   Prefer regenerating from the built binary:

```bash
go build -o xdlc ./cmd/xdlc
./xdlc -h
# replace the ```text ... ``` Commands block in USER_GUIDE.md with that exact output
```

3. Grep for the **old** version string; update any leftover user-facing copies (help is driven by `config.Version` — fence is the usual drift).
4. Commit the bump (invoke `git-commit` or equivalent). Message example: `chore(xdlc): bump version to 0.2.2`.

### 3. Land on BASE_BRANCH

1. If not already on **BASE_BRANCH**: open/merge PR into **BASE_BRANCH** (`git-create-pr` if needed).
2. Do **not** tag from a feature branch that is not yet on **BASE_BRANCH** unless the user explicitly wants a pre-release from that commit.
3. After merge: `git checkout <BASE_BRANCH> && git pull`.

### 4. Tag and push

Only after **BASE_BRANCH** contains the version bump and CI on **BASE_BRANCH**/PR was green:

```bash
git tag vX.Y.Z          # must match Version, with leading v
git push origin vX.Y.Z
```

Rules:

- Tag format: `v` + three numeric components (`v1.2.3`).
- Tag the commit that has the bumped `Version`.
- Ask before `git push` of the tag unless the user already authorized push/release.
- Never `git push --force` a release tag.

### 5. Watch release CI

1. Open the Actions run for workflow **release** on the new tag (`.github/workflows/release.yml`).
2. Confirm it builds all **6** GOOS/GOARCH assets and creates the GitHub Release.
3. Verify Release page lists exactly the expected asset names (see table above).
4. Spot-check:

```bash
# version on Release should match
curl -fsSL -o /tmp/xdlc-check "https://github.com/TechX-Corp/xoradlc-plugin/releases/download/vX.Y.Z/xdlc-$(uname -s | tr A-Z a-z)-$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')"
chmod +x /tmp/xdlc-check && /tmp/xdlc-check --version
```

### 6. Post-release checks (optional but recommended)

1. Fresh install path (no `XDLC_PLUGIN_DIR`):

```bash
curl -fsSL "https://raw.githubusercontent.com/TechX-Corp/xoradlc-plugin/<BASE_BRANCH>/install.sh" | sh
xdlc --version   # expect new Version
```

   (`install.sh` on **BASE_BRANCH** still points users at `releases/latest`.)

2. From an older install: `xdlc update` should pull the newer Release (needs network + existing older binary).
3. Report Release URL + tag + asset list to the user.

## Failure / recovery

| Problem | Action |
|---|---|
| Smoke/CI red | Fix on a branch; **do not tag** |
| Tag pushed but release workflow failed | Fix workflow on **BASE_BRANCH**, delete **failed** tag/release only if nothing useful uploaded and user agrees, retag same version on the fixed commit — or bump patch |
| Wrong Version in binary | Bump patch, new tag; do not rewrite history of a shipped tag |
| User asks to yank a bad release | Prefer publishing a fixed newer tag; deleting tags/releases is last resort and must be explicit |

## Rules

- One human decision to ship → one `vX.Y.Z` tag → CI publishes **6** assets.
- Always propose a version from the bump rules and **get user confirmation** before bump / tag / push.
- `Version` in code and tag semver must match (`0.2.2` ↔ `v0.2.2`).
- Never claim a release succeeded without checking the Release assets (or the workflow conclusion).
- Never auto-tag from routine PR merges into **BASE_BRANCH**.
- Prefer tagging the release commit **on BASE_BRANCH**; tags off-branch fail the release workflow gate (see below).
- Dev overrides (`XDLC_PLUGIN_DIR`, `XDLC_REF`, `XDLC_API_URL`) are for local/CI tests, not for end users.

## Tag from a non-BASE_BRANCH commit

`.github/workflows/release.yml` still **starts** on any `push` of tags matching `v*`, but the first job step **fails** unless the tagged commit is an **ancestor of `origin/<BASE_BRANCH>`** (`git merge-base --is-ancestor HEAD origin/<BASE_BRANCH>`).

So if someone tags a feature-branch commit and pushes `vX.Y.Z`:

- Workflow is triggered, then **exits before build/upload**.
- No Release assets are published from that tag.

`env.BASE_BRANCH` in `release.yml` must stay in sync with `BASE_BRANCH` in root `AGENTS.md`.

Default policy: merge to **BASE_BRANCH** first, then tag.