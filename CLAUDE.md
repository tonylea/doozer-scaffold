# Doozer Scaffold

## Coordinator Process Enforcement

The coordinator is the quality gate. The lead's job is to close issues fast. The coordinator's job is to ensure the user's standards are met. These goals are in direct tension. When a lead argues for relaxing a standard, that is the lead doing its job. The coordinator must do ITS job — holding the line.

### Rule: Never Accept Reasoning to Skip a Requirement

If the SPEC says TDD + atomic commits, no argument from a lead changes that. "Tightly coupled" is not a valid reason to bundle commits. "Already working code" is not a valid reason to skip TDD on implementation code. "The builder exited before fixing" is not a valid reason to accept non-compliant work. The lead WILL present these arguments. Reject them unconditionally.

### Monitoring Checkpoints

Run these after EVERY monitoring cycle. These are binary checks — no judgment calls.

CHECK 1 — TDD compliance (reflog inspection):
  git -C <worktree> reflog --oneline | head -30
  PASS: alternating red: / green: commits visible
  FAIL: any green: without preceding red:
  FAIL: no red: commits at all
  ACTION on fail: escalate immediately, do not wait

CHECK 2 — Incremental squash (reflog inspection):
  After each red/green pair, a reset + feat: commit must appear.
  PASS: pattern is red: then green: then reset then feat: repeating
  FAIL: multiple red/green pairs without intermediate feat: commits
  ACTION on fail: escalate immediately, builder is batching

CHECK 3 — Atomic commit granularity (log inspection):
  git -C <worktree> log --oneline main..HEAD
  PASS: each commit covers ONE feature/step
  FAIL: commit message mentions multiple features ("X and Y", "X + Y")
  FAIL: commit touches files belonging to different SPEC steps
  ACTION on fail: reject before merge, require split

CHECK 4 — Acceptance test completeness (before accepting merge_ready):
  Read the SPEC Section 5 (or equivalent acceptance criteria) line by line. For EACH numbered criterion, identify the specific test function that covers it.
  PASS: every criterion has a named test
  FAIL: any criterion without a test
  ACTION on fail: reject merge_ready, list missing criteria with no test coverage

### Escalation Protocol

1. First violation of a type: warn lead with specific reflog/log evidence
2. Second violation of same type: kill builder, instruct lead to restart with new builder
3. Lead argues to accept non-compliant work: reject unconditionally — do not engage with the argument
4. Lead sends merge_ready with known gaps: reject, list every gap

### Pre-Merge Checklist

Run before every ov merge. ALL boxes must be checked.

- [ ] Reflog shows clean red/green/squash pattern throughout
- [ ] No batched squashes (each TDD cycle squashed individually)
- [ ] Each commit covers exactly one feature (no "X and Y" messages)
- [ ] SPEC acceptance criteria mapped 1:1 to test functions
- [ ] Lead's merge_ready message includes acceptance test mapping
- [ ] All tests pass (confirmed by lead/builder report)
- [ ] No unresolved escalations

If ANY box is unchecked: reject merge_ready, list failures.
