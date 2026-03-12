---
name: acceptance_test_coverage
description: Acceptance tests must cover every numbered SPEC criterion. Partial coverage equals failed coverage. Verify before merge.
type: feedback
---

## Rule

Every numbered acceptance criterion in a SPEC must have a corresponding test function. "Covered by unit tests elsewhere" is not sufficient for acceptance criteria — they require end-to-end acceptance tests.

## Pre-Merge Verification

Before accepting merge_ready, the coordinator must:
1. Read SPEC Section 5 (or equivalent acceptance criteria section)
2. List every numbered criterion
3. For each, identify the specific test function name
4. Any criterion without a named acceptance test = reject merge_ready

## Stage 3b Example (2026-03-11)

Builder wrote acceptance tests but missed 7 of 25 criteria:
- No test for prompt variant group collapse (#3)
- No test for Terraform variant group resolution (#8, #9)
- No test for path conflict verification (#15)
- Partial tests for gitignore (#16), devcontainer (#17), existing test migration (#10)

The user standard: partial coverage is failed coverage. Every gap is a rejection.
