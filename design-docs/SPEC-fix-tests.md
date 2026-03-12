# SPEC — Maintenance: Test Decoupling

**Version:** 1.0
**Status:** Draft
**Related ADR:** None (maintenance — no architectural decisions required)

---

## 0. Process Requirements

The coordinator must enforce two non-negotiable process requirements throughout all implementation work. These are hard constraints. Violation means the work is rejected, reverted, and restarted.

### 0.1 Test-Driven Development (TDD) — Mandatory

Every unit of work MUST follow strict red/green/refactor TDD:

1. **Red:** Write a failing test that specifies the expected behaviour. Run it. Confirm it fails for the expected reason.
2. **Green:** Write the minimum production code to make the test pass. Run it. Confirm it passes.
3. **Refactor:** Clean up the production code while keeping all tests green. Tests are not modified during refactor — the tests are the specification.

This cycle applies to every testable behaviour — there are no exceptions.

The project repository contains a TDD skill file at `.claude/skills/`. Lead and builder agents MUST read and follow this skill file before beginning any implementation work.

**Coordinator responsibility:** Monitor lead and builder agents for TDD compliance. If an agent writes production code before a failing test, or skips the red step, reject the work immediately. Do not allow the agent to continue and "add tests later" — revert and restart from the red step. Zero exceptions.

### 0.2 Atomic Commits — Mandatory

After TDD cycles are complete for a logical unit of work, the TDD commits (red, green, refactor) MUST be squashed into atomic commits. Each atomic commit must be:

1. **The smallest complete, meaningful change** that leaves the codebase in a working state — all tests pass, no broken imports, no dead code.
2. **Self-contained** — it does not depend on a future commit to be valid.
3. **Focused** — it does one thing.

An atomic commit always contains both the test and the production code it drove. A commit that is only tests (which would fail) is not atomic. A commit that is only production code (untested) is not atomic. The pairing of test + production code that makes it pass is the atomic unit.

The project repository contains an atomic commits skill file at `.claude/skills/`. Lead and builder agents MUST read and follow this skill file before beginning any implementation work.

**Coordinator responsibility:** The most common failure mode observed in prior stages is: TDD is followed correctly, but then the entire step is squashed into one large commit instead of multiple atomic commits. This is not acceptable. Monitor for this specific pattern and reject it. The agent must re-squash into atomic units.

### 0.3 Enforcement Summary

| Violation                                        | Action                                |
| ------------------------------------------------ | ------------------------------------- |
| Production code written before failing test      | Reject, revert, restart from red step |
| Tests and production code written simultaneously | Reject, revert, restart from red step |
| Tests modified during refactor step              | Reject, revert, redo refactor         |
| Single large commit covering entire step         | Reject, require atomic re-squash      |
| TDD correct but squashed into one big commit     | Reject, require atomic re-squash      |
| Agent claims "it's faster to skip TDD here"      | Reject unconditionally                |
| Agent claims "I'll add tests after"              | Reject unconditionally                |

---

## 1. Purpose

Three existing tests hardcode the complete inventory of technologies. When a future stage adds a new technology definition to `technologies/`, these tests break — even though the new technology has nothing to do with what the test is verifying. This forces builders of new stages to modify tests they did not write for technologies they are not changing, which risks regressions and invalidates the TDD build process.

This SPEC decouples these three tests so that adding a new technology YAML file to `technologies/` requires only new tests for that technology — never modifications to existing tests for other technologies.

### 1.1 The Coupling Principle

A test that is tightly coupled to the technology it is testing is correct and expected. `TestPythonHasPackageNamePrompt` asserting facts about `python.yaml` is fine — if Python's definition changes, that test should break.

A test that couples multiple unrelated technologies into a single assertion is the problem. `TestTechDefStandaloneField` asserting the standalone status of PowerShell, Terraform Module, Go, Terraform Infrastructure, and Python in one test means adding Dockerfile breaks it — even though Dockerfile has nothing to do with the standalone assertion for Go.

**Rule: A test may assert facts about one technology, or about the engine's behaviour with synthetic/constructed definitions. It must not enumerate or count the full set of real technology definitions.**

---

## 2. Affected Tests

### 2.1 `TestLoadAllTechDefs` — `internal/techdef/techdef_test.go`

**Current behaviour:** Loads all tech definitions, builds a sorted list of keys, and asserts exact equality against a hardcoded list:

```go
expectedKeys := []string{"dockerfile-image", "dockerfile-service", "go", "helm-chart",
    "helm-deployment", "powershell", "python", "terraform-infrastructure", "terraform-module"}
// ...
assert.Equal(t, expectedKeys, actualKeys)
```

