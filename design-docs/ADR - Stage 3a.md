# Architecture Decision Record — doozer-scaffold Stage 3a

**Date:** 2026-03-09  
**Status:** Accepted  
**Parent PRD:** doozer-scaffold PRD v0.2-draft  
**Related SPEC:** SPEC Stage 3a v1.0  
**Builds on:** ADR Stage 1, ADR Stage 2

---

## ADR-030: Stage 3 Split into Sub-Stages per Technology

### Context

The PRD defines Stage 3 as "Continue growing the set of supported technologies" with 2–3 technologies per stage. As the project scales, each technology is a self-contained unit of work — a new YAML definition file and corresponding acceptance tests. Bundling multiple technologies into a single stage creates larger, harder-to-review increments.

### Decision

Stage 3 is split into sub-stages (3a, 3b, etc.), with each sub-stage adding exactly one technology (which may include both a standalone and composable variant, as with Terraform in Stage 2). Stage 3a adds Dockerfile.

### Rationale

- Each sub-stage is a small, focused increment that is easy to review, test, and merge.
- The declarative technology definition model (ADR-001) means each technology is fully self-contained — no engine changes are needed, so there is no coupling between technologies that would benefit from being shipped together.
- Smaller increments reduce the blast radius of bugs and make CI failures easier to diagnose.
- The same principle applies to Stages 4 and 5 if they follow the same pattern.

### Consequences

- Each sub-stage produces its own SPEC and ADR.
- The PRD's "2–3 technologies per stage" guidance is preserved in aggregate (Stage 3a + 3b + 3c = 2–3 technologies) but each sub-stage is independently deliverable.
- Versioning and changelog entries are per sub-stage.

---

## ADR-031: Dockerfile as Dual-Definition Technology (Standalone + Composable)

### Context

Dockerfile-based projects fall into two distinct usage patterns:

1. **Container image projects** — the repository's primary purpose is to build and publish a Docker image (e.g. a base image, a utility container, a self-contained service image). The Dockerfile is the central artefact and lives at the project root.
2. **Containerised services** — Docker is a supporting concern within a larger project that has a primary technology (Go, Python, etc.). The Dockerfile packages the application for deployment but is not the project's primary artefact.

These patterns have different directory structures, different CI contexts, and different relationships to other technologies in the scaffold.

### Decision

Dockerfile is represented as two separate technology definitions, following the precedent set by Terraform in Stage 2 (ADR-024):

- **Dockerfile (Image)** — standalone. `Dockerfile`, `.dockerignore`, and `scripts/` at the project root.
- **Dockerfile (Service)** — composable. `Dockerfile` and `.dockerignore` nested under `docker/`.

### Rationale

- The dual-definition pattern is proven by Terraform (Module vs Infrastructure) and works cleanly with the existing standalone/composable model.
- A standalone container image project should not be mixed with other primary technologies — the Dockerfile is the project, not a supporting file.
- A composable Docker addition needs to coexist with other technologies without path conflicts. Nesting under `docker/` keeps the project root clean and mirrors the `infrastructure/` pattern used by Terraform Infrastructure.
- The `docker/` subdirectory was chosen over alternatives like `build/`, `container/`, or `deploy/` because it is the most widely recognised convention for Docker-related files in multi-technology projects.

### Consequences

- Two YAML files: `dockerfile-image.yaml` (standalone) and `dockerfile-service.yaml` (composable).
- Both share the same `ci.job_name: "docker"`. Since one is standalone and the other composable, they can never appear in the same scaffold, so no CI job name collision is possible.
- Both share the same devcontainer feature (`docker-in-docker:2`) and VS Code extension (`ms-azuretools.vscode-docker`). If both appeared together (which standalone prevents), the deduplication logic would handle this. In practice it's a non-issue.
- The composable variant's CI steps reference `docker/Dockerfile` rather than `Dockerfile`, matching the nested structure.

---

## ADR-032: Docker-in-Docker for Devcontainer

### Context

Building Docker images inside a devcontainer requires Docker access. Two devcontainer features are available:

1. **`docker-in-docker`** (`ghcr.io/devcontainers/features/docker-in-docker:2`) — runs a separate Docker daemon inside the devcontainer. Full isolation from the host Docker.
2. **`docker-outside-of-docker`** (`ghcr.io/devcontainers/features/docker-outside-of-docker:1`) — mounts the host's Docker socket into the container. Shares the host's Docker daemon.

