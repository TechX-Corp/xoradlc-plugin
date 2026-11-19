---
name: git-create-pr
description: >-
  Sync with base, rebase, resolve conflicts, verify, then (when shipping) propose
  semver, bump Version in a dedicated commit after user confirms, and open a PR.
  Merge triggers CI auto-tag and Release binaries. Use for PR, ship, or review.
  Follow root AGENTS.md Git safety.
---

# Create Pull Request (and ship xdlc)

Follow root `AGENTS.md` Git safety and §7 (auto-tag on **BASE_BRANCH** after smoke ×3).

**Resolve base branch:**

1. User-specified base branch.
2. `BASE_BRANCH` in root `AGENTS.md`.
3. Default `main`.

Below, **BASE_BRANCH** means that resolved name.

## 1. Determine Base Branch

Verify **BASE_BRANCH** exists. If not, stop and ask.

## 2. Check Working Tree

1. Save the current branch (must not be **BASE_BRANCH**).
2. Feature work should be **committed** before sync; the bump commit (§6) is added later.

Uncommitted changes the user did not ask to include → stop and list files.

## 3. Synchronize with Base Branch

Checkout **BASE_BRANCH**, pull. Checkout working branch, **rebase** onto **BASE_BRANCH**.

On conflicts → `git-resolve-conflict`, then `git rebase --continue` until done.

## 4. Resolve Conflicts

Inspect conflicted files, resolve, stage, continue rebase. Standalone fix outside rebase → `git-commit` skill.

## 5. Verify Changes (before bump)

1. Confirm branch and **clean** working tree (no pending bump yet).
2. Review full diff and history vs **BASE_BRANCH**.
3. Run checks appropriate to the diff:

```bash
go test ./...
```

When diff touches shipped surfaces (§6 triggers) or user asked to ship, also:

```bash
go build -o xdlc ./cmd/xdlc   # Windows: -o xdlc.exe
python3 scripts/smoke_xdlc.py
```

Do not proceed if verification fails.

## 6. Version bump — last step before the PR

Run **only after** §3–§5 succeed. This is the **last** commit on the branch before push/PR.

### 6.0 When to run

**Run** when any of:

1. User asked to **ship**, **release**, **cut a version**, or **publish** `xdlc`.
2. Diff vs **BASE_BRANCH** touches: `cmd/xdlc/`, `internal/`, `roles/`, `host-agents/`, `standards/`, `install.sh`, `install.ps1`.

**Skip** when all of:

- No ship/release intent.
- Diff is only e.g. `aidlc-plugin/`, docs, or `.agents/` with no CLI/pack change.
- **Or** branch already bumps `Version` vs **BASE_BRANCH** and `USER_GUIDE.md` Commands fence matches `xdlc -h` (fix fence only if drift).

### 6.1 Propose semver

1. Read `Version` in `internal/config/config.go`.
2. Optional: `git fetch --tags` and compare to latest `v*.*.*`.
3. Propose next **SemVer** (highest applicable since last tag):

| Bump | Example | When |
|---|---|---|
| **patch** | `0.2.1` → `0.2.2` | Bugfix, small UX; no breaking contract |
| **minor** | `0.2.x` → `0.3.0` | New capability; backward compatible |
| **major** | `0.x` → `1.0.0` | Breaking: MCP rename, `~/.xdlc` schema, removed commands, install layout |

4. **Stop and ask the user** to accept or change. No file edits until confirmed (unless they already gave the exact version — show it once).

### 6.2 Apply bump commit

After confirmation:

1. Update `Version` in `internal/config/config.go`.
2. Sync `USER_GUIDE.md` Commands block from `./xdlc -h` (rebuild if needed).
3. Grep for stale version strings.
4. Commit bump only: `chore(xdlc): bump version to X.Y.Z` (`git-commit`).

Re-run after bump:

```bash
go test ./...
python3 scripts/smoke_xdlc.py
```

Do **not** push a release tag; merge → CI `tag-release` → `release.yml`.

## 7. Prepare Pull Request Information

Title, body: Summary, Changes, Test plan.

If §6 ran, note: merge publishes `vX.Y.Z` (CI auto-tag + 6 Release assets); checklist PR smokes → merge → watch Actions.

Do not invent CI results.

## 8. Push and Create Pull Request

Push (`--force-with-lease` if rebase rewrote history). `gh pr create` to **BASE_BRANCH**. Self-assign `@me` unless user said otherwise. Return PR URL.

If user authorized push/PR in the same request, do not ask again.

## After merge (optional)

If user asks whether release landed: **BASE_BRANCH** CI → job `tag release if new version` → workflow `release` → GitHub Release page and six assets (`AGENTS.md` §7).

## Rules

- Never push to **BASE_BRANCH** directly.
- Never `git push --force` (only `--force-with-lease` after rebase).
- Never bump without user confirmation (§6.1).
- Order: **rebase → conflicts → verify → bump (last commit) → PR**.
- Never open PR before §5 passes and §6 completes or is skipped.
- Shipping: one bump commit on top → merge once → auto release.