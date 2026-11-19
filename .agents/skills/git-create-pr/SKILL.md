---
name: git-create-pr
description: >-
  Synchronize the current branch with its base branch, resolve conflicts,
  verify changes, and create a pull request. Use when the user asks to create a
  PR, open a pull request, push for review, or ship the branch. Follow root
  AGENTS.md Git safety and BASE_BRANCH.
---

# Create Pull Request

Follow root `AGENTS.md` Git safety.

## 1. Determine Base Branch

Determine the base branch using this priority:

1. Base branch explicitly specified by the user.
2. `BASE_BRANCH` configured in the project's root `AGENTS.md`.
3. Default to `main` if neither is specified.

Verify that the selected base branch exists locally or remotely.

If the base branch cannot be found, stop and ask the user.

## 2. Check Working Tree

1. Identify the current working branch.
2. Save the current branch name.
3. Ensure the current branch is not the base branch.
4. Inspect the working tree using `git status`.
5. Ensure there are no uncommitted changes.

If uncommitted changes exist:

- Stop the workflow.
- Inform the user which files have uncommitted changes.
- Do not stash, discard, or commit them automatically.

## 3. Synchronize with Base Branch

1. Checkout the selected base branch.
2. Pull the latest changes from its remote.
3. Checkout the original working branch.
4. Rebase the current branch onto the updated base branch.

Use the appropriate remote tracking branch when pulling.

Do not rebase onto an outdated local base branch.

## 4. Resolve Conflicts

If rebase encounters conflicts:

1. Inspect the conflicted files.
2. Invoke the `git-resolve-conflict` skill to analyze and resolve conflicts.
3. Review the resolved changes.
4. Stage the resolved files.
5. Continue the rebase using `git rebase --continue`.
6. Repeat until the rebase completes successfully.

If a standalone commit is required outside the rebase process, invoke the `git-commit` skill.

Important:

- Do not create unnecessary commits for changes already included in a replayed commit.
- During rebase, use `git rebase --continue` rather than creating a new commit manually.
- If conflicts cannot be resolved safely, stop and ask the user.
- Do not automatically abort the rebase without explaining the situation.

## 5. Verify Changes

After rebase completes:

1. Verify that the current branch is the original working branch.
2. Verify that the working tree is clean.
3. Review the complete diff against the selected base branch.
4. Review the commit history.
5. Run relevant tests, lint, and type checks.
6. Ensure no unintended changes were introduced.

Do not proceed if verification fails.

## 6. Prepare Pull Request Information

Only after synchronization and verification are complete:

1. Generate a concise PR title.
2. Generate a PR description containing:

   - Summary
   - Changes
   - Testing

3. Follow the project's existing PR conventions if available.

Do not invent completed work or test results.

## 7. Push and Create Pull Request

After preparing the PR information:

1. Check whether the current branch has an upstream remote.
2. Push the current branch to its remote if necessary.
3. If rebase rewrote previously pushed commits, use `--force-with-lease` only when necessary.
4. Create the pull request using the available Git provider CLI or integration.
5. Set the selected base branch as the PR target.
6. Return the PR URL.

If the Git provider is unavailable or authentication fails, report the issue without claiming that the PR was created.

## Rules

- Always use the selected base branch as the PR target.
- Never push directly to the base branch.
- Never use `git push --force`.
- Never discard changes without understanding their purpose.
- Never claim tests passed if they were not executed.
- Never create the PR before synchronization and verification are complete.
- Never commit automatically except when required by a standalone conflict-resolution change.
- Ask for confirmation before pushing or creating a PR unless explicitly authorized.
