# Architecture Decision Record — doozer-scaffold Stage 1

**Date:** 2026-03-08
**Status:** Accepted
**Parent PRD:** doozer-scaffold PRD v0.2-draft
**Related SPEC:** SPEC Stage 1 v2.0

---

## ADR-001: Declarative Technology Definitions (Plugin Model)

### Context

The tool needs to support a growing set of technologies across Stages 1–5. Each technology contributes a directory structure, files, gitignore entries, devcontainer features, VS Code extensions, and setup script commands. The original approach was to encode this logic imperatively in the Go scaffold engine, with template files and conditional code paths per technology.

### Decision

Technologies are defined declaratively via YAML definition files in a `technologies/` directory. Each file defines the complete contribution of a single technology: its directory structure, files (with optional content), gitignore entries, and devcontainer configuration. The scaffold engine is entirely data-driven — it reads these definitions and acts on them without any technology-specific conditional logic.

### Consequences

- Adding a new technology requires only a new YAML file and acceptance tests. No changes to the scaffold engine, prompt code, or template rendering logic.
- The prompt's technology options are populated dynamically from loaded definitions, so new technologies appear automatically.
- The gitignore content and devcontainer setup script content live in the YAML definitions rather than in separate template files, eliminating the `templates/gitignore/` directory and per-technology setup script templates.
- The engine must support three structure entry types from Stage 1: directories (with `.gitkeep`), files with content, and empty files. The PowerShell definition only uses directories initially, but the full capability is built upfront to avoid extending the engine later.
- Technology definitions are validated on load (non-empty name, valid paths, no path traversal), catching errors early.

---

## ADR-002: charmbracelet/huh for Interactive Prompts

### Context

The CLI requires interactive terminal prompts supporting text input, single-select, multi-select, and confirmation. Go has several TUI/prompt libraries available, including `charmbracelet/huh`, `charmbracelet/bubbletea`, `AlecAivazis/survey`, and `manifoldco/promptui`.

### Decision

Use `charmbracelet/huh` for all interactive prompts.

### Rationale

- Provides clean, composable form-based prompts with built-in support for all required input types (text, select, multi-select, confirm).
- Produces polished terminal output with minimal configuration.
- Part of the well-maintained Charm ecosystem, which is the de facto standard for Go CLI user interfaces.
- Higher-level abstraction than `bubbletea` (which `huh` is built on), avoiding unnecessary complexity for a prompt-driven flow that doesn't need custom TUI widgets.
- `survey` is archived/unmaintained. `promptui` has limited multi-select support.

---

## ADR-003: testify for Test Assertions

### Context

The project follows strict TDD. Tests are written first and serve as living documentation. The standard Go `testing` package provides only `t.Error` and `t.Fatal`, requiring manual comparison and error message construction.

### Decision

Use `github.com/stretchr/testify` (`assert` and `require` packages) for all test assertions.

### Rationale

- Expressive assertions (`assert.Equal`, `assert.Contains`, `assert.FileExists`, `assert.ElementsMatch`, `require.NoError`) make test intent immediately clear.
- `assert` vs `require` distinction maps naturally to "continue on failure" vs "stop immediately" semantics, which is important when later assertions depend on earlier ones (e.g. "the directory must exist before we check its contents").
- Widely adopted in the Go ecosystem, well-maintained, and familiar to most Go developers.

---

## ADR-004: gopkg.in/yaml.v3 for YAML Parsing

### Context

The technology definition plugin model (ADR-001) requires parsing YAML files. Go has multiple YAML libraries available.

### Decision

Use `gopkg.in/yaml.v3` for parsing technology definition YAML files.

### Rationale

- De facto standard YAML library for Go.
- Well-maintained with full YAML spec support.
- Clean struct tag-based unmarshalling that maps directly to the `TechDef` data structure.

---

## ADR-005: text/template with go:embed for Template Rendering

### Context

