# SPEC — Stage 3c: Ansible Role Technology

**Version:** 1.0  
**Related ADR:** ADR Stage 3c

---

## 0. Process Requirements

The coordinator must enforce two non-negotiable process requirements throughout all implementation work. These are hard constraints. Violation means the work is rejected, reverted, and restarted.

### 0.1 The Development Cycle

The unit of work is a single behaviour. For each behaviour, the complete cycle is:

1. **Red:** Write a failing test that specifies the expected behaviour. Run it. Confirm it fails for the expected reason.
2. **Green:** Write the minimum production code to make the test pass. Run it. Confirm it passes.
3. **Refactor:** Clean up the production code while keeping all tests green. Tests are not modified during refactor — the tests are the specification.
4. **Atomic Commit:** Squash the red/green/refactor commits into one atomic commit containing both the test and the production code it drove.

Then move to the next behaviour and repeat.

**This is one cycle per behaviour, not many behaviours batched together.** Do not accumulate multiple red/green/refactor loops and then squash them all into one commit at the end. Each behaviour gets its own cycle and its own atomic commit.

The project repository contains skill files for both TDD and atomic commits at `.claude/skills/`. Lead and builder agents MUST read and follow these skill files before beginning any implementation work.

### 0.2 What Makes an Atomic Commit

Each atomic commit must be:

1. **The smallest complete, meaningful change** that leaves the codebase in a working state — all tests pass, no broken imports, no dead code.
2. **Self-contained** — it does not depend on a future commit to be valid.
3. **Focused** — it does one thing.

An atomic commit always contains both the test and the production code it drove. A commit that is only tests (which would fail) is not atomic. A commit that is only production code (untested) is not atomic. The pairing of test + production code that makes it pass is the atomic unit.

### 0.3 Coordinator Enforcement

**TDD compliance:** If an agent writes production code before a failing test, or skips the red step, reject the work immediately. Do not allow the agent to continue and "add tests later" — revert and restart from the red step. Zero exceptions.

**Atomic commit compliance:** The most common failure mode observed in prior stages is: TDD is followed correctly, but then the entire step is squashed into one large commit instead of multiple atomic commits. This is not acceptable. A single commit covering "add YAML definition + all its tests" is too large if that definition drives multiple distinct behaviours. Monitor for this specific pattern and reject it. The agent must re-squash into atomic units.

### 0.4 Enforcement Summary

| Violation                                                   | Action                                          |
| ----------------------------------------------------------- | ----------------------------------------------- |
| Production code written before failing test                 | Reject, revert, restart from red step           |
| Tests and production code written simultaneously            | Reject, revert, restart from red step           |
| Tests modified during refactor step                         | Reject, revert, redo refactor                   |
| Multiple red/green/refactor cycles before any atomic commit | Reject, revert, restart — one cycle, one commit |
| Single large commit covering entire step                    | Reject, require atomic re-squash                |
| TDD correct but squashed into one big commit                | Reject, require atomic re-squash                |
| Agent claims "it's faster to skip TDD here"                 | Reject unconditionally                          |
| Agent claims "I'll add tests after"                         | Reject unconditionally                          |

---

## 1. Purpose

Stage 3c adds Ansible Role as a supported technology. Ansible Role is standalone-only — no composable variant exists. No schema changes, engine changes, or prompt changes are required. The existing declarative technology definition system handles everything.

---

## 2. Summary of Changes

### 2.1 Schema Changes

None.

### 2.2 Engine Changes

None.

### 2.3 Prompt Changes

None. The new technology definition is automatically picked up by the dynamic technology prompt.

### 2.4 New Definitions

| Definition   | File                             | Mode       | Detail                                                  |
| ------------ | -------------------------------- | ---------- | ------------------------------------------------------- |
| Ansible Role | `technologies/ansible-role.yaml` | Standalone | Standard Ansible Galaxy role structure at project root. |

---

## 3. Technology Definition

