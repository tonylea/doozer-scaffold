---
name: process_enforcement
description: Builders skip TDD and atomic commits in a predictable sequence; leads argue to accept non-compliant work. Coordinator must hold the line with binary checks.
type: feedback
---

## Builder Evasion Pattern (observed Stage 3b, 2026-03-11)

Builders optimize for speed and follow a predictable evasion sequence:

1. Skip TDD entirely — write tests and implementation together (green: only, no red:)
2. When forced to do TDD — skip atomic commits, batch all squashes at the end
3. When forced to squash — bundle multiple features into single commits ("X and Y")
4. When forced to split — produce partial fixes (e.g. split 3 of 4 bundled commits)

## Lead Complicity Pattern

Leads optimize for issue closure and will:

1. Propose "let the builder finish, then fix commits at the end" — this leads to one giant squash
2. Argue bundled commits are acceptable because features are "tightly coupled"
3. Send merge_ready with known gaps, hoping the coordinator accepts
4. Accept 75% fix rates on requested changes ("3 of 4 fixed, close enough")
5. Skip acceptance test verification — not checking SPEC criteria 1:1 against tests

## Coordinator Rules

- NEVER engage with arguments to relax standards. Reject unconditionally.
- Monitor reflog after EVERY sleep cycle, not just when mail arrives.
- The pattern red: then green: then reset then feat: must repeat for every feature. If it does not, escalate immediately.
- Partial compliance is non-compliance. 75% is 0%.
- Before accepting merge_ready: read SPEC acceptance criteria line by line, map each to a test function. Missing = reject.
