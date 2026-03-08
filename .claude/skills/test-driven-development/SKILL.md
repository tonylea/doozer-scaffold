---
name: test-driven-development
description: Use when implementing any feature or bugfix, before writing implementation code. Triggers include any mention of 'TDD', 'test-driven', 'red green refactor', 'write tests first', or when asked to build functionality incrementally with tests leading the design. Also use when asked to develop something 'the Kent Beck way' or when micro-commits during development are mentioned. This skill covers the full TDD cycle including commit discipline — micro-commits at each green phase, squashed into atomic commits at the end of a development cycle.
---

# Test-Driven Development (TDD)

## Overview

Write the test first. Watch it fail. Write minimal code to pass. Commit. Refactor. Commit. Repeat.

**Core principle:** If you didn't watch the test fail, you don't know if it tests the right thing.

**Violating the letter of the rules is violating the spirit of the rules.**

## When to Use

**Always:**
- New features
- Bug fixes
- Refactoring
- Behaviour changes

**Exceptions (ask your human partner):**
- Throwaway prototypes
- Generated code
- Configuration files

Thinking "skip TDD just this once"? Stop. That's rationalisation.

## The Iron Law

```
NO PRODUCTION CODE WITHOUT A FAILING TEST FIRST
```

Write code before the test? Delete it. Start over.

**No exceptions:**
- Don't keep it as "reference"
- Don't "adapt" it while writing tests
- Don't look at it
- Delete means delete

Implement fresh from tests. Period.

## Red-Green-Refactor

### RED — Write Failing Test

Write one minimal test showing what should happen.

<Good>
```typescript
test('retries failed operations 3 times', async () => {
  let attempts = 0;
  const operation = () => {
    attempts++;
    if (attempts < 3) throw new Error('fail');
    return 'success';
  };

  const result = await retryOperation(operation);

  expect(result).toBe('success');
  expect(attempts).toBe(3);
});
```
Clear name, tests real behaviour, one thing
</Good>

<Bad>
```typescript
test('retry works', async () => {
  const mock = jest.fn()
    .mockRejectedValueOnce(new Error())
    .mockRejectedValueOnce(new Error())
    .mockResolvedValueOnce('success');
  await retryOperation(mock);
  expect(mock).toHaveBeenCalledTimes(3);
});
```
Vague name, tests mock not code
</Bad>

**Requirements:**
- One behaviour
- Clear name describing the behaviour, not the implementation
- Real code (no mocks unless unavoidable)

#### Structural Red vs Behavioural Red

The RED phase has two sub-steps. Both matter.

**Step 1a — Structural Red:** The test references a function, class, or module that does not yet exist. Run the test. It fails with an import error, `NameError`, or equivalent. This confirms the test is wired up correctly.

**Step 1b — Behavioural Red:** Create the minimal skeleton (function signature, class stub) so the test *runs* but fails on the assertion. This confirms the test actually checks intended behaviour, not just that something exists.

Step 1a catches wiring problems. Step 1b catches meaningless tests.

### Verify RED — Watch It Fail

**MANDATORY. Never skip.**

```bash
npm test path/to/test.test.ts
```

Confirm:
- Test fails (not errors — after Step 1b)
- Failure message is expected
- Fails because feature missing (not typos)

**Test passes?** You're testing existing behaviour. Fix test.

**Test errors?** Fix error, re-run until it fails correctly.

### GREEN — Minimal Code

Write the simplest code to pass the test. Nothing more.

<Good>
```typescript
async function retryOperation<T>(fn: () => Promise<T>): Promise<T> {
  for (let i = 0; i < 3; i++) {
    try {
      return await fn();
    } catch (e) {
      if (i === 2) throw e;
    }
  }
  throw new Error('unreachable');
}
```
Just enough to pass
</Good>

<Bad>
```typescript
async function retryOperation<T>(
  fn: () => Promise<T>,
  options?: {
    maxRetries?: number;
    backoff?: 'linear' | 'exponential';
    onRetry?: (attempt: number) => void;
  }
): Promise<T> {
  // YAGNI
}
```
Over-engineered
</Bad>

Don't add features, refactor other code, or "improve" beyond the test.

If the test expects `42`, returning `42` is legitimate. It feels wrong, but the next test will force you to generalise. Trust the process.

### Verify GREEN — Watch It Pass

**MANDATORY.**

```bash
npm test path/to/test.test.ts
```

Confirm:
- Test passes
- Other tests still pass
- Output pristine (no errors, warnings)

**Test fails?** Fix code, not test.

**Other tests fail?** Fix now.

**Once the test passes → COMMIT.**

```
git add -A && git commit -m "green: <short description of behaviour>"
```

