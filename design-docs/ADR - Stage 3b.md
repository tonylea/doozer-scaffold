# Architecture Decision Record — Stage 3b

**Date:** 2026-03-11  
**Status:** Accepted  
**Related SPEC:** SPEC Stage 3b v2.1

---

## ADR-038: Technology Variant Groups with Auto-Selection

### Context

Technologies like Terraform (Module vs Infrastructure), Dockerfile (Image vs Service), and now Helm (Chart vs Deployment) follow a dual-definition pattern: one standalone variant and one composable variant. This creates two YAML files per technology and two entries in the technology prompt, forcing the user to choose between them. But the correct choice is always deterministic:

- If the technology is the **only** technology selected → use the standalone variant.
- If it is selected **alongside** other technologies → use the composable variant.

Asking the user to choose is unnecessary friction. The tool already has the information needed to decide.

### Decision

The technology definition schema gains a `variant_group` field. Two definitions that share the same `variant_group` value are variants of the same technology. The engine and prompt use this field to:

1. **Prompt:** Present one entry per `variant_group` instead of two. The display name is the `variant_group` value (e.g. "Helm"). Technologies without `variant_group` are presented as before.
2. **Selection:** After the user selects technologies, the engine resolves each `variant_group` to the correct definition. If the technology is the sole selection, pick the standalone variant (`standalone: true`). If other technologies are also selected, pick the composable variant (`standalone: false`).
3. **Validation:** A `variant_group` must contain exactly one standalone and one composable definition. If not, validation fails on load.

The two YAML files remain separate. Humans can open `helm-chart.yaml` and see the standalone layout, open `helm-deployment.yaml` and see the composable layout. The auto-selection logic is entirely in the engine.

All existing dual-mode technologies (Terraform, Dockerfile) are updated to use variant groups in this stage alongside Helm. Problems are solved when the mechanism is introduced, not deferred.

### Rationale

- Two files remain human-readable. Merging into one file would create a complex, harder-to-reason-about definition.
- The selection logic is data-driven and universal — it works for any current or future dual-mode technology. No technology-specific conditionals.
- Applying `variant_group` to all dual-mode technologies immediately avoids tech debt. If the mechanism exists but only some technologies use it, the codebase is inconsistent and the debt becomes invisible.
- This forwards the project's modular, data-driven goal: the engine becomes smarter about technology resolution without any technology-specific code paths.

### Consequences

- The YAML schema gains `variant_group` (optional string). When present, two definitions share the group.
- The prompt collapses variant groups into single entries. The display name is the `variant_group` value.
- The engine resolves variant groups after technology selection, before generation. The resolved definitions are passed to `Generate` as before.
- The `standalone` boolean remains meaningful — it distinguishes which variant is which within a group.
- Technology-driven prompts may be variant-specific. The `prompts` field gains an optional `mode` qualifier (`"standalone"` or `"composable"`) so prompts can be scoped to a specific variant. Prompts without `mode` are always presented. For Helm, `chart_name` has `mode: "composable"` because the standalone variant uses `ProjectName` directly.
- Config validation must be updated: when a `variant_group` technology is selected and other technologies are also selected, the composable variant must exist. When it is the sole selection, the standalone variant must exist.
- Existing definitions updated in this stage:
  - `terraform-module.yaml`: add `variant_group: "Terraform"`
  - `terraform-infrastructure.yaml`: add `variant_group: "Terraform"`
  - `dockerfile-image.yaml`: add `variant_group: "Dockerfile"`
  - `dockerfile-service.yaml`: add `variant_group: "Dockerfile"`
- Existing acceptance tests for Terraform and Dockerfile must be updated to work through variant group resolution rather than selecting specific definition keys.

---

## ADR-039: deploy/helm/ as Composable Chart Path

### Context

The composable Helm chart needs a subdirectory path within the project. Several conventions exist:

1. `charts/` — Helm's own convention for subchart dependencies within a chart.
2. `chart/` — used in some single-chart repos.
3. `helm/` — simple but ambiguous at the project root.
4. `deploy/helm/` — groups deployment tooling under a `deploy/` namespace.

### Decision

Use `deploy/helm/{{.chart_name}}/` for the composable variant.

### Rationale

- `charts/` is a reserved directory name within the Helm chart format — it holds subchart dependencies. Using it at the project root creates ambiguity and could interfere with tooling.
- `deploy/helm/` clearly communicates purpose. The nested `{{.chart_name}}/` directory matches the Helm convention that the chart directory name matches the chart name.
- The `deploy/` prefix is extensible. Future deployment technologies could use `deploy/kustomize/`, `deploy/pulumi/`, etc.

### Path Conflict Analysis

All composable structure files live under `deploy/helm/{{.chart_name}}/`. No existing composable technology defines paths under `deploy/`:

| Composable Technology      | Paths used                                           | Conflict? |
| -------------------------- | ---------------------------------------------------- | --------- |
| Go                         | `cmd/app/`, `internal/`                              | No        |
| Python                     | `src/{{.package_name}}/`, `tests/`, `pyproject.toml` | No        |
| Terraform (Infrastructure) | `infrastructure/`                                    | No        |
| Dockerfile (Service)       | `docker/`, `.dockerignore`                           | No        |

Python's `tests/` at project root does not conflict with `deploy/helm/{{.chart_name}}/tests/`.

The CI `job_name` `"helm"` is unique among composable technologies (`go`, `python`, `terraform`, `docker`).

The prompt key `chart_name` is unique across all technologies (`package_name` for Python).

The `ms-kubernetes-tools.vscode-kubernetes-tools` extension is unique to Helm.

### Consequences

- Composable structure paths live under `deploy/helm/{{.chart_name}}/`.
- CI steps reference `deploy/helm/{{.chart_name}}/` in composable mode, `.` in standalone mode.