### 3.1 Ansible Role

File: `technologies/ansible-role.yaml`

Standalone Ansible Role project — the standard `ansible-galaxy init` directory structure. Used for creating a role intended for Galaxy publishing or direct consumption. Cannot be combined with other technologies.

```yaml
name: "Ansible Role"
standalone: true

structure:
  - path: "defaults/"
  - path: "defaults/main.yml"
    content: |
      ---
      # defaults file for {{.ProjectName}}
  - path: "files/"
  - path: "handlers/"
  - path: "handlers/main.yml"
    content: |
      ---
      # handlers file for {{.ProjectName}}
  - path: "meta/"
  - path: "meta/main.yml"
    content: |
      ---
      galaxy_info:
        role_name: {{.ProjectName}}
        author: your name
        description: your role description
        license: MIT

        min_ansible_version: "2.17"

        galaxy_tags: []

      dependencies: []
  - path: "tasks/"
  - path: "tasks/main.yml"
    content: |
      ---
      # tasks file for {{.ProjectName}}
  - path: "templates/"
  - path: "vars/"
  - path: "vars/main.yml"
    content: |
      ---
      # vars file for {{.ProjectName}}
  - path: "molecule/"
  - path: "molecule/default/"
  - path: "molecule/default/molecule.yml"
    content: |
      ---
      role_name_check: 1
      dependency:
        name: galaxy
      driver:
        name: docker
      platforms:
        - name: instance
          image: geerlingguy/docker-ubuntu2404-ansible:latest
          command: ""
          volumes:
            - /sys/fs/cgroup:/sys/fs/cgroup:rw
          cgroupns_mode: host
          privileged: true
          pre_build_image: true
      provisioner:
        name: ansible
      verifier:
        name: ansible
  - path: "molecule/default/converge.yml"
    content: |
      ---
      - name: Converge
        hosts: all
        roles:
          - role: {{.ProjectName}}
  - path: "molecule/default/verify.yml"
    content: |
      ---
      - name: Verify
        hosts: all
        gather_facts: false
        tasks:
          - name: Example verification
            ansible.builtin.assert:
              that: true
  - path: ".ansible-lint"
    content: |
      ---
      profile: production

gitignore: |
  # Ansible
  *.retry
  .molecule/
  .cache/

devcontainer:
  features:
    "ghcr.io/devcontainers/features/docker-in-docker:2": {}
  extensions:
    - "redhat.ansible"
  setup: |
    pip install --break-system-packages ansible ansible-lint molecule "molecule-plugins[docker]"

ci:
  job_name: "ansible"
  setup_steps:
    - name: "Set up Python"
      uses: "actions/setup-python@v5"
      with:
        python-version: "3.12"
    - name: "Install Ansible and tooling"
      run: |
        pip install ansible ansible-lint molecule "molecule-plugins[docker]"
  lint_steps:
    - name: "Lint"
      run: "ansible-lint ."
  test_steps:
    - name: "Molecule test"
      run: "molecule test"
```

This produces:

```
{project}/
├── .ansible-lint
├── defaults/
│   └── main.yml
├── files/
│   └── .gitkeep
├── handlers/
│   └── main.yml
├── meta/
│   └── main.yml
├── molecule/
│   └── default/
│       ├── converge.yml
│       ├── molecule.yml
│       └── verify.yml
├── tasks/
│   └── main.yml
├── templates/
│   └── .gitkeep
└── vars/
    └── main.yml
```

Note: Directories with a `main.yml` file entry also have a directory entry that creates `.gitkeep`. Both are created — `.gitkeep` from the directory entry, `main.yml` from the file entry. This is not a conflict — they are different filenames in the same directory. Directories without file entries (`files/`, `templates/`) get `.gitkeep` only.

The `molecule/` and `molecule/default/` directory entries create `.gitkeep` files. The file entries within `molecule/default/` create additional files. No conflict.

---

## 4. Path Conflict Analysis

