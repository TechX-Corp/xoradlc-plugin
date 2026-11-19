---
name: git-resolve-conflict
description: >-
  Analyze and resolve Git merge or rebase conflicts safely without blind
  ours/theirs. Use when there are merge or rebase conflicts, conflict markers,
  or the user asks to resolve git conflicts. Also used from git-create-pr during
  rebase. Follow root AGENTS.md Git safety.
---

# Resolve Git Conflicts

Follow root `AGENTS.md` Git safety.

## Workflow

1. Inspect Git status and identify conflicted files.
2. Understand the current branch and target branch.
3. Read the conflicting changes from both sides.
4. Understand the intended behavior of each change.
5. Resolve conflicts while preserving valid functionality.
6. Remove conflict markers.
7. Review the final diff.
8. Run relevant tests and checks.
9. Report the resolved files and any remaining issues.

## Rules

- Never blindly accept ours or theirs for all conflicts.
- Never discard changes without understanding their purpose.
- Never use destructive Git commands such as `git reset --hard` unless explicitly authorized.
- Do not commit automatically after resolving conflicts (unless the caller skill requires `git rebase --continue` only).
- Ask for clarification when the intended behavior cannot be determined.