### REFACTOR — Clean Up

After green only:
- Remove duplication
- Improve names
- Extract helpers
- Simplify conditionals

Keep tests green. Don't add behaviour. Run tests after each change.

**If a refactor breaks a test** and the fix is not immediately obvious (within ~60 seconds), revert to the last green commit rather than debugging. The micro-commit discipline makes this cheap.

**Once refactoring is complete and tests are still green → COMMIT.**

```
git add -A && git commit -m "refactor: <what was improved>"
```

### Repeat

Next failing test for the next behaviour.

**Test progression order:** Start with degenerate cases (empty input, None, zero), then happy paths, then edge cases, then error handling.

## Commit Discipline

### During Development — Micro-commits

Commit at every green state. This gives you a safety net to revert to at any point. Commit messages during this phase are lightweight:

| Prefix       | When                                    |
|--------------|-----------------------------------------|
| `green:`     | A new test passes for the first time    |
| `refactor:`  | Code restructured, all tests still pass |
| `fix:`       | A broken test corrected during red phase|

Example sequence:
```
green: return empty list when no items
refactor: extract item validation to helper
green: filter items by status
green: raise ValueError on invalid status
refactor: replace if-chain with dict lookup
```

### End of Development Cycle — Squash to Atomic Commit

A "development cycle" ends when you have a complete, coherent unit of functionality — a feature, a bugfix, a behaviour. At that point, squash the micro-commits into one or more atomic commits that make sense in the shared history.

**At squash time, switch to the `atomic-commits` skill** for commit message format and quality checks. The squashed commits must follow Conventional Commits (v1.0.0) format and describe the completed behaviour, not the TDD steps that produced it.

```bash
# Interactive rebase from the commit before you started this cycle
git rebase -i <commit-before-cycle>

# Squash all micro-commits into one (or a few logical) commits
# Write a proper commit message describing the completed behaviour
```

**Example:**

Before squash (local history):
```
green: return empty list when no items
refactor: extract item validation to helper
green: filter items by status
green: raise ValueError on invalid status
refactor: replace if-chain with dict lookup
green: add pagination support
```

After squash (shared history):
```
feat: add item filtering with status validation and pagination

Items can now be filtered by status (active, archived, pending).
Invalid status values raise ValueError with supported options listed.
Results are paginated with configurable page size (default 20).
```

## Good Tests

| Quality | Good | Bad |
|---------|------|-----|
| **Minimal** | One thing. "and" in name? Split it. | `test('validates email and domain and whitespace')` |
| **Clear** | Name describes behaviour | `test('test1')` |
| **Shows intent** | Demonstrates desired API | Obscures what code should do |

## Why Order Matters

**"I'll write tests after to verify it works"**

Tests written after code pass immediately. Passing immediately proves nothing:
- Might test wrong thing
- Might test implementation, not behaviour
- Might miss edge cases you forgot
- You never saw it catch the bug

Test-first forces you to see the test fail, proving it actually tests something.

**"I already manually tested all the edge cases"**

Manual testing is ad-hoc. You think you tested everything but:
- No record of what you tested
- Can't re-run when code changes
- Easy to forget cases under pressure
- "It worked when I tried it" ≠ comprehensive

Automated tests are systematic. They run the same way every time.

**"Deleting X hours of work is wasteful"**

Sunk cost fallacy. The time is already gone. Your choice now:
- Delete and rewrite with TDD (X more hours, high confidence)
- Keep it and add tests after (30 min, low confidence, likely bugs)

The "waste" is keeping code you can't trust. Working code without real tests is technical debt.

**"TDD is dogmatic, being pragmatic means adapting"**

TDD IS pragmatic:
- Finds bugs before commit (faster than debugging after)
- Prevents regressions (tests catch breaks immediately)
- Documents behaviour (tests show how to use code)
- Enables refactoring (change freely, tests catch breaks)

"Pragmatic" shortcuts = debugging in production = slower.

**"Tests after achieve the same goals — it's spirit not ritual"**

No. Tests-after answer "What does this do?" Tests-first answer "What should this do?"

Tests-after are biased by your implementation. You test what you built, not what's required. You verify remembered edge cases, not discovered ones.

