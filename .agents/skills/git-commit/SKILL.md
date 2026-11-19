---
name: git-commit
description: >-
  Analyze Git changes and create commits following project conventions
  (Conventional Commits if none). Use when the user asks to commit, create a
  commit, git commit, or stage and save work. Follow root AGENTS.md Git safety.
---

# Git Commit

Follow root `AGENTS.md` Git safety. **Không hỏi confirm** trước khi commit khi user đã yêu cầu commit.

## Workflow

1. Inspect branch và working tree (`git status`, staged + unstaged diff).
2. Hiểu mục đích thay đổi; soạn message ngắn (Conventional Commits nếu project không có convention riêng).
3. **Stage:**
   - Prompt chỉ định file/path cụ thể → stage đúng phạm vi đó.
   - Prompt **không** mô tả chi tiết file nào → mặc định `git add .` (sau khi đã xem status/diff; vẫn **bỏ** secret / `.env` / credential khỏi commit).
4. Commit (HEREDOC cho message nếu shell yêu cầu).
5. `git status` sau commit để xác nhận.

## Rules

- Never commit secrets or environment files (unstage / exclude nếu `git add .` bắt nhầm).
- Never amend unless explicitly requested.
- Do not commit if there are no relevant changes.
- Do not ask for confirmation again when the user already asked to commit.
- Never force push; never push unless the user asked to push (commit ≠ push).