### Decision

Use `docker-in-docker:2`.

### Rationale

- Full isolation — the devcontainer's Docker daemon is independent of the host. Images built inside the devcontainer don't pollute the host's image cache, and host Docker state doesn't leak into the development environment.
- Consistency — the devcontainer behaves identically regardless of the host's Docker configuration (Docker Desktop settings, running containers, network configuration).
- ARM compatibility — `docker-in-docker:2` works correctly on Apple Silicon (M-series) Macs, which is the primary development environment (ADR-008). The feature detects the host architecture and runs the appropriate daemon.
- `docker-outside-of-docker` has known issues with file ownership and permission mismatches when the host and container UIDs differ, which is common on macOS.
- The slight performance overhead of running a nested Docker daemon is acceptable for a development environment.

### Consequences

- The devcontainer feature `ghcr.io/devcontainers/features/docker-in-docker:2` is added to both Dockerfile technology definitions.
- No additional Dockerfile modification is needed — the feature installs Docker CLI and daemon automatically.
- Images built inside the devcontainer are ephemeral — they are lost when the devcontainer is destroyed. This is the intended behaviour for a development environment.

---

## ADR-033: Hadolint for Dockerfile Linting

### Context

Dockerfile linting catches common issues: missed best practices (not pinning package versions, using `latest` tag), security concerns (running as root), and style issues (redundant commands, missing cleanup). Several tools are available:

1. **Hadolint** — the de facto standard Dockerfile linter. Checks against a well-maintained rule set based on Dockerfile best practices.
2. **Docker Scout** — primarily a vulnerability scanning tool, not a linter.
3. **Dockle** — focused on CIS benchmarks and built image security, not Dockerfile source linting.

### Decision

Use Hadolint for Dockerfile linting in CI.

### Rationale

- Hadolint is the most widely adopted and mature Dockerfile linter. It integrates ShellCheck for linting the shell commands within `RUN` instructions, providing comprehensive coverage.
- It produces clear, actionable output with rule IDs (e.g. `DL3008: Pin versions in apt-get install`) that can be individually suppressed via inline comments when needed.
- Hadolint is available as a standalone binary, making CI installation straightforward — a single `wget` command with no runtime dependencies.
- Docker Scout and Dockle operate on built images, not Dockerfiles. They complement Hadolint but don't replace it.

### Consequences

- The CI `setup_steps` for both Dockerfile definitions install Hadolint as a standalone binary.
- The CI `lint_steps` run `hadolint Dockerfile` (standalone) or `hadolint docker/Dockerfile` (composable).
- Hadolint is not installed in the devcontainer — it is a CI-only tool. Developers who want local linting can install it themselves or rely on the VS Code Docker extension's built-in linting.

---

## ADR-034: docker build as Dockerfile Validation in CI

### Context

Beyond linting, the CI pipeline needs a validation step to confirm the Dockerfile actually produces a working image. Options considered:

1. **`docker build`** — builds the image, proving the Dockerfile is syntactically valid and all referenced files exist.
2. **`docker build --check`** — a newer Docker feature that validates without building. Not yet widely available.
3. **`docker compose build`** — requires a `docker-compose.yml` and adds unnecessary complexity.

### Decision

Use `docker build -t test-image .` (standalone) or `docker build -t test-image -f docker/Dockerfile .` (composable) as the CI test step.

### Rationale

- `docker build` is universally available on all CI runners with Docker installed (which `ubuntu-latest` includes by default).
- It validates the entire Dockerfile: syntax, base image availability, COPY/ADD source existence, and RUN command execution.
- The `-t test-image` tag is ephemeral — the image is built for validation only and discarded when the CI runner terminates.
- `--check` is not yet stable across all Docker versions available on CI runners.
- No `docker-compose.yml` is generated because not all Dockerfile projects need composition. Users who need it can add it themselves.

### Consequences

- The CI `test_steps` for both definitions run `docker build`.
- The composable variant uses `-f docker/Dockerfile .` to specify the Dockerfile path while keeping the build context at the project root. This ensures `COPY . .` in the Dockerfile has access to the full project.
- CI runners must have Docker installed. GitHub Actions `ubuntu-latest` includes Docker by default, so no additional setup is needed for the build step. Hadolint is the only tool requiring explicit installation.

