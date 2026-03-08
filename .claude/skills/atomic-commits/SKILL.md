---
name: atomic-commits
description: Use this skill whenever you are about to make a git commit, or when squashing TDD micro-commits at the end of a development cycle. Triggers include running `git commit`, preparing a commit message, or completing a TDD cycle squash. This skill defines the required commit format and quality checks. During active TDD development, the TDD skill governs micro-commit conventions — this skill takes over at squash time and for all non-TDD commits.
---

# Atomic Commits

Every commit in the shared history must be atomic: the smallest complete, meaningful change that leaves the codebase in a working state.

## Rules

These are not guidelines. Do not rationalise exceptions.

1. **One logical change per commit.** A commit does one thing — a feature, a fix, a refactor, a documentation update. If the commit message needs "and" to describe unrelated changes, stop and split it. Thinking "it's easier to commit these together"? That's the problem. Split it.
2. **Codebase stays functional.** After the commit, the code builds and all tests pass. The commit is safe to revert, cherry-pick, or bisect against. A commit that breaks the build is not atomic — it's damage.
3. **No mixed concerns.** Formatting, logic, dependencies, and documentation go in separate commits. "While I was in there" changes go in their own commit.

## Commit Message Format

Use Conventional Commits (v1.0.0):

```
type(scope?): subject

body?

footer?
```

**Types:** `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `perf`, `ci`, `build`

- Subject line: under 72 characters, explains *why* (the diff shows *what*).
- Body: additional context when the subject is insufficient.
- Footer: `BREAKING CHANGE:` or `!` after type/scope for breaking changes.

## Integration with TDD Skill

When working in a TDD cycle, two commit modes apply:

| Context | Commit format | Governed by |
|---|---|---|
| TDD micro-commits (`green:`, `refactor:`, `fix:`) | Lightweight prefixed messages | TDD skill |
| TDD squash at end of cycle | Conventional Commits, atomic | **This skill** |
| All non-TDD commits | Conventional Commits, atomic | **This skill** |

At TDD squash time, the micro-commit history collapses into one or more atomic commits. The squashed message must follow Conventional Commits format and describe the completed behaviour, not the TDD steps that produced it.

**Example — TDD squash:**
```
# Before (micro-commits from TDD cycle):
green: return empty list when no items
refactor: extract validation helper
green: filter items by status
green: raise ValueError on invalid status

# After (atomic commit for shared history):
feat(items): add status filtering with validation

Items can be filtered by status (active, archived, pending).
Invalid status values raise ValueError listing valid options.
```

## Pre-Commit Check

**Run this before every commit. Not optional.**

1. Exactly one logical change?
2. Build passes, tests pass?
3. No mixed concerns (formatting + logic, etc.)?
4. Message follows Conventional Commits format?
5. Subject explains intent, not just mechanics?

If any answer is no, do not commit. Split or revise first. Do not commit with the intention of "fixing it in the next commit."
