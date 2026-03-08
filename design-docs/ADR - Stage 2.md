# Architecture Decision Record — doozer-scaffold Stage 2

**Date:** 2026-03-08  
**Status:** Accepted  
**Parent PRD:** doozer-scaffold PRD v0.2-draft  
**Related SPEC:** SPEC Stage 2 v1.0  
**Builds on:** ADR Stage 1

---

## ADR-019: Multi-Select Technology Prompt Replaces Single-Select

### Context

Stage 1 uses a single-select prompt for technology. Stage 2 introduces multi-select so users can combine composable technologies (e.g. Go + Terraform Infrastructure) in a single project. The prompt mechanism needs to change, and the downstream engine must handle a list of technologies rather than a single one.

### Decision

The technology prompt changes from `huh.Select` to `huh.MultiSelect`. The `Config` struct changes `Technology string` to `Technologies []string`. The scaffold engine iterates over all selected technology definitions to compose the output. At least one technology must be selected — the prompt validates this.

### Consequences

- The `scaffold.Generate` function signature changes to accept `[]*techdef.TechDef` rather than a single `*techdef.TechDef`.
- All composition logic (gitignore, devcontainer features, extensions, setup.sh, CI jobs) iterates over the selected definitions.
- Technology definitions are processed in alphabetical order by key to produce deterministic output regardless of selection order.
- Existing Stage 1 acceptance tests must be updated to use the new `Technologies []string` field.

---

## ADR-020: Standalone vs Composable Technology Mode

### Context

Some technologies represent a complete project type that should not be combined with other technologies. For example, a Terraform Module follows a strict registry-published layout and should be the sole technology in the project. Similarly, a PowerShell Module has a self-contained structure. Other technologies, like Terraform Infrastructure, exist as a supporting concern within a larger project and are designed to be combined with a primary language.

### Decision

The technology definition schema gains a `standalone` boolean field (default: `false`). When `standalone: true`, selecting that technology locks out all other technology selections. The prompt enforces this — if a standalone technology is selected, no other technologies can be selected alongside it. If a composable technology is already selected, standalone options become unavailable, and vice versa.

### Consequences

- The YAML schema gains `standalone: true|false`. Existing definitions (PowerShell Module) must be updated to include `standalone: true`.
- The prompt implementation must enforce mutual exclusivity between standalone and all other technologies. This is handled in the prompt layer, not the engine — the engine simply processes whatever list it receives.
- Future technology types that are project-level (e.g. a hypothetical "React App" template) can use the same mechanism.
- Validation is added: if more than one technology is provided and any is standalone, the config is invalid. This is a defence-in-depth check — the prompt should prevent this, but the engine validates too.

---

## ADR-021: Technology Merge Strategy as Engine Behaviour, Not Schema

### Context

When multiple technologies are selected, their directory structures, gitignore entries, devcontainer contributions, CI jobs, and setup.sh blocks must be combined. The merge strategy (flat merge into the project root vs namespaced directories) could be encoded in the schema or kept as an engine implementation detail.

### Decision

The merge strategy is an engine implementation detail. Technology definitions continue to define their structures relative to the project root, and the engine merges them as-is (flat merge). If the merge strategy needs to change in a future stage, only the engine changes — definitions remain unchanged.

### Consequences