The tool generates files from templates (README, LICENSE, CI config, etc.). Templates need to be bundled with the binary for single-file distribution.

### Decision

Use Go's standard `text/template` package for template rendering and `go:embed` to embed template files into the binary at compile time.

### Rationale

- `text/template` is part of the standard library — no additional dependency required.
- Sufficient for the tool's needs (simple variable substitution in text files).
- `html/template` was considered but adds unnecessary HTML escaping for non-HTML output.
- `go:embed` produces a self-contained binary with no external file dependencies, which simplifies distribution via Homebrew and GitHub Releases.

### Exception

`devcontainer.json` is rendered programmatically via `encoding/json` with `json.MarshalIndent` rather than via `text/template`. This is because the JSON structure varies dynamically based on which features and extensions are contributed by the selected technology. Constructing JSON via string templates is fragile and error-prone; building a Go struct and marshalling it produces guaranteed-valid JSON.

---

## ADR-006: Devcontainer Features for Language Runtimes

### Context

The scaffolded project includes a devcontainer configuration. Language runtimes (Node.js, PowerShell) could be installed either directly in the Dockerfile (via apt/nvm/manual installation) or via devcontainer features.

### Decision

Use devcontainer features for language runtimes. The Dockerfile is kept minimal (base image + OS-level `apt` packages only). The `devcontainer.json` features array is composed from a base set (Node.js, always present) plus contributions from the selected technology definition.

### Rationale

- Current best practice for devcontainers. Features handle idempotent installation and version management cleanly.
- Keeps the Dockerfile lean and focused on OS-level dependencies.
- Composable — each technology definition contributes its features to the array, and the engine performs a simple merge. No Dockerfile surgery required when adding technologies.
- PowerShell modules (Pester, PSScriptAnalyzer, PlatyPS, BuildHelpers) and Node.js global packages (markdownlint-cli2, commitlint) are installed via `setup.sh` rather than in the Dockerfile, because they sit on top of the runtimes provided by features.

---

## ADR-007: Composable setup.sh for Post-Create Tooling

### Context

Each technology requires tooling installed after the container is created (PowerShell modules, linting tools, etc.). With multiple technologies selected, these commands need to be combined. Three approaches were considered:

1. Inline `postCreateCommand` in `devcontainer.json` — concatenating shell commands into a single string.
2. Multi-stage Dockerfile builds — each technology contributes a build stage.
3. A composed `setup.sh` script — assembled from template blocks, called by `postCreateCommand`.

### Decision

Use a composed `setup.sh` script. The script is assembled from a base block (always present, installs markdownlint-cli2 and commitlint) followed by technology-specific blocks taken from each technology definition's `devcontainer.setup` field.

### Rationale

- Option 1 (inline `postCreateCommand`) becomes fragile with multiple technologies — string concatenation of shell commands is hard to read, test, and debug.
- Option 2 (multi-stage Dockerfile) is over-engineered for installing tooling that depends on runtimes provided by features. It also couples the Dockerfile tightly to technology selection.
- Option 3 (setup.sh) is clean to template, easy to test (assert on script content like any other generated file), and scales to any number of technologies. Each technology contributes a clearly delimited block with a header comment.

---

## ADR-008: ARM-Only Devcontainer in Stage 1

### Context

The existing PowerShell module project used dual-architecture devcontainers (amd64 and arm64) with separate `devcontainer.json` files for each.

### Decision

Stage 1 supports only the ARM architecture for devcontainers. A single `.devcontainer/` directory contains one `Dockerfile`, one `devcontainer.json`, and one `setup.sh`. AMD64 support is deferred to the end of the project.

### Rationale

- The primary development environment is ARM-based (M4 Mac Mini).
- AMD64 devcontainers cannot be tested locally on ARM hardware without cross-architecture emulation, which is slow and unreliable.
- Shipping untested configuration would undermine the project's TDD principles.
- Adding AMD64 later is a straightforward extension (additional devcontainer directory or multi-platform image) once proper testing infrastructure is in place.

