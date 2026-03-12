# Architecture Decision Record — Stage 3c

**Date:** 2026-03-12  
**Related SPEC:** SPEC Stage 3c

---

## ADR-045: Ansible Role as Standalone-Only Technology

### Context

Ansible Roles are self-contained automation units published to Ansible Galaxy or used directly. Unlike Terraform or Dockerfile, which have natural standalone and composable variants, an Ansible Role *is* the project — there is no use case for nesting a role inside a Go or Python project. The role's directory structure (tasks, handlers, defaults, vars, meta, etc.) occupies the project root and follows a rigid layout defined by the Ansible ecosystem.

### Decision

Ansible Role is a standalone-only technology (`standalone: true`, no `variant_group`). It follows the same pattern as PowerShell Module — a single YAML definition file, no composable variant.

### Rationale

- An Ansible Role's directory structure is standardised and fills the project root. Nesting it under a subdirectory would break `ansible-galaxy`, `molecule`, and `ansible-lint` which all expect the role layout at the working directory root.
- No real-world project pattern combines an Ansible Role as a supporting concern within another technology's project.
- Adding a composable variant would create a definition that no user would select. Avoid unused code.

### Consequences

- Single definition file: `technologies/ansible-role.yaml` with `standalone: true`.
- The technology appears in the prompt alongside other standalone technologies (PowerShell Module, Terraform Module).
- Selecting Ansible Role locks out all other technology selections, enforced by existing standalone validation logic.

---

## ADR-046: Ansible Galaxy Role Directory Structure

### Context

The scaffolded Ansible Role needs a directory structure. Two approaches: a custom layout, or the standard layout produced by `ansible-galaxy init`.

### Decision

Use the standard `ansible-galaxy init` directory structure with all seven standard subdirectories: `defaults/`, `files/`, `handlers/`, `meta/`, `tasks/`, `templates/`, `vars/`. Each directory that expects a `main.yml` gets one with minimal boilerplate. Empty directories (`files/`, `templates/`) get `.gitkeep` files.

### Rationale

- `ansible-galaxy init` produces the community-standard layout. Deviating from it confuses Ansible developers and breaks tooling expectations.
- All seven directories are included even though a minimal role might not use all of them. The cost of empty directories is negligible, and removing unused directories is trivial — adding missing ones later requires knowing they should exist.
- `meta/main.yml` is populated with Galaxy metadata boilerplate (role name, author, description, licence, minimum Ansible version, supported platforms) because Galaxy publishing and `ansible-lint` expect it.

### Consequences

- The scaffold produces a complete, Galaxy-compatible role structure.
- `{{.ProjectName}}` is substituted into `meta/main.yml` for the role name.
- The `README.md` is the universal output (not duplicated inside the role).

---

## ADR-047: Molecule for Role Testing Instead of Legacy tests/ Directory

### Context

`ansible-galaxy init` generates a `tests/` directory containing `inventory` and `test.yml` — a minimal ad-hoc test approach. Molecule is the community-standard testing framework for Ansible Roles, providing automated instance provisioning, convergence, idempotency checking, and verification.

### Decision

Replace the legacy `tests/` directory with a `molecule/default/` scenario directory containing `molecule.yml`, `converge.yml`, and `verify.yml`. No `tests/` directory is generated.

### Rationale

- Molecule is the Red Hat-backed, community-standard testing framework for Ansible. The legacy `tests/inventory` + `tests/test.yml` pattern is outdated and not recommended by current best practice.
- Molecule integrates with CI via `molecule test`, providing a single command that handles create → converge → idempotence → verify → destroy.
- The default scenario uses the Docker driver, which aligns with the project's existing Docker-in-Docker devcontainer feature pattern.
- Molecule's `converge.yml` applies the role to a test instance. `verify.yml` asserts expected state using Ansible modules (the Ansible verifier, which is now the default in Molecule).

### Consequences