---

## ADR-040: Standard Helm Chart Structure from helm create

### Context

The generated Helm chart needs a directory structure and starter files.

### Decision

Use the standard `helm create` structure: `Chart.yaml`, `values.yaml`, `.helmignore`, `templates/` with `deployment.yaml`, `service.yaml`, `_helpers.tpl`, `NOTES.txt`, `serviceaccount.yaml`, `hpa.yaml`, `ingress.yaml`, and `templates/tests/test-connection.yaml`. The `charts/` subdirectory is included empty. A chart-root `tests/` directory contains `helm-unittest` test files.

### Rationale

- `helm create` produces the community-standard structure. Deviating from it would confuse Helm developers.
- Templates use Helm best practices: helper templates for naming/labels, conditional resource creation, parameterised values.
- `templates/tests/test-connection.yaml` provides a working `helm test` target out of the box.
- The chart-root `tests/` directory holds `helm-unittest` YAML files (offline template rendering tests). This is a separate concern from `templates/tests/` (runtime test hooks). Both are standard Helm community convention.

### Consequences

- Standalone and composable variants generate identical internal chart layout. Only the root path differs.
- `Chart.yaml` uses `{{.ProjectName}}` (standalone) or `{{.chart_name}}` (composable) for the chart name.
- `values.yaml` provides sensible defaults (nginx image, 1 replica, ClusterIP service) for a deployable chart.

---

## ADR-041: Helm Template Content Escaping Strategy

### Context

Helm charts contain Go template expressions (`{{ .Values.image.repository }}`, `{{ include "mychart.fullname" . }}`) using the same `{{ }}` delimiters as the scaffold engine's `text/template` processor. The engine would attempt to evaluate Helm expressions, causing errors.

This is a new challenge — no previous technology generates files containing Go template syntax.

### Decision

Helm template content in the YAML definition uses `{{"{{"}}` and `{{"}}"}}` to escape Helm expressions through the scaffold engine. The engine evaluates `{{"{{"}}` to produce the literal string `{{`, written to the output file as intended Helm syntax.

Example in YAML definition:

```
image: "{{"{{"}} .Values.image.repository {{"}}"}}:{{"{{"}} .Values.image.tag | default .Chart.AppVersion {{"}}"}}"
```

Produces in output file:

```
image: "{{ .Values.image.repository }}:{{ .Values.image.tag | default .Chart.AppVersion }}"
```

### Rationale

- Standard Go `text/template` escaping. No engine changes required.
- Self-contained within the YAML definition. No special flags or engine modifications.
- Scaffold variables (`{{.ProjectName}}`, `{{.chart_name}}`) work normally alongside escaped Helm expressions.
- Alternatives rejected: disabling template processing (breaks scaffold variable substitution), alternative delimiters (requires engine changes), post-processing (fragile).

### Consequences

- YAML definitions are more verbose due to escaping. One-time authoring cost.
- Engine requires no changes for escaping — only the variant group mechanism changes the engine.
- All Helm template expressions render correctly in output.

---

## ADR-042: helm lint and helm unittest for CI

### Context

Helm charts need linting and testing in CI. Available tools: `helm lint` (built-in validation), `ct lint` (chart-testing, designed for monorepos), `helm unittest` (BDD-style template rendering tests), `kubeconform` (Kubernetes API schema validation).

### Decision

Use `helm lint` for CI lint and `helm unittest` for CI test. `ct` is not used.

### Rationale

- `helm lint` validates chart structure, YAML syntax, and best practices. No installation beyond Helm itself.
- `helm unittest` validates template rendering logic — the most important correctness check. Catches logic errors in conditionals, loops, and value references.
- `ct` is designed for monorepo workflows with multiple charts. Overkill for a single chart. Adds Python as a dependency.
- `kubeconform` is valuable but requires specifying a Kubernetes version. Users who need it can add it.
- The scaffold also generates `templates/tests/test-connection.yaml` for `helm test` (runtime testing against a live cluster), complementary to `helm unittest` but not run in CI.

### Consequences

- CI `setup_steps` install Helm and the `helm-unittest` plugin.
- CI `lint_steps` run `helm lint`.
- CI `test_steps` run `helm unittest`.

---

## ADR-043: Helm Devcontainer Setup

### Context

Helm needs devcontainer support. It could be installed as a devcontainer feature, via setup script, or assumed present.

### Decision

Install Helm via the `setup` script using the official install script. Install `helm-unittest` in the same block. No devcontainer feature. Add the `ms-kubernetes-tools.vscode-kubernetes-tools` VS Code extension.

### Rationale

- No official devcontainer feature for Helm exists in the `devcontainers/features` registry.
- The official install script handles architecture detection (ARM/AMD64) automatically.
- Installing `helm-unittest` ensures developers can run chart tests locally.
- The Kubernetes Tools extension provides Helm template syntax highlighting.

### Consequences

- `devcontainer.setup` contains Helm and unittest plugin install commands.
- No devcontainer feature contributed.
- VS Code Kubernetes Tools extension added.

---

## ADR-044: .helmignore as Technology-Owned File

### Context

A `.helmignore` controls which files are excluded when packaging a Helm chart. It could be a universal output or technology-specific.

### Decision

`.helmignore` is defined in the Helm technology's `structure` entries with explicit content. For standalone it lives at the project root. For composable it lives inside `deploy/helm/{{.chart_name}}/`.

### Rationale

- Not every project uses Helm. Universal generation would produce a purposeless file.
- Content is Helm-workflow-specific. Consistent with `.dockerignore` (ADR-037).
- Keeps Helm configuration self-contained in the YAML definition, consistent with ADR-001.

### Consequences

- Both variants include `.helmignore` at the appropriate path.