- Technology YAML files do not contain any merge-strategy metadata. They define paths relative to the project root, same as Stage 1.
- If two technologies define a path that would conflict, the engine detects this at generation time and returns an error. This is a generation-time check, not a schema-level constraint.
- Path conflict detection compares all structure entries across all selected technologies. Directory entries that share a path are allowed (both want `src/` to exist — that's fine). File entries at the same path are a conflict.
- Changing to a namespaced merge strategy in a future stage would require only engine changes and potentially updated technology definitions — no schema changes.

---

## ADR-022: Deterministic Output Ordering

### Context

When composing output from multiple technology definitions, the order of sections in `.gitignore`, features in `devcontainer.json`, extensions, CI jobs, and blocks in `setup.sh` needs to be predictable.

### Decision

All composed output is ordered alphabetically by technology definition key (the YAML filename without extension). Within a single technology's contributions, the order from the YAML definition is preserved.

### Consequences

- `.gitignore` sections appear in alphabetical order by technology name (the `name` field from the definition). Base content (if any in future) comes first.
- `devcontainer.json` features are ordered: base features first (Node.js), then technology-contributed features in alphabetical order by technology key.
- VS Code extensions are merged, deduplicated, and sorted alphabetically.
- `setup.sh` blocks appear: base block first, then technology blocks in alphabetical order by technology key.
- CI jobs appear in alphabetical order by technology key.
- Output is identical regardless of the order technologies were selected in the prompt.

---

## ADR-023: Go CLI/Library as Second Technology

### Context

A Go technology definition is needed for Stage 2. Go projects have a well-established conventional layout.

### Decision

The Go technology uses the standard Go project layout: `cmd/` for entry points, `internal/` for private packages, and standard Go tooling (golangci-lint) in the devcontainer. The technology is composable (not standalone), allowing it to be combined with supporting technologies like Terraform Infrastructure.

### Rationale

- `cmd/` + `internal/` is the conventional Go layout endorsed by the Go team and community.
- `pkg/` is deliberately excluded — the Go team has stated it is not a recommendation, and `internal/` provides enforced encapsulation.
- golangci-lint is installed via the devcontainer setup script as it is the standard Go linting aggregator.

---

## ADR-024: Terraform Module as Standalone Technology

### Context

Terraform has two usage modes: as a standalone module published to a registry, and as infrastructure supporting another technology. These have fundamentally different directory structures.

### Decision

Terraform is represented as two separate technology definitions:

1. **Terraform Module** (`terraform-module.yaml`) — standalone (`standalone: true`). Uses the standard Terraform registry layout (`main.tf`, `variables.tf`, `outputs.tf`, `modules/`, `examples/`). Cannot be combined with other technologies.
2. **Terraform Infrastructure** (`terraform-infrastructure.yaml`) — composable (`standalone: false`). Places all Terraform files under an `infrastructure/` directory. Designed to be selected alongside a primary language.

### Consequences

- Two YAML files, two prompt entries. The standalone flag on Terraform Module prevents combination.
- Both share the same devcontainer tooling (Terraform CLI, tflint) but contribute it independently — the engine's deduplication handles any overlap if both were somehow selected (which the prompt prevents for Terraform Module, but infrastructure + another tech is fine).
- The `infrastructure/` prefix on Terraform Infrastructure keeps its files cleanly separated from the primary language's structure.

---

## ADR-025: Python Package as Third Technology

### Context

A Python technology definition is needed for Stage 2. Python's packaging ecosystem has evolved significantly, and there are multiple layout conventions.

### Decision

The Python technology uses the modern `src` layout with `pyproject.toml` (PEP 621), pytest for testing, ruff for linting and formatting, and uv as the package manager in the devcontainer. The technology is composable (not standalone).

### Rationale

- The `src` layout (`src/{package_name}/`) is recommended by the Python Packaging Authority (PyPA) and prevents accidental imports from the source tree during testing.
- `pyproject.toml` is the current standard for Python project metadata (PEP 621). `setup.py` and `setup.cfg` are legacy.
- Ruff has replaced flake8, black, isort, and pyflakes as the standard Python linter/formatter. It's written in Rust, extremely fast, and covers all the use cases of the tools it replaces.
- uv is a fast, modern Python package manager and virtual environment tool. It handles venv creation, dependency resolution, and package installation in a single tool.
- pytest is the de facto standard Python test framework.

---

## ADR-026: Config Validation Enforces Standalone Constraint

### Context

The prompt layer prevents users from selecting a standalone technology alongside other technologies. However, `scaffold.Generate` can be called programmatically (as acceptance tests do), bypassing the prompt.

### Decision

`config.Validate()` checks for standalone constraint violations: if `Technologies` contains more than one entry and any selected technology definition has `standalone: true`, validation fails with a clear error message. This is a defence-in-depth check.

### Consequences

- Tests that attempt invalid combinations get a clear error rather than undefined behaviour.
- The prompt and config validation enforce the same rule independently — neither relies on the other.
- The `Validate` method needs access to the loaded technology definitions to check the standalone flag.

---

## ADR-027: Technology-Driven Prompts

### Context

Some technologies need additional user input that is specific to that technology. For example, Python needs a package name (which determines the directory under `src/`), and a future Go technology might need a Go module path. This input cannot be hardcoded in the YAML definition because it varies per project.

Three approaches were considered:

1. **Template variables in paths** — structure paths use `{{.ProjectName}}` and the engine substitutes automatically.
2. **A dedicated field per input** — e.g. `package_name` as a first-class schema field.
3. **A generic prompt mechanism** — the YAML definition declares prompts, the engine presents them, and the responses are available for substitution.

### Decision

Technology definitions can declare additional prompts via a `prompts` field. Each prompt defines a key, title, input type (text, single-select, or multi-select), and optionally a default value derived from the project name. The engine collects these prompts from all selected technologies, presents them in the prompt flow (after technology selection, before licence), and makes the responses available as template variables in that technology's `structure` paths and `content` fields.

### Rationale

- Option 1 (template variables) is too limited — it only provides the project name, and different technologies may need different transformations or entirely different inputs.
- Option 2 (dedicated fields) requires schema changes for every new technology need. It doesn't scale.
- Option 3 (generic prompts) is fully data-driven, consistent with the plugin model (ADR-001). A new technology can ask any question it needs without engine changes.

### Consequences

- The YAML schema gains a `prompts` field. Each entry has `key`, `title`, `type` (text, select, multi_select), optional `options` (for select/multi_select), and optional `default_from` to derive a default.
- Structure entry paths and content fields are processed through `text/template` with prompt responses as template variables, alongside the standard `ProjectName` and `Year`.
- `default_from: "project_name"` derives the default from the project name. For Python, this includes sanitisation (hyphens to underscores, strip leading digits) to produce a valid Python identifier.
- The sanitisation logic is specific to the derivation, not to the prompt mechanism. The `default_from` field names a source, and the engine applies context-appropriate sanitisation when deriving the default.
- Prompt keys must be unique across all selected technologies. If two technologies define the same prompt key, the engine reports an error.
- The prompt mechanism supports text input, single-select, and multi-select from the start, matching the full range of `huh` input types already used elsewhere.

---

## ADR-028: Makefile as Universal Output

### Context

A `Makefile` is useful across all project types as a standardised entry point for build commands. The content varies per project and is populated by the user. In a multi-technology scaffold, having each technology contribute a Makefile would create path conflicts.

### Decision

An empty `Makefile` is generated as a universal output for every scaffold, alongside `README.md`, `.gitignore`, and the CI config. It is always present and always empty (zero bytes).

### Rationale

- A Makefile is the project's standardised build tool entry point. Its presence signals intent and provides a consistent developer experience across projects.
- Generating it empty avoids path conflicts between technologies — there is nothing technology-specific to merge.
- The user populates it with project-specific targets. This is consistent with the tool's philosophy of providing structure, not application boilerplate.
- Making it universal (rather than conditional) means every project has a Makefile from the start, reinforcing the standard.

---

## ADR-029: Three-Stage CI Pipeline with Per-Technology Jobs

### Context

Stage 1 generated a placeholder CI config with TODO comments. As technologies are added, each technology has specific linting and testing tools that should run in CI. The CI config should be composed from technology contributions, similar to how devcontainer features and setup.sh blocks are composed. The pipeline must maintain three distinct stages — lint, test, build — as a gated progression.

### Decision

The technology definition schema gains a `ci` field. Each technology contributes a job name, optional setup steps (for runner-agnostic environment preparation), and separate lists of lint steps and test steps. The engine composes these into a three-stage GitHub Actions workflow:

1. **Lint stage:** One job per technology (`lint-{job_name}`), running in parallel. No dependencies.
2. **Test stage:** One job per technology (`test-{job_name}`), running in parallel. Each test job depends on **all** lint jobs passing — the entire lint stage must be green before any test job starts.
3. **Build stage:** A single `build` job with placeholder TODO comments. Depends on **all** test jobs passing.

The engine adds checkout as a standard preamble to every job. Setup steps from the technology definition are added after checkout, before the lint or test steps.

### Rationale

- Three stages (lint → test → build) provide a clean gated pipeline. Fast feedback on lint failures before running slower tests. Build only runs when everything is green.
- The gate model is stage-level, not job-level: all lint jobs must pass before any test job runs. This prevents wasted compute on tests when linting has already found problems.
- Per-technology parallel jobs within each stage give clear feedback on which technology failed, and allow independent technologies to run concurrently.
- The build stage is a placeholder in Stage 2. Standalone technologies like Terraform Module will use it for publishing, and language technologies will use it for compilation/packaging. The exact build steps are project-specific and deferred.
- Setup steps (not setup actions) are used for environment preparation. This is runner-agnostic — PowerShell installs `pwsh` as a step rather than relying on it being pre-installed, ensuring the CI works on any runner OS.

### Consequences

- The YAML schema gains `ci.job_name` (string), `ci.setup_steps` (optional list of name/run pairs for environment preparation), `ci.lint_steps` (list of name/run pairs), and `ci.test_steps` (list of name/run pairs).
- The generated `ci.yml` contains `lint-{job_name}` and `test-{job_name}` jobs per technology, plus a single `build` job.
- All `test-*` jobs have `needs: [lint-go, lint-python, ...]` (all lint jobs).
- The `build` job has `needs: [test-go, test-python, ...]` (all test jobs).
- Jobs within each stage are ordered alphabetically by technology key for deterministic output.
- The CI rendering is programmatic (not via `text/template`) to avoid fragile YAML-in-YAML templating.
- Technologies that do not define a `ci` field do not contribute any CI jobs.