- The scaffold produces `molecule/default/` with three files instead of `tests/`.
- `molecule.yml` configures the Docker driver with a suitable test image.
- `converge.yml` applies the role under test.
- `verify.yml` provides a minimal verification playbook skeleton.
- CI test stage runs `molecule test`.

---

## ADR-048: ansible-lint for CI Linting

### Context

Ansible Roles need linting in CI. Available tools: `ansible-lint` (Ansible-specific best practices + YAML linting via yamllint integration), `yamllint` (pure YAML syntax), custom linting scripts.

### Decision

Use `ansible-lint` as the sole linter. Do not include a separate `yamllint` step or config file. Include a `.ansible-lint` configuration file in the scaffold with sensible defaults.

### Rationale

- `ansible-lint` subsumes `yamllint` — when yamllint is installed (which it is, as a dependency of ansible-lint), ansible-lint runs YAML linting internally with its own compatible defaults.
- Running yamllint separately alongside ansible-lint produces duplicate findings and requires maintaining a separate `.yamllint` config that must be kept compatible with ansible-lint's expectations.
- A `.ansible-lint` config file at the role root is community best practice. It allows the user to customise profiles and skip rules. The scaffold provides a sensible starting configuration using the `production` profile.
- A single tool in the lint stage keeps CI simple and consistent with the project's one-tool-per-lint-job pattern.

### Consequences

- CI `lint_steps` run `ansible-lint .` only.
- `.ansible-lint` is included as a structure entry with a sensible default configuration.
- No `.yamllint` file is generated. Users who want custom yamllint configuration can add one and ansible-lint will respect it.

---

## ADR-049: Molecule test for CI Test Stage

### Context

The CI test stage needs a meaningful test step for Ansible Roles. Options: `molecule test` (full lifecycle testing), `ansible-playbook --syntax-check` (syntax only), or no test step.

### Decision

Use `molecule test` for the CI test stage. The CI runner must have Docker available.

### Rationale

- `molecule test` is the community standard for testing Ansible roles in CI. It exercises the full lifecycle: dependency resolution, syntax check, instance creation, convergence, idempotency, verification, and cleanup.
- `ansible-playbook --syntax-check` is a strict subset of what ansible-lint already covers in the lint stage. Using it in the test stage would add no value beyond lint.
- GitHub Actions runners have Docker pre-installed, so the Docker driver works without additional setup.
- The existing CI pattern requires each technology to have a meaningful test step. For Ansible, Molecule is that step.

### Consequences

- CI `test_steps` run `molecule test`.
- CI `setup_steps` install Python, Ansible, ansible-lint, Molecule, and the Molecule Docker plugin via pip.
- The CI runner must support Docker. GitHub Actions ubuntu-latest runners do.

---

## ADR-050: Ansible Role Devcontainer Configuration

### Context

The devcontainer needs Ansible tooling. Options: devcontainer feature, pip install in setup script, or assume pre-installed.

### Decision

Install Ansible, ansible-lint, Molecule, and molecule-plugins[docker] via pip in the `setup` script. Use the Docker-in-Docker devcontainer feature for Molecule. Add the `redhat.ansible` VS Code extension.

### Rationale

- No official devcontainer feature for Ansible exists in the `devcontainers/features` registry.
- pip install is the community-standard installation method for Ansible and its tooling ecosystem.
- Docker-in-Docker is required because Molecule's Docker driver needs a Docker daemon. This is the same feature used by Dockerfile technology (ADR-033), so it deduplicates if both were ever present (though Ansible Role is standalone, so this is theoretical).
- The `redhat.ansible` VS Code extension provides Ansible-specific language support, autocompletion, and linting integration.

### Consequences

- `devcontainer.features` includes `ghcr.io/devcontainers/features/docker-in-docker:2`.
- `devcontainer.extensions` includes `redhat.ansible`.
- `devcontainer.setup` installs the full Ansible + Molecule toolchain via pip.