---

## ADR-009: Single-Option Prompts Presented as Selection Prompts

### Context

In Stage 1, some prompts have only one option (e.g. "Remote hosting provider: GitHub", "Technology: PowerShell Module"). These could be hard-coded/skipped or still presented as interactive selections.

### Decision

Single-option prompts are still presented as interactive selection prompts, not hard-coded.

### Rationale

- Additional providers (Azure DevOps, Stage 6) and technologies (Stages 2–5) will be added in future stages. Keeping the prompt architecture consistent from Stage 1 means no structural changes are needed when options expand.
- The user sees the full prompt flow from the start, setting the right expectation for the tool's interaction model.
- The implementation cost is zero — the code is identical whether there is one option or ten.

---

## ADR-010: scaffold.Generate Accepts a Base Directory Parameter

### Context

The `Generate` function needs to create the scaffold on the filesystem. Tests need to verify the output without affecting the real working directory.

### Decision

`scaffold.Generate` accepts a `baseDir` parameter. The scaffold is created in a subdirectory of `baseDir` named after the project. In production, `baseDir` is `"."` (current working directory). In tests, `baseDir` is `t.TempDir()`.

### Rationale

- Full filesystem isolation in tests without monkeypatching, environment variable manipulation, or working-directory changes.
- Every test gets a unique temporary directory that is automatically cleaned up.
- The function signature makes the dependency on the filesystem explicit, which improves testability and readability.

---

## ADR-011: Acceptance Tests Drive Generate Programmatically

### Context

Acceptance tests need to verify the complete scaffold output. Two approaches were considered:

1. Shell out to the compiled binary and inspect the resulting filesystem.
2. Call `scaffold.Generate` programmatically with a `config.Config` struct and assert on the filesystem.

### Decision

Acceptance tests call `scaffold.Generate` directly, passing a `Config`, a loaded `TechDef`, and a `t.TempDir()` base directory.

### Rationale

- No dependency on a pre-compiled binary. Tests run with `go test` and don't require a separate build step.
- Tests can set up precise configurations without simulating interactive terminal input, which is inherently fragile and platform-dependent.
- The interactive prompt layer is thin (it populates a `Config` struct) and is validated through manual testing. The bulk of the logic — file generation, template rendering, technology definition processing — is covered by programmatic tests.
- Filesystem assertions (`assert.FileExists`, `assert.ElementsMatch` on walked file lists) verify both inclusion and exclusion of files, catching cases where conditional files leak into the wrong scenario.

---

## ADR-012: Static Files Stored as .tmpl Files

### Context

Some generated files are static (no template variables), such as `.editorconfig` and `.gitattributes`. These could be stored as plain files or as `.tmpl` files like the rest of the templates.

### Decision

All template files use the `.tmpl` extension, including static files with no template variables.

### Rationale

- Consistency — the template loading and rendering pipeline handles all files uniformly without needing to distinguish between "static" and "dynamic" files.
- Future-proofing — if a currently-static file later needs template variables (e.g. `.gitattributes` gaining technology-specific entries), no pipeline changes are required.
- The overhead of processing a template with no variables is negligible.

---

## ADR-013: No External CLI Framework in Stage 1

### Context

Go has several CLI frameworks (Cobra, urfave/cli, Kong) that provide argument parsing, subcommands, and help text generation. The tool needs only a single optional positional argument in Stage 1.

### Decision

Use Go's standard `os.Args` or `flag` package for CLI argument handling. No external CLI framework.

### Rationale

- Stage 1 has minimal CLI requirements — an optional positional argument for the project name. A full CLI framework would be over-engineering.
- A framework can be introduced in Stage 8 (polish and distribution) if needed for `--version`, `--dry-run`, or other flags. At that point the requirements will be clearer.
- Fewer dependencies in Stage 1 means faster builds and a simpler dependency graph.

---

## ADR-014: PowerShell as the First Supported Technology

### Context

