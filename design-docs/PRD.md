# Product Requirements Document — doozer-scaffold

**Version:** 0.2-draft  
**Status:** Draft  
**Owner:** Tony Lea  

---

## 1. Overview

`doozer-scaffold` is an opinionated, interactive CLI tool written in Go that scaffolds new software projects. Given a set of user choices — technologies used, remote hosting provider, licence, and supplementary documentation — it produces a consistent, ready-to-use project structure with appropriate folder layouts, template files, `.gitignore`, and CI pipeline configuration.

The tool is primarily built for the author's own use but will be open-sourced. It is intentionally opinionated: rather than being a general-purpose scaffolder, it encodes specific structural and tooling preferences that reflect the author's working style.

---

## 2. Goals

- Eliminate repetitive manual setup when starting a new project.
- Enforce consistent project structure, documentation, and CI configuration across all future projects.
- Provide a pleasant, interactive terminal experience with sensible defaults.
- Be distributable via Homebrew tap and GitHub Releases.

## 3. Non-Goals

- This is not a general-purpose project generator intended to satisfy arbitrary user preferences.
- It will not generate application boilerplate code beyond structural scaffolding and template files.
- It will not manage existing projects or perform migrations.
- GUI or web interface delivery is out of scope.

---

## 4. Users

The primary user is the author. As an open-source project, secondary users may adopt it, but their needs do not drive prioritisation. The tool should nonetheless be sufficiently self-explanatory for a technically competent user to pick up without external documentation beyond the README.

---

## 5. Core Concepts

### 5.1 Technologies

A project may comprise one or more technologies. Each technology selection drives:

- A corresponding subdirectory structure within the project.
- Relevant template files appropriate to that technology.
- Contribution to a composite `.gitignore` covering all selected technologies.
- Relevant job blocks within the generated CI pipeline configuration.

The specific technologies supported at each stage, and their exact folder structures and template files, are defined in the SPEC for that stage. The full set of technologies the tool will eventually support is maintained in the project backlog.

### 5.2 Remote Hosting Provider

The user selects a remote hosting provider. This choice determines:

- Which CI/CD pipeline format is generated.
- Which repository configuration files are scaffolded.
- Which provider-specific options are presented to the user.

The specific providers supported at each stage, and their configuration details, are defined in the SPEC for that stage.

### 5.3 Repository Configuration

Depending on the chosen provider, the user may select provider-specific repository configuration to include, such as issue templates, pull request templates, or pipeline definitions. The specific options available for each provider are defined in the SPEC for that stage.

### 5.4 Licence

The user selects a licence to apply to the project. A `LICENSE` file is generated from a bundled static template. The set of licences available at each stage is defined in the SPEC for that stage, with coverage expanding over successive stages.

### 5.5 Supplementary Documentation

The user may select one or more standard community documents to include. These are generic, community-standard static files bundled with the tool. The specific documents available at each stage are defined in the SPEC for that stage, with coverage expanding over successive stages.

---

## 6. User Experience

The tool is invoked from the terminal with no required arguments. It presents an interactive prompt-driven flow:

1. Select remote hosting provider.
2. Select one or more technologies.
3. Select a licence.
4. Select which supplementary documentation to include.
5. Select any provider-specific repository configuration.
6. Confirm selections.
7. Scaffold is generated in the current working directory or a named subdirectory.

The interaction model supports keyboard-driven multi-select and single-select inputs, clear prompts, and a confirmation step before any files are written. The tool must not write any files until the user confirms.

Error states — such as a target directory that already exists and is non-empty — must be handled gracefully with clear messaging. The tool must not leave partial output on failure.

---

## 7. Output Structure

Every scaffold, regardless of technology selection, produces:

- `README.md` — stub, pre-populated with the project name.
- `LICENSE` — populated from the selected licence template, if a licence is selected.
- `.gitignore` — composite, generated from all selected technologies.
- A CI configuration file — location and format determined by the chosen remote provider.

Beyond these, the directory structure and template files are determined by the selected technologies. The precise output for each technology is defined in the SPEC for that stage.

---

## 8. Testing Strategy

All development follows strict Test-Driven Development (TDD). No production code is written without a failing test driving it first.

In addition to unit-level TDD, each stage has acceptance-level tests that verify the complete expected output of that stage as a whole. These tests operate at the boundary of the tool — invoking it as a user would and asserting on the resulting file system state, directory structure, and file contents. This provides confidence that the sum of the parts delivers the intended stage goal, not just that individual functions behave correctly in isolation.

The tool's own CI pipeline runs linting and unit tests on every push, and linting, unit tests, and acceptance tests on pull request creation.

## 9. Versioning and Changelog

The project follows Conventional Commits for all commit messages. Versioning is automated based on commit history, and a changelog is generated automatically from conventional commit messages on release.

Pre-commit hooks are used to enforce linting and validate that commit messages conform to the Conventional Commits specification before a commit is accepted locally.

---

## 10. Distribution

- **GitHub Releases:** Pre-compiled binaries for macOS (arm64, amd64) and Linux (amd64) attached to each tagged release.
- **Homebrew tap:** A dedicated tap repository providing `brew install` support.
- The release process is automated via the CI pipeline on tag push.