**Why it breaks:** Any new YAML file in `technologies/` adds a key to `actualKeys` that is not in `expectedKeys`. The test fails.

**What it is actually testing:** Two things conflated into one assertion:
1. The `Load()` function can read and parse all YAML files in the embedded `technologies/` directory without error.
2. Every loaded definition passes validation.

Intent (1) is already covered by `TestAllNewDefsPassValidation`, which iterates all loaded definitions and validates each one. Intent (2) — that `Load()` works at all — is covered by any test that calls `Load()` successfully.

What is *not* covered elsewhere, and is worth preserving, is the assertion that `Load()` returns a non-empty map and that every entry is keyed by filename-without-extension. The exact set of keys is not the engine's concern — the engine is designed to be technology-agnostic.

### 2.2 `TestTechDefStandaloneField` — `internal/techdef/techdef_test.go`

**Current behaviour:** Loads all tech definitions and asserts the standalone flag for five specific technologies in one test:

```go
assert.True(t, defs["powershell"].Standalone)
assert.True(t, defs["terraform-module"].Standalone)
assert.False(t, defs["go"].Standalone)
assert.False(t, defs["terraform-infrastructure"].Standalone)
assert.False(t, defs["python"].Standalone)
```

**Why it breaks:** It does not break when a new technology is added (unlike `TestLoadAllTechDefs`). However, it creates a maintenance trap: every stage that adds a technology feels obligated to add another assertion line here, growing a single test that couples every technology in the project. If a builder adds a line for a new tech and gets the expected value wrong, it silently corrupts the test for all the other technologies that share the function.

**What it is actually testing:** That the `standalone` YAML field is correctly parsed from definitions. This is valid — but each technology's standalone status should be asserted in a test that is scoped to that technology alone.

Per-technology tests for this already exist in some cases (e.g. `TestDockerfileImageIsStandalone`, `TestDockerfileServiceIsComposable`). The pattern is correct — it just was not applied consistently, and the umbrella test was never removed.

### 2.3 `TestAcceptance_PromptCollapsesVariantGroupsToSingleEntry` — `tests/acceptance/acceptance_test.go`

**Current behaviour:** Loads all tech definitions, builds the prompt option list, and asserts:
- Specific variant group keys are present (`Helm`, `Terraform`, `Dockerfile`)
- Specific variant definition keys are absent (`helm-chart`, `helm-deployment`, etc.)
- The total count is exactly 6

```go
assert.Len(t, options, 6, "expected exactly 6 technology prompt entries after Stage 3b")
```

**Why it breaks:** Adding a new technology (whether a new variant group or a standalone tech) changes the total count. The `assert.Len` fails.

**What it is actually testing:** Two things conflated:
1. The variant group collapsing mechanism works — variant groups appear as one entry, not as individual definitions.
2. The exact inventory of technologies matches a known count.

Intent (1) is the valuable behaviour to preserve. Intent (2) is an inventory assertion that creates the same coupling problem as `TestLoadAllTechDefs`.

---

## 3. Replacement Tests

### 3.1 Replacing `TestLoadAllTechDefs`

**Delete** the existing `TestLoadAllTechDefs` function.

**Add** the following test in its place:

```go
func TestLoadReturnsNonEmptyMap(t *testing.T) {
    defs, err := techdef.Load()
    require.NoError(t, err)
    assert.NotEmpty(t, defs, "Load() must return at least one technology definition")
}

func TestLoadKeysMatchYAMLFilenames(t *testing.T) {
    defs, err := techdef.Load()
    require.NoError(t, err)

    for key, def := range defs {
        assert.NotEmpty(t, key, "technology key must not be empty")
        assert.NotContains(t, key, ".yaml", "technology key must not include file extension")
        assert.NotContains(t, key, "/", "technology key must not include path separators")
        assert.NotEmpty(t, def.Name, "technology %q must have a non-empty name", key)
    }
}
```

**What these tests preserve:**
- `Load()` successfully reads from the embedded filesystem and returns results.
- Keys are well-formed (no extensions, no path fragments).
- Every loaded definition has a name.

**What these tests do not assert:**
- The specific set of keys. That is the responsibility of each technology's own test (e.g. `TestDockerfileImageDefinitionLoads` asserts `defs["dockerfile-image"]` exists).
- The count. New technologies increase the count without breaking anything.

### 3.2 Replacing `TestTechDefStandaloneField`

**Delete** the existing `TestTechDefStandaloneField` function.

**Do not add a direct replacement.** The standalone flag for each technology is already asserted in per-technology tests:

| Technology                 | Existing test that asserts standalone status                                                   |
| -------------------------- | ---------------------------------------------------------------------------------------------- |
| `dockerfile-image`         | `TestDockerfileImageIsStandalone` — `assert.True(t, defs["dockerfile-image"].Standalone)`      |
| `dockerfile-service`       | `TestDockerfileServiceIsComposable` — `assert.False(t, defs["dockerfile-service"].Standalone)` |
| `helm-chart`               | `TestHelmChartDefinitionLoads` — `assert.True(t, def.Standalone)`                              |
| `helm-deployment`          | `TestHelmDeploymentDefinitionLoads` — `assert.False(t, def.Standalone)`                        |
| `terraform-module`         | `TestTerraformDefsHaveVariantGroup` — does NOT currently assert standalone (see gap below)     |
| `terraform-infrastructure` | `TestTerraformDefsHaveVariantGroup` — does NOT currently assert standalone (see gap below)     |
| `powershell`               | No dedicated standalone assertion exists (see gap below)                                       |
| `go`                       | No dedicated standalone assertion exists (see gap below)                                       |
| `python`                   | No dedicated standalone assertion exists (see gap below)                                       |

**Coverage gaps to fill:** Four technologies lack a per-technology standalone assertion. Add the following tests:

```go
func TestPowerShellIsStandalone(t *testing.T) {
    defs, err := techdef.Load()
    require.NoError(t, err)
    require.Contains(t, defs, "powershell")
    assert.True(t, defs["powershell"].Standalone)
}

func TestGoIsComposable(t *testing.T) {
    defs, err := techdef.Load()
    require.NoError(t, err)
    require.Contains(t, defs, "go")
    assert.False(t, defs["go"].Standalone)
}

func TestPythonIsComposable(t *testing.T) {
    defs, err := techdef.Load()
    require.NoError(t, err)
    require.Contains(t, defs, "python")
    assert.False(t, defs["python"].Standalone)
}

func TestTerraformModuleIsStandalone(t *testing.T) {
    defs, err := techdef.Load()
    require.NoError(t, err)
    require.Contains(t, defs, "terraform-module")
    assert.True(t, defs["terraform-module"].Standalone)
}

func TestTerraformInfrastructureIsComposable(t *testing.T) {
    defs, err := techdef.Load()
    require.NoError(t, err)
    require.Contains(t, defs, "terraform-infrastructure")
    assert.False(t, defs["terraform-infrastructure"].Standalone)
}
```

Each test is coupled only to the technology it names. Adding a new technology does not require touching any of these.

### 3.3 Replacing `TestAcceptance_PromptCollapsesVariantGroupsToSingleEntry`

**Delete** the existing `TestAcceptance_PromptCollapsesVariantGroupsToSingleEntry` function.

**Add** the following tests in its place:

```go
func TestAcceptance_VariantGroupsCollapseToSinglePromptEntry(t *testing.T) {
    defs, err := techdef.Load()
    require.NoError(t, err)

    options := prompt.BuildTechOptionList(defs)

    keys := make([]string, len(options))
    for i, o := range options {
        keys[i] = o.Key
    }

    // Variant group names appear as keys — individual definition keys do not.
    // Each variant group is verified independently so adding a new group
    // does not break assertions about existing groups.
    assert.Contains(t, keys, "Helm", "Helm variant group must collapse to single entry")
    assert.NotContains(t, keys, "helm-chart", "helm-chart must not appear as individual entry")
    assert.NotContains(t, keys, "helm-deployment", "helm-deployment must not appear as individual entry")

    assert.Contains(t, keys, "Terraform", "Terraform variant group must collapse to single entry")
    assert.NotContains(t, keys, "terraform-module", "terraform-module must not appear as individual entry")
    assert.NotContains(t, keys, "terraform-infrastructure", "terraform-infrastructure must not appear as individual entry")

    assert.Contains(t, keys, "Dockerfile", "Dockerfile variant group must collapse to single entry")
    assert.NotContains(t, keys, "dockerfile-image", "dockerfile-image must not appear as individual entry")
    assert.NotContains(t, keys, "dockerfile-service", "dockerfile-service must not appear as individual entry")
}

func TestAcceptance_NonVariantTechsAppearByKey(t *testing.T) {
    defs, err := techdef.Load()
    require.NoError(t, err)

    options := prompt.BuildTechOptionList(defs)

    keys := make([]string, len(options))
    for i, o := range options {
        keys[i] = o.Key
    }

    // Non-variant technologies appear by their definition key.
    assert.Contains(t, keys, "go")
    assert.Contains(t, keys, "powershell")
    assert.Contains(t, keys, "python")
}

func TestAcceptance_PromptOptionListIsSorted(t *testing.T) {
    defs, err := techdef.Load()
    require.NoError(t, err)

    options := prompt.BuildTechOptionList(defs)

    names := make([]string, len(options))
    for i, o := range options {
        names[i] = o.Name
    }

    assert.IsNonDecreasing(t, names, "prompt options must be sorted alphabetically by display name")
}
```