Stage 1 requires a single technology to prove the full toolchain. Candidates included Go, Python, Terraform, and PowerShell.

### Decision

PowerShell Module is the first supported technology.

### Rationale

- The structure is simple (src/public, src/private, src/classes, tests) and produces a clear, easily verifiable scaffold.
- The devcontainer configuration exercises the features model well (PowerShell runtime via feature, multiple modules via setup.sh).
- Go was considered but would create a bootstrapping problem — the tool is itself written in Go, and scaffolding Go projects while building a Go project adds unnecessary cognitive load to Stage 1.
- Starting with a non-Go technology keeps the tool's own structure cleanly separated from the scaffolded output in the developer's mind.

---

## ADR-015: Devcontainer Base Image Selection

### Context

The devcontainer Dockerfile needs a base image. Options considered were `mcr.microsoft.com/powershell:latest` (PowerShell-specific) and `mcr.microsoft.com/devcontainers/base:ubuntu` (Microsoft's general-purpose devcontainer base).

### Decision

Use `mcr.microsoft.com/devcontainers/base:ubuntu` as the base image.

### Rationale

- The general-purpose devcontainer base is more extensible for multi-technology support in later stages. A PowerShell-specific base would need replacing when non-PowerShell technologies are added.
- The devcontainers base image is designed to work with devcontainer features, which is the chosen mechanism for installing language runtimes (ADR-006).
- PowerShell is installed via a devcontainer feature on top of the base image, keeping each technology's contribution modular and composable.

---

## ADR-016: Base Devcontainer Tooling

### Context

The devcontainer setup includes a base tooling layer installed regardless of technology selection. The scope of this base layer needed to be defined.

### Decision

The Stage 1 base layer includes git, Node.js (via devcontainer feature), markdownlint-cli2, and commitlint (with conventional config). Additional base tools (cspell, secretlint) are deferred to a later stage.

### Rationale

- Git is universally required for any project.
- Node.js is required as a runtime for markdownlint-cli2 and commitlint.
- Markdownlint and commitlint directly support the project's documentation standards and Conventional Commits requirement from the PRD.
- Cspell and secretlint (both present in the reference PowerShell project) are useful but not essential for Stage 1 foundations. Including them would expand the base layer beyond what's needed to prove the architecture. They can be added as the tooling set matures.

---

## ADR-017: .editorconfig and .gitattributes as Optional Tooling Selections

### Context

`.editorconfig` and `.gitattributes` are project configuration files that could be treated as either universal outputs (always generated) or optional selections. They differ from supplementary documentation (CONTRIBUTING.md) in that they're tooling configuration rather than community documents.

### Decision

`.editorconfig` and `.gitattributes` are presented as optional selections in a dedicated "Project tooling" prompt step, separate from the "Supplementary documentation" prompt.

### Rationale

- Not every project needs these files. Making them optional respects the user's choice.
- They are functionally distinct from community documentation — one configures editors, the other configures Git's handling of line endings. Mixing them into the documentation prompt would be a category error.
- A dedicated "Project tooling" prompt step creates a clear extension point for future tooling options (e.g. `.prettierrc`, `.eslintrc`) without overcrowding the documentation prompt.

---

## ADR-018: Full Structure Entry Capability from Stage 1

### Context

The technology definition schema supports three types of structure entries: directories (with `.gitkeep`), files with content, and empty files. The PowerShell definition in Stage 1 only uses directory entries.

### Decision

The scaffold engine implements all three structure entry types from Stage 1, even though the first technology definition only exercises one.

### Rationale

- Future technology definitions (e.g. Go with `go.mod`, Python with `pyproject.toml`) will require files with content. Building the capability now avoids engine changes in Stages 2–5.
- The implementation is straightforward — the distinction between entry types is a simple conditional based on path suffix and content field presence.
- Unit tests cover all three entry types from Stage 1, so the capability is proven from the start.
- This aligns with the project's principle of establishing the full toolchain early and expanding incrementally.