Distribution infrastructure is established in the final stage, once the tool's feature set is complete.

---

## 11. Delivery Stages

The project is delivered in discrete stages. Each stage is a self-contained, releasable increment that builds on the previous one. Each stage produces a SPEC document that defines the detailed technical requirements for that sprint.

The sequencing principle is: establish the full toolchain and delivery pipeline with the smallest possible scope first, then expand technology coverage incrementally, then expand provider support, then add polish and hardening.

---

### Stage 1 — Foundations

**Goal:** A working CLI that scaffolds a single-technology project on GitHub, proving the full toolchain, project structure, TDD approach, and CI pipeline from the outset. Everything that ships later is built on top of what is established here.

**Scope:**

- Single technology selection (no multi-select yet). The specific technology is defined at SPEC stage.
- GitHub as the only supported remote provider.
- A minimal initial set of licence options.
- A minimal initial set of supplementary documentation options.
- GitHub-specific repository configuration options.
- Confirmation step before any file generation.
- Correct generation of universal outputs: README, LICENSE, .gitignore, CI config.
- Correct generation of technology-specific directory structure and template files.
- The tool's own GitHub Actions pipeline running tests on PR and merge to main.
- Full unit test coverage via TDD.
- Acceptance tests verifying end-to-end scaffold output for the supported technology.

**Success criteria:** A user can run `doozer-scaffold`, make selections, confirm, and receive a correctly structured project directory that matches the expected output defined in the Stage 1 SPEC.

---

### Stage 2 — Multi-Technology Support

**Goal:** Extend the tool to support selecting multiple technologies simultaneously, with correctly merged output across all selections.

**Scope:**

- Multi-select technology input replaces single-select.
- 2–3 additional technologies added. The specific technologies are defined at SPEC stage.
- Composite `.gitignore` generation from multiple technology selections.
- CI pipeline configuration correctly incorporates relevant blocks for each selected technology.
- Directory structure correctly reflects all selected technologies without conflict.
- Acceptance tests cover representative multi-technology combinations.

**Success criteria:** A user can select multiple technologies and receive a coherent scaffold that correctly reflects all choices without conflict.

---

### Stage 3 — Technology Expansion

**Goal:** Continue growing the set of supported technologies.

**Scope:**

- 2–3 further technologies added. The specific technologies are defined at SPEC stage.
- Any structural or CI patterns introduced by the new technologies are implemented correctly.
- Acceptance test coverage extended for new technologies and relevant combinations.

**Success criteria:** All technologies added in this stage produce correct, conflict-free output in isolation and in combination with previously supported technologies.

---

### Stage 4 — Technology Expansion (continued)

**Goal:** Continue growing the set of supported technologies.

**Scope:** As per Stage 3. Specific technologies defined at SPEC stage.

**Success criteria:** As per Stage 3.

---

### Stage 5 — Technology Expansion (continued)

**Goal:** Complete the full set of supported technologies.

**Scope:** As per Stage 3. Specific technologies defined at SPEC stage. At the close of this stage all technologies in the project backlog should be supported, unless the backlog has been revised.

**Success criteria:** As per Stage 3.

---

### Stage 6 — Azure DevOps Provider Support

**Goal:** Add Azure DevOps as a second remote hosting provider, with ADO-specific pipeline and repository configuration output.

**Scope:**

- Provider selection expanded to include Azure DevOps.
- GitHub-specific options are conditionally presented only when GitHub is selected.
- ADO-specific pipeline configuration generated when ADO is selected.
- ADO-specific repository configuration scaffolded where applicable.
- Acceptance tests cover ADO provider selection across representative technology combinations.

**Success criteria:** Selecting ADO produces correct ADO-specific output with no GitHub-specific files present, and vice versa.

---

### Stage 7 — Licence and Documentation Expansion

**Goal:** Complete licence coverage and expand supplementary documentation options to their full intended set.

**Scope:**

- Full intended licence set supported. Specific licences defined at SPEC stage.
- Full intended supplementary documentation set supported. Specific documents defined at SPEC stage.
- Acceptance tests verify correct output for all licence and documentation options.

**Success criteria:** All intended licences and supplementary documents are available and produce correct output.

---

### Stage 8 — Robustness, Polish, and Distribution

**Goal:** Harden the tool against edge cases, improve the user experience, complete distribution infrastructure, and ensure the tool is ready for wider open-source use.

**Scope:**

- Graceful handling of non-empty target directories.
- Dry-run mode: show what would be generated without writing files.
- Flag-driven non-interactive mode for scripted and CI use.
- Improved error messaging throughout.
- `--version` flag.
- Full README and usage documentation.
- Homebrew tap established and tested.
- Automated release pipeline producing GitHub Releases binaries on tag push.
- Review of all bundled static content for correctness.
- Any outstanding acceptance test gaps addressed.

**Success criteria:** The tool handles all common error conditions gracefully, supports scripted use, is installable via Homebrew, and is documented sufficiently for an external user to adopt it without prior context.