**What these tests preserve:**
- Variant groups collapse to a single prompt entry (the core behaviour being tested).
- Individual variant definition keys do not leak into the prompt.
- Non-variant technologies appear by their own key.
- The list is sorted alphabetically.

**What these tests do not assert:**
- The total count of options. Adding a new technology or variant group changes the count without breaking anything.
- An exhaustive list of all technologies. Each technology's presence in the prompt is that technology's own concern.

**Note on `TestAcceptance_NonVariantTechsAppearByKey`:** This test names `go`, `powershell`, and `python` — technologies that already exist and whose presence in the prompt is a valid thing to verify. It is not an inventory test because it does not claim these are the *only* non-variant technologies. Adding a new non-variant technology does not break it.

---

## 4. Files Changed

| File                                  | Action                                                                                                                                                                                                                        |
| ------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `internal/techdef/techdef_test.go`    | Remove `TestLoadAllTechDefs`. Add `TestLoadReturnsNonEmptyMap` and `TestLoadKeysMatchYAMLFilenames`. Remove `TestTechDefStandaloneField`. Add five per-technology standalone assertion tests.                                 |
| `tests/acceptance/acceptance_test.go` | Remove `TestAcceptance_PromptCollapsesVariantGroupsToSingleEntry`. Add `TestAcceptance_VariantGroupsCollapseToSinglePromptEntry`, `TestAcceptance_NonVariantTechsAppearByKey`, and `TestAcceptance_PromptOptionListIsSorted`. |

No production code changes. No new files. No changes to any other test file.

---

## 5. Acceptance Criteria

This SPEC is complete when all of the following are true:

1. `TestLoadAllTechDefs` no longer exists in the codebase.
2. `TestTechDefStandaloneField` no longer exists in the codebase.
3. `TestAcceptance_PromptCollapsesVariantGroupsToSingleEntry` no longer exists in the codebase.
4. The replacement tests specified in Section 3.1 exist and pass.
5. The per-technology standalone tests specified in Section 3.2 exist and pass.
6. The replacement acceptance tests specified in Section 3.3 exist and pass.
7. All pre-existing tests that were not modified by this SPEC continue to pass unchanged.
8. No test in the codebase contains a hardcoded list or count of all technology definition keys. Specifically: no test uses `assert.Equal` or `assert.Len` against the complete set of keys returned by `techdef.Load()` or `prompt.BuildTechOptionList()`.
9. Adding a hypothetical new YAML file `technologies/rust.yaml` (with valid content) would not cause any existing test to fail. The only tests that would need to be written are tests for Rust itself.
10. All development followed strict TDD (Section 0.1).
11. All commits are atomic (Section 0.2).
12. The project's own CI pipeline passes.

---

## 6. Implementation Order

All steps follow strict TDD (Section 0.1) and produce atomic commits (Section 0.2).

### Step 1: Replace `TestLoadAllTechDefs`

**Goal:** The inventory assertion is removed. The engine-level loading behaviour is still tested without coupling to specific technologies.

**Deliverables:**
- `TestLoadAllTechDefs` deleted.
- `TestLoadReturnsNonEmptyMap` and `TestLoadKeysMatchYAMLFilenames` added.
- All tests pass.

### Step 2: Replace `TestTechDefStandaloneField`

**Goal:** The umbrella standalone assertion is removed. Per-technology standalone tests exist for every technology that was previously covered.

**Deliverables:**
- `TestTechDefStandaloneField` deleted.
- `TestPowerShellIsStandalone`, `TestGoIsComposable`, `TestPythonIsComposable`, `TestTerraformModuleIsStandalone`, and `TestTerraformInfrastructureIsComposable` added.
- All tests pass.

### Step 3: Replace `TestAcceptance_PromptCollapsesVariantGroupsToSingleEntry`

**Goal:** The variant group collapsing behaviour is still verified. The exact count assertion is removed.

**Deliverables:**
- `TestAcceptance_PromptCollapsesVariantGroupsToSingleEntry` deleted.
- `TestAcceptance_VariantGroupsCollapseToSinglePromptEntry`, `TestAcceptance_NonVariantTechsAppearByKey`, and `TestAcceptance_PromptOptionListIsSorted` added.
- All tests pass.

### Step 4: CI verification

**Goal:** All CI checks pass.

**Deliverables:**
- Push changes, verify lint + unit tests + acceptance tests pass in CI.