---

## ADR-035: Empty setup.sh Contribution for Dockerfile

### Context

Every technology can contribute a block to the devcontainer's `setup.sh` via the `devcontainer.setup` field. The Dockerfile technology's devcontainer needs are handled entirely by the `docker-in-docker` feature — no additional setup commands are required.

### Decision

Both Dockerfile definitions set `setup: ""` (empty string). They do not contribute a block to `setup.sh`.

### Rationale

- The `docker-in-docker` feature handles Docker daemon and CLI installation as part of the devcontainer feature lifecycle. No additional post-create setup is needed.
- Adding an empty or no-op block (e.g. `# === Dockerfile (Service) ===` with no commands) would be confusing and add visual noise to the generated `setup.sh`.
- The existing engine logic already handles this correctly — `strings.TrimSpace(tech.Devcontainer.Setup)` returns empty, so no block is appended.

### Consequences

- When Dockerfile (Service) is combined with other technologies, the `setup.sh` contains blocks for the other technologies but no Dockerfile block.
- When Dockerfile (Image) is the only technology, the `setup.sh` contains only the base tooling block.

---

## ADR-036: Dockerfile Template Content

### Context

The generated `Dockerfile` needs starter content that is functional but generic enough to serve as a useful starting point for any container image project. The content must use template variables where appropriate.

### Decision

The generated Dockerfile uses `ubuntu:24.04` as the base image, installs minimal base packages (`ca-certificates`), sets a `WORKDIR`, and includes a `LABEL maintainer` that uses the `{{.ProjectName}}` template variable. The content is intentionally minimal — it is a scaffold, not application boilerplate.

### Rationale

- `ubuntu:24.04` is a well-known, stable, LTS base image that works on both AMD64 and ARM64 architectures. It provides a familiar environment for most users.
- Alpine was considered but rejected — while smaller, it uses `musl` instead of `glibc`, which causes compatibility issues with many compiled binaries. The opinionated choice of Ubuntu aligns with the tool's philosophy of encoding specific preferences (PRD Section 1).
- The `LABEL maintainer` uses the project name rather than asking for an email address. This keeps the scaffold simple and avoids adding a technology-driven prompt for a single label value.
- `ca-certificates` is included because HTTPS requests fail without it on a minimal Ubuntu install, which is a common gotcha for new Docker users.
- The `CMD ["/bin/bash"]` default is a safe fallback that lets users immediately `docker run -it` the image for testing. Users replace it with their actual entrypoint.

### Consequences

- The Dockerfile content is the same for both standalone and composable definitions. The only difference is the file location (root vs `docker/`).
- The `{{.ProjectName}}` template variable in the `LABEL` is resolved by the existing template resolution logic from Stage 2.
- Users are expected to modify the generated Dockerfile for their specific needs. The scaffold provides structure and a working starting point, not a production-ready image.

---

## ADR-037: .dockerignore as Technology-Owned File

### Context

A `.dockerignore` file controls which files are excluded from the Docker build context. It could be generated as a universal output (like `.gitignore`) or as a technology-specific file within the Dockerfile definition.

### Decision

The `.dockerignore` is defined in the Dockerfile technology's `structure` entries with explicit content. It is a technology-owned file, not a universal output.

### Rationale

- Not every project uses Docker. Generating `.dockerignore` universally would produce a file with no purpose in non-Docker projects.
- The `.dockerignore` content is specific to the Docker workflow — it excludes development-only files (`.devcontainer`, `.git`, Makefiles) that should not be in the build context.
- Placing it in the technology definition keeps Docker-related configuration self-contained within the Dockerfile YAML, consistent with how each technology owns its files.
- Unlike `.gitignore` (which is composed from multiple technologies), `.dockerignore` is singular — there is only one Docker build context per project, so composition is not needed.

### Consequences

- Both the standalone and composable definitions place `.dockerignore` at the project root. This is correct for both cases: the standalone variant builds from root, and the composable variant uses `docker build -f docker/Dockerfile .` which also uses the project root as the build context.
- No engine changes are needed — the `.dockerignore` is created like any other structure entry with content.
- No path conflict exists between the two definitions' `.dockerignore` files because one is standalone and the other composable — they can never appear in the same scaffold.