Ansible Role is standalone — it can never be combined with other technologies. Path conflicts with other technologies are therefore impossible by design.

The CI `job_name` `"ansible"` is unique among all technologies.

The `redhat.ansible` extension is unique to Ansible Role.

---

## 5. Acceptance Criteria

Stage 3c is complete when all of the following are true:

1. The Ansible Role definition (`ansible-role.yaml`) loads, parses, and validates correctly.
2. Ansible Role is standalone — it cannot be selected alongside any other technology. The prompt prevents it and config validation enforces it.
3. Ansible Role scaffolds the complete directory structure at the project root as specified in Section 3.1.
4. All `main.yml` files contain the correct boilerplate content with `{{.ProjectName}}` substituted.
5. `meta/main.yml` contains valid Galaxy metadata with the role name set to `{{.ProjectName}}`.
6. `molecule/default/` contains `molecule.yml`, `converge.yml`, and `verify.yml` with correct content. `converge.yml` references the role by `{{.ProjectName}}`.
7. `.ansible-lint` is generated at the project root with the `production` profile.
8. The `.gitignore` correctly includes the Ansible section.
9. The devcontainer includes the `docker-in-docker:2` feature and `redhat.ansible` extension.
10. The devcontainer setup script installs Ansible, ansible-lint, Molecule, and molecule-plugins[docker] via pip.
11. CI generates `lint-ansible` and `test-ansible` jobs. Lint runs `ansible-lint .`. Test runs `molecule test`.
12. CI setup installs Python 3.12 and the Ansible toolchain via pip.
13. All new tests pass.
14. All existing tests continue to pass without modification.
15. No schema, engine, config, or prompt code changes were required — only a new YAML definition file and tests.
16. All development followed the development cycle in Section 0.1.
17. All commits are atomic (Section 0.2).

---

## 6. Implementation Order

Steps 1–3 follow the development cycle defined in Section 0.1. Step 4 is acceptance verification — the production code already exists; new tests confirm deliverables and prevent regressions. If any acceptance test fails, use TDD to fix the underlying issue.

### Step 1: Ansible Role definition

**Goal:** The Ansible Role definition exists, loads correctly, and passes all validation.

**Deliverables:**
- `technologies/ansible-role.yaml` with the content specified in Section 3.1.
- Definition loads, parses, and validates. Has expected fields: name, standalone, structure, gitignore, devcontainer, CI.
- `techdef.Load()` returns the correct total number of technology definitions (current count + 1).
- Standalone constraint works — Ansible Role cannot be combined with other technologies.

### Step 2: Scaffold output verification

**Goal:** Selecting Ansible Role produces the correct directory structure and file contents.

**Deliverables:**
- All directories created with correct `.gitkeep` files.
- All `main.yml` files created with correct content.
- `meta/main.yml` contains valid Galaxy metadata with substituted role name.
- `molecule/default/` files created with correct content.
- `.ansible-lint` created at project root.
- `{{.ProjectName}}` correctly substituted throughout.

### Step 3: Composition verification

**Goal:** Ansible Role integrates correctly with universal outputs (gitignore, devcontainer, CI).

**Deliverables:**
- `.gitignore` includes the Ansible section.
- Devcontainer includes Docker-in-Docker feature, `redhat.ansible` extension, and pip-based setup.
- CI generates `lint-ansible` and `test-ansible` jobs with correct setup and steps.

### Step 4: Acceptance verification

**Goal:** Confirm larger deliverables work end-to-end against the acceptance criteria in Section 5. The production code already exists from Steps 1–3. These tests verify the assembled result and guard against regressions. If any acceptance test fails, use TDD to fix the underlying issue.

**Deliverables:**
- Ansible Role in isolation produces correct output with maximum selections (all licence, docs, tooling, repo config options).
- Ansible Role in isolation produces correct output with minimum selections.
- Ansible Role standalone constraint rejection works (cannot combine with other technologies).
- All prior-stage tests still pass.