Tests-first force edge case discovery before implementing. Tests-after verify you remembered everything (you didn't).

30 minutes of tests after ≠ TDD. You get coverage, lose proof tests work.

## Common Rationalisations

| Excuse | Reality |
|--------|---------|
| "Too simple to test" | Simple code breaks. Test takes 30 seconds. |
| "I'll test after" | Tests passing immediately prove nothing. |
| "Tests after achieve same goals" | Tests-after = "what does this do?" Tests-first = "what should this do?" |
| "Already manually tested" | Ad-hoc ≠ systematic. No record, can't re-run. |
| "Deleting X hours is wasteful" | Sunk cost fallacy. Keeping unverified code is technical debt. |
| "Keep as reference, write tests first" | You'll adapt it. That's testing after. Delete means delete. |
| "Need to explore first" | Fine. Throw away exploration, start with TDD. |
| "Test hard = design unclear" | Listen to test. Hard to test = hard to use. |
| "TDD will slow me down" | TDD faster than debugging. Pragmatic = test-first. |
| "Manual test faster" | Manual doesn't prove edge cases. You'll re-test every change. |
| "Existing code has no tests" | You're improving it. Add tests for existing code. |

## Red Flags — STOP and Start Over

- Code before test
- Test after implementation
- Test passes immediately
- Can't explain why test failed
- Tests added "later"
- Rationalising "just this once"
- "I already manually tested it"
- "Tests after achieve the same purpose"
- "It's about spirit not ritual"
- "Keep as reference" or "adapt existing code"
- "Already spent X hours, deleting is wasteful"
- "TDD is dogmatic, I'm being pragmatic"
- "This is different because..."

**All of these mean: Delete code. Start over with TDD.**

## Example: Bug Fix

**Bug:** Empty email accepted

**RED**
```typescript
test('rejects empty email', async () => {
  const result = await submitForm({ email: '' });
  expect(result.error).toBe('Email required');
});
```

**Verify RED**
```bash
$ npm test
FAIL: expected 'Email required', got undefined
```

**GREEN**
```typescript
function submitForm(data: FormData) {
  if (!data.email?.trim()) {
    return { error: 'Email required' };
  }
  // ...
}
```

**COMMIT**
```
git add -A && git commit -m "green: reject empty email with validation error"
```

**Verify GREEN**
```bash
$ npm test
PASS
```

**REFACTOR**
Extract validation for multiple fields if needed.

## Verification Checklist

Before marking work complete:

- [ ] Every new function/method has a test
- [ ] Watched each test fail before implementing
- [ ] Each test failed for expected reason (feature missing, not typo)
- [ ] Wrote minimal code to pass each test
- [ ] All tests pass
- [ ] Output pristine (no errors, warnings)
- [ ] Tests use real code (mocks only if unavoidable)
- [ ] Edge cases and errors covered
- [ ] Committed at every green phase
- [ ] Micro-commits squashed into atomic commits with Conventional Commits format

Can't check all boxes? You skipped TDD. Start over.

## When Stuck

| Problem | Solution |
|---------|----------|
| Don't know how to test | Write wished-for API. Write assertion first. Ask your human partner. |
| Test too complicated | Design too complicated. Simplify interface. |
| Must mock everything | Code too coupled. Use dependency injection. |
| Test setup huge | Extract helpers. Still complex? Simplify design. |
| Refactor broke a test | Revert to last green commit (60-second rule). |
| Cycle feels slow | Steps are too large. Make increments smaller, not bigger. |

## Debugging Integration

Bug found? Write failing test reproducing it. Follow TDD cycle. Test proves fix and prevents regression.

Never fix bugs without a test.

## Testing Anti-Patterns

When adding mocks or test utilities, read the testing-anti-patterns reference to avoid common pitfalls:
- Testing mock behaviour instead of real behaviour
- Adding test-only methods to production classes
- Mocking without understanding dependencies
- Incomplete mocks that hide structural assumptions
- Tests as afterthought rather than part of implementation

## Final Rule

```
Production code → test exists and failed first
Otherwise → not TDD
```

No exceptions without your human partner's permission.

## Workflow Summary

```
┌─────────────────────────────────────────────┐
│           DEVELOPMENT CYCLE                 │
│                                             │
│   ┌─────┐    ┌───────┐    ┌──────────┐      │
│   │ RED │───▶│ GREEN │───▶│ REFACTOR │──┐   │
│   └─────┘    └───┬───┘    └────┬─────┘  │   │
│                  │             │        │   │
│               commit        commit      │   │
│                  │             │        │   │
│                  ▼             ▼        │   │
│   ┌─────┐   micro-commits accumulate    │   │
│   │ RED │◀──────────────────────────────┘   │
│   └─────┘                                   │
│       │                                     │
│       ▼  (cycle complete?)                  │
│                                             │
│   ┌──────────────────────────────────────┐  │
│   │ SQUASH & PUSH                        │  │
│   │ git rebase -i → atomic commit(s)     │  │
│   │ Use atomic-commits skill for format  │  │
│   └──────────────────────────────────────┘  │
└─────────────────────────────────────────────┘
```
