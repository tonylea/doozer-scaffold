---
name: contributing-best-practices
description: Use this skill whenever working on a git repository — creating branches, opening issues, submitting pull requests, writing or updating a README, or reviewing code. Triggers include starting work on a feature or fix, preparing a PR, setting up a new repository, or any task involving collaboration workflow on GitHub. This skill defines the expected workflow and quality standards for contributing to a codebase. It complements the atomic-commits and TDD skills, which govern commit format and development cycle respectively.
---

# Contributing Best Practices

This skill defines the workflow and standards for contributing to a codebase. It covers branching strategy, issue tracking, pull requests, code review, repository documentation, and repo hygiene.

## Branching Strategy: Trunk-Based Development

Use trunk-based development. All work integrates into `main` (the trunk) frequently.

**Rules — non-negotiable:**

1. **Short-lived feature branches only.** Branch from `main`, name it descriptively (e.g. `feat/add-status-filter`, `fix/null-ref-sidebar`), and merge back within 1–2 days maximum. If the work takes longer, break it into smaller deliverables. A branch that lives longer than two days is a long-lived branch. Don't let it happen.
2. **No long-lived feature branches.** Long-lived branches cause merge conflicts, code drift, and integration pain. If a feature is too large for a short-lived branch, use feature flags to merge incomplete work safely. "I'll merge it when it's done" is how branches die. Merge early, merge often.
3. **Keep the trunk deployable.** Every merge to `main` must leave the codebase in a working, releasable state. Merging broken code to `main` is not acceptable under any circumstances — not "temporarily", not "to unblock someone", not for any reason.
4. **Rebase before merging.** Before opening a PR, rebase your branch onto the latest `main` to maintain a linear history and catch conflicts early: `git fetch origin && git rebase origin/main`.
5. **Delete branches after merge.** Once a PR is merged, delete the feature branch immediately. It has served its purpose.

## Issue Tracking

Every piece of work must be linked to a GitHub issue. No issue, no work. This is not bureaucracy — it's traceability.

**Before starting any work:**

1. Check if an issue already exists for the task. If not, create one. Do this before writing code, not after.
2. The issue should clearly describe: what needs to change, why it needs to change, and acceptance criteria where applicable.
3. Use labels to categorise issues (e.g. `bug`, `enhancement`, `documentation`, `good first issue`).
4. Reference the issue number in your branch name and commit messages (e.g. `feat/add-filter-#42`, and `Closes #42` in the PR body).

## Pull Requests

A pull request is a proposal to merge your changes. It should be easy to review.

**PR quality checklist:**

1. **Descriptive title** using Conventional Commits format where appropriate (e.g. `feat(auth): add JWT refresh before expiry`).
2. **Body includes:** a summary of what changed and why, a link to the related issue (`Closes #42`), any breaking changes or migration steps, and screenshots or examples for UI changes.
3. **Small and focused.** A PR should address one concern. If you find yourself writing "also" in the description, split the PR. Not "consider splitting" — split it.
4. **All checks pass.** CI, linting, and tests must be green before requesting review. Do not request review on a failing PR. Do not merge a failing PR.
5. **Self-review first.** Read through your own diff before requesting review. Catch typos, debug statements, and unintended changes. If you wouldn't approve this diff from someone else, don't submit it.

**PR template (if the repo doesn't already have one):**

```markdown
## What

Brief description of the change.

## Why

Link to issue or explanation of motivation.

## How

High-level approach taken.

## Checklist

- [ ] Tests added/updated
- [ ] Documentation updated if needed
- [ ] No unrelated changes included
- [ ] Self-reviewed the diff
```

## Code Review

When reviewing others' code:

1. Focus on correctness, readability, and maintainability — not style preferences already covered by linters.
2. Ask questions rather than making demands. "Could this be simplified by...?" is better than "Change this."
3. Approve when satisfied. Don't block on trivial nits — leave them as non-blocking comments.
4. Respond to review feedback on your own PRs promptly. Stale PRs create merge conflicts and block others.

## Repository Documentation

Every repository should have at minimum:

**README.md** — the front door of the project. It should include:
- Project name and one-line description of what it does
- Prerequisites and setup instructions (how to get it running locally)
- Usage examples
- How to run tests
- Link to CONTRIBUTING.md if it exists
- Licence information

**CONTRIBUTING.md** — guidelines for contributors. It should cover:
- How to report bugs and request features
- Branch naming conventions
- Commit message format (reference the atomic-commits skill)
- PR process and expectations
- Code style and linting requirements

**Keep documentation current.** If you change how to set up or use the project, update the docs in the same PR. Not in a follow-up PR, not "later" — in the same PR. Stale documentation is worse than no documentation.

## Repo Hygiene

1. **Lint before committing.** Run the project's linter and fix issues before pushing. Don't rely on CI to catch formatting problems.
2. **Don't commit generated files** (build outputs, `node_modules`, `.env` files with secrets). Use `.gitignore` properly.
3. **Use issue and PR templates** if the project supports them, to maintain consistency.
4. **Do not manually tag releases.** Tagging and versioning are handled automatically by `commit-and-tag-version` in the CI pipeline. This is why correct Conventional Commits format matters — the tool parses commit types to determine the semantic version bump.
5. **Keep the issue tracker clean.** Close resolved issues, label active ones, and triage regularly.

## Integration with Other Skills

| Concern | Governed by |
|---|---|
| Commit message format and atomicity | `atomic-commits` skill |
| TDD cycle and micro-commit conventions | `tdd` skill |
| `gh` CLI operations (PRs, issues, runs) | `github` skill (steipete/github) |
| Branching, PRs, reviews, repo docs, hygiene | **This skill** |
