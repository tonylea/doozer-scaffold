# SPEC — doozer-scaffold Stage 3a: Dockerfile Technology

**Version:** 1.0  
**Status:** Draft  
**Parent PRD:** doozer-scaffold PRD v0.2-draft  
**Related ADR:** ADR Stage 1, ADR Stage 2, ADR Stage 3a  
**Owner:** Tony Lea

---

## 1. Purpose

This document is the complete technical specification for Stage 3a of `doozer-scaffold`. It defines every input, output, behaviour, and test expectation in sufficient detail that an implementer can build the stage without ambiguity.

Stage 3a adds Dockerfile as a supported technology with two definitions — a standalone container image project and a composable Docker addition — following the same dual-definition pattern established by Terraform in Stage 2. No schema changes, engine changes, or prompt changes are required. The existing declarative technology definition system handles everything.

---

## 2. Summary of Changes from Stage 2

### 2.1 Schema Changes

None. The Stage 2 schema supports all fields required by the Dockerfile definitions.

### 2.2 Config Changes

None.

### 2.3 Prompt Changes

None. The new technology definitions are automatically picked up by the dynamic technology prompt.

### 2.4 Engine Changes

None. The scaffold engine is entirely data-driven. The new YAML definitions are processed by the existing engine without modification.

### 2.5 New Definitions

| Definition           | File                                   | Mode       | Detail                                                               |
| -------------------- | -------------------------------------- | ---------- | -------------------------------------------------------------------- |
| Dockerfile (Image)   | `technologies/dockerfile-image.yaml`   | Standalone | A standalone container image project. Dockerfile and config at root. |
| Dockerfile (Service) | `technologies/dockerfile-service.yaml` | Composable | Docker as supporting infrastructure. Files nested under `docker/`.   |

---

## 3. Technology Definitions

### 3.1 Dockerfile (Image)

File: `technologies/dockerfile-image.yaml`

This represents a standalone container image project — a repository whose primary purpose is to build and publish a Docker image (e.g. a base image, a utility image, or a self-contained service image).

```yaml
name: "Dockerfile (Image)"
standalone: true

structure:
  - path: "Dockerfile"
    content: |
      FROM ubuntu:24.04

      LABEL maintainer="{{.ProjectName}}"

      RUN apt-get update && apt-get install -y --no-install-recommends \
          ca-certificates \
          && rm -rf /var/lib/apt/lists/*

      WORKDIR /app

      COPY . .

      CMD ["/bin/bash"]
  - path: ".dockerignore"
    content: |
      .git
      .github
      .devcontainer
      .editorconfig
      .gitattributes
      .gitignore
      *.md
      LICENSE
      Makefile
  - path: "scripts/"

gitignore: |
  # Docker
  .docker/

devcontainer:
  features:
    "ghcr.io/devcontainers/features/docker-in-docker:2": {}
  extensions:
    - "ms-azuretools.vscode-docker"
  setup: ""

ci:
  job_name: "docker"
  setup_steps:
    - name: "Install Hadolint"
      run: |
        sudo wget -O /usr/local/bin/hadolint https://github.com/hadolint/hadolint/releases/latest/download/hadolint-Linux-x86_64
        sudo chmod +x /usr/local/bin/hadolint
  lint_steps:
    - name: "Lint Dockerfile"
      run: "hadolint Dockerfile"
  test_steps:
    - name: "Build image"
      run: "docker build -t test-image ."
```

This produces:

```
{project}/
├── .dockerignore
├── Dockerfile
└── scripts/
    └── .gitkeep
```

### 3.2 Dockerfile (Service)

File: `technologies/dockerfile-service.yaml`

This represents Docker as a supporting concern within a larger project — the project has a primary technology (Go, Python, etc.) and uses Docker for containerisation. Files are nested under a `docker/` subdirectory to keep the project root clean and avoid conflicting with the primary technology's files.

```yaml
name: "Dockerfile (Service)"
standalone: false

structure:
  - path: "docker/"
  - path: "docker/Dockerfile"
    content: |
      FROM ubuntu:24.04

      LABEL maintainer="{{.ProjectName}}"

      RUN apt-get update && apt-get install -y --no-install-recommends \
          ca-certificates \
          && rm -rf /var/lib/apt/lists/*

      WORKDIR /app

      COPY . .

      CMD ["/bin/bash"]
  - path: ".dockerignore"
    content: |
      .git
      .github
      .devcontainer
      .editorconfig
      .gitattributes
      .gitignore
      *.md
      LICENSE
      Makefile

gitignore: |
  # Docker
  .docker/

devcontainer:
  features:
    "ghcr.io/devcontainers/features/docker-in-docker:2": {}
  extensions:
    - "ms-azuretools.vscode-docker"
  setup: ""

ci:
  job_name: "docker"
  setup_steps:
    - name: "Install Hadolint"
      run: |
        sudo wget -O /usr/local/bin/hadolint https://github.com/hadolint/hadolint/releases/latest/download/hadolint-Linux-x86_64
        sudo chmod +x /usr/local/bin/hadolint
  lint_steps:
    - name: "Lint Dockerfile"
      run: "hadolint docker/Dockerfile"
  test_steps:
    - name: "Build image"
      run: "docker build -t test-image -f docker/Dockerfile ."
```

This produces:

```
{project}/
├── .dockerignore
└── docker/
    ├── .gitkeep
    └── Dockerfile
```

Note: The `.dockerignore` is placed at the project root (not under `docker/`) because `docker build -f docker/Dockerfile .` uses the project root as the build context, and Docker resolves `.dockerignore` relative to the build context root. The `docker/` directory entry creates a `.gitkeep`. The Dockerfile file entry also creates a file in the same directory. Both are created — `.gitkeep` from the directory entry, `Dockerfile` from the file entry. This is not a conflict — they are different filenames in the same directory.

---

## 4. Path Conflict Analysis

The Dockerfile definitions must not conflict with any existing technology's structure entries. This section documents the analysis.

### 4.1 Dockerfile (Image) — Standalone

As a standalone technology, it can never be combined with other technologies. Path conflicts with other technologies are therefore impossible by design.

### 4.2 Dockerfile (Service) — Composable

Structure files are under `docker/` except for `.dockerignore` which is at the project root (required for Docker build context). No existing composable technology defines `.dockerignore` or paths under `docker/`:

| Composable Technology      | Paths used                                           | Conflict? |
| -------------------------- | ---------------------------------------------------- | --------- |
| Go                         | `cmd/app/`, `internal/`                              | No        |
| Python                     | `src/{{.package_name}}/`, `tests/`, `pyproject.toml` | No        |
| Terraform (Infrastructure) | `infrastructure/`                                    | No        |

The `.dockerignore` at root is unique to Dockerfile (Service) — no other composable technology produces this file.

### 4.3 CI Job Name Conflict

Both Dockerfile definitions use `job_name: "docker"`. Since one is standalone and the other composable, they can never appear in the same scaffold. No CI job name conflict is possible.

However, the Dockerfile (Service) `ci.job_name` of `"docker"` must not conflict with any other composable technology's `ci.job_name`:

| Composable Technology      | CI job_name | Conflict? |
| -------------------------- | ----------- | --------- |
| Go                         | `go`        | No        |
| Python                     | `python`    | No        |
| Terraform (Infrastructure) | `terraform` | No        |

---

## 5. User Interaction Flow

No changes to the prompt sequence. The new technology definitions are automatically populated into the technology multi-select prompt via the existing dynamic loading mechanism.

After Stage 3a, the technology prompt displays (alphabetical order by `name`):

- Dockerfile (Image) — standalone
- Dockerfile (Service) — composable
- Go — composable
- PowerShell Module — standalone
- Python — composable
- Terraform (Infrastructure) — composable
- Terraform Module — standalone

The standalone enforcement logic from Stage 2 applies unchanged:

- Selecting "Dockerfile (Image)" locks out all other technologies.
- Selecting "Dockerfile (Service)" alongside other composable technologies is permitted.
- Selecting "Dockerfile (Service)" alongside a standalone technology is prevented.

---

## 6. Output Specification

### 6.1 Complete Output Tree — Dockerfile (Image), Maximum Selections

When selecting Dockerfile (Image) with all optional selections (MIT licence, CONTRIBUTING.md, both tooling files, all GitHub repo config):

```
my-image/
├── .devcontainer/
│   ├── devcontainer.json
│   ├── Dockerfile
│   └── setup.sh
├── .dockerignore
├── .editorconfig
├── .gitattributes
├── .github/
│   ├── dependabot.yml
│   ├── ISSUE_TEMPLATE/
│   │   ├── bug_report.yaml
│   │   └── feature_request.yaml
│   ├── pull_request_template.md
│   └── workflows/
│       └── ci.yml
├── .gitignore
├── CONTRIBUTING.md
├── Dockerfile
├── LICENSE
├── Makefile
├── README.md
└── scripts/
    └── .gitkeep
```

### 6.2 Complete Output Tree — Dockerfile (Image), Minimum Selections

```
my-image/
├── .devcontainer/
│   ├── devcontainer.json
│   ├── Dockerfile
│   └── setup.sh
├── .dockerignore
├── .github/
│   └── workflows/
│       └── ci.yml
├── .gitignore
├── Dockerfile
├── Makefile
├── README.md
└── scripts/
    └── .gitkeep
```

### 6.3 Complete Output Tree — Go + Dockerfile (Service), Maximum Selections

When selecting Go + Dockerfile (Service) with all optional selections:

```
my-project/
├── .devcontainer/
│   ├── devcontainer.json
│   ├── Dockerfile
│   └── setup.sh
├── .dockerignore
├── .editorconfig
├── .gitattributes
├── .github/
│   ├── dependabot.yml
│   ├── ISSUE_TEMPLATE/
│   │   ├── bug_report.yaml
│   │   └── feature_request.yaml
│   ├── pull_request_template.md
│   └── workflows/
│       └── ci.yml
├── .gitignore
├── cmd/
│   └── app/
│       └── .gitkeep
├── CONTRIBUTING.md
├── docker/
│   ├── .gitkeep
│   └── Dockerfile
├── internal/
│   └── .gitkeep
├── LICENSE
├── Makefile
└── README.md
```

### 6.4 Complete Output Tree — Go + Python + Dockerfile (Service) + Terraform (Infrastructure), Minimum Selections

Assume `package_name` prompt response is `my_project`:

```
my-project/
├── .devcontainer/
│   ├── devcontainer.json
│   ├── Dockerfile
│   └── setup.sh
├── .dockerignore
├── .github/
│   └── workflows/
│       └── ci.yml
├── .gitignore
├── cmd/
│   └── app/
│       └── .gitkeep
├── docker/
│   ├── .gitkeep
│   └── Dockerfile
├── infrastructure/
│   ├── .gitkeep
│   ├── main.tf
│   ├── outputs.tf
│   ├── variables.tf
│   └── versions.tf
├── internal/
│   └── .gitkeep
├── Makefile
├── pyproject.toml
├── README.md
├── src/
│   └── my_project/
│       ├── .gitkeep
│       └── __init__.py
└── tests/
    ├── .gitkeep
    └── __init__.py
```

### 6.5 Composite Outputs for Go + Dockerfile (Service)

#### .gitignore

Sections in alphabetical order by technology name:

```
# Dockerfile (Service)
# Docker
.docker/

# Go
# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
# Test binary
*.test
# Output
*.out
# Go workspace
go.work
go.work.sum
# Dependency directory
vendor/
# Build output
bin/
dist/
```

#### devcontainer.json

Features merged from base (Node.js) + Docker-in-Docker + Go:

```json
{
    "name": "my-project",
    "build": {
        "dockerfile": "Dockerfile"
    },
    "features": {
        "ghcr.io/devcontainers/features/docker-in-docker:2": {},
        "ghcr.io/devcontainers/features/go:1": {},
        "ghcr.io/devcontainers/features/node:1": {}
    },
    "customizations": {
        "vscode": {
            "extensions": [
                "golang.go",
                "ms-azuretools.vscode-docker"
            ]
        }
    },
    "postCreateCommand": "bash .devcontainer/setup.sh"
}
```

Note: Features are ordered alphabetically by key. Extensions are merged, deduplicated, and sorted alphabetically. This is the existing Stage 2 behaviour.

#### setup.sh

Base block first, then technology blocks in alphabetical order by technology key. Since Dockerfile (Service) has an empty `setup` field, it does not contribute a setup.sh block:

```bash
#!/bin/bash
set -e

# === Base tooling ===
npm install -g markdownlint-cli2 @commitlint/cli @commitlint/config-conventional

# === Go ===
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

#### ci.yml

Example for Go + Dockerfile (Service):

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  lint-docker:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install Hadolint
        run: |
          sudo wget -O /usr/local/bin/hadolint https://github.com/hadolint/hadolint/releases/latest/download/hadolint-Linux-x86_64
          sudo chmod +x /usr/local/bin/hadolint
      - name: Lint Dockerfile
        run: hadolint docker/Dockerfile

  lint-go:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: stable
      - name: Lint
        run: |
          go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
          golangci-lint run

  test-docker:
    runs-on: ubuntu-latest
    needs: [lint-docker, lint-go]
    steps:
      - uses: actions/checkout@v4
      - name: Install Hadolint
        run: |
          sudo wget -O /usr/local/bin/hadolint https://github.com/hadolint/hadolint/releases/latest/download/hadolint-Linux-x86_64
          sudo chmod +x /usr/local/bin/hadolint
      - name: Build image
        run: docker build -t test-image -f docker/Dockerfile .

  test-go:
    runs-on: ubuntu-latest
    needs: [lint-docker, lint-go]
    steps:
      - uses: actions/checkout@v4
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: stable
      - name: Test
        run: go test ./... -v

  build:
    runs-on: ubuntu-latest
    needs: [test-docker, test-go]
    steps:
      - uses: actions/checkout@v4
      - name: Build
        run: echo 'TODO: Add build steps'
```

---

## 7. Testing Strategy

### 7.1 Principles

Unchanged from Stage 2. All development follows strict TDD.

### 7.2 No Existing Test Updates Required

Stage 3a adds new technology definitions only. No schema, engine, config, or prompt changes are made. All existing Stage 1 and Stage 2 tests continue to pass without modification.

### 7.3 New Unit Tests

#### Definition loading and validation

```go
func TestDockerfileImageDefinitionLoads(t *testing.T) {
    defs, err := techdef.Load()
    require.NoError(t, err)
    require.Contains(t, defs, "dockerfile-image")

    def := defs["dockerfile-image"]
    assert.Equal(t, "Dockerfile (Image)", def.Name)
    assert.True(t, def.Standalone)
    assert.NotEmpty(t, def.Structure)
    assert.NotEmpty(t, def.Gitignore)
    assert.NotNil(t, def.CI)
    assert.Equal(t, "docker", def.CI.JobName)
    assert.Empty(t, def.Prompts)
}

func TestDockerfileServiceDefinitionLoads(t *testing.T) {
    defs, err := techdef.Load()
    require.NoError(t, err)
    require.Contains(t, defs, "dockerfile-service")

    def := defs["dockerfile-service"]
    assert.Equal(t, "Dockerfile (Service)", def.Name)
    assert.False(t, def.Standalone)
    assert.NotEmpty(t, def.Structure)
    assert.NotEmpty(t, def.Gitignore)
    assert.NotNil(t, def.CI)
    assert.Equal(t, "docker", def.CI.JobName)
    assert.Empty(t, def.Prompts)
}
```

#### Standalone and composable constraints

```go
func TestDockerfileImageIsStandalone(t *testing.T) {
    defs, err := techdef.Load()
    require.NoError(t, err)
    assert.True(t, defs["dockerfile-image"].Standalone)
}

func TestDockerfileServiceIsComposable(t *testing.T) {
    defs, err := techdef.Load()
    require.NoError(t, err)
    assert.False(t, defs["dockerfile-service"].Standalone)
}

func TestDockerfileImageCannotCombineWithOtherTechs(t *testing.T) {
    defs, _ := techdef.Load()
    cfg := &config.Config{
        ProjectName:  "test",
        Technologies: []string{"dockerfile-image", "go"},
    }
    err := cfg.Validate(defs)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "standalone")
}

func TestDockerfileServiceCanCombineWithGo(t *testing.T) {
    defs, _ := techdef.Load()
    cfg := &config.Config{
        ProjectName:  "test",
        Technologies: []string{"dockerfile-service", "go"},
    }
    err := cfg.Validate(defs)
    assert.NoError(t, err)
}

func TestDockerfileServiceCanCombineWithMultipleTechs(t *testing.T) {
    defs, _ := techdef.Load()
    cfg := &config.Config{
        ProjectName:  "test",
        Technologies: []string{"dockerfile-service", "go", "python", "terraform-infrastructure"},
        TechPromptResponses: map[string]string{"package_name": "test_app"},
    }
    err := cfg.Validate(defs)
    assert.NoError(t, err)
}
```

#### Structure entry validation

```go
func TestDockerfileImageStructureHasDockerfile(t *testing.T) {
    defs, err := techdef.Load()
    require.NoError(t, err)

    def := defs["dockerfile-image"]
    paths := make([]string, len(def.Structure))
    for i, entry := range def.Structure {
        paths[i] = entry.Path
    }
    assert.Contains(t, paths, "Dockerfile")
    assert.Contains(t, paths, ".dockerignore")
    assert.Contains(t, paths, "scripts/")
}

func TestDockerfileServiceStructureHasCorrectPaths(t *testing.T) {
    defs, err := techdef.Load()
    require.NoError(t, err)

    def := defs["dockerfile-service"]
    paths := make([]string, len(def.Structure))
    for i, entry := range def.Structure {
        paths[i] = entry.Path
    }
    assert.Contains(t, paths, "docker/")
    assert.Contains(t, paths, "docker/Dockerfile")
    assert.Contains(t, paths, ".dockerignore")
}
```

#### Devcontainer features

```go
func TestDockerfileDefinitionsHaveDockerInDockerFeature(t *testing.T) {
    defs, err := techdef.Load()
    require.NoError(t, err)

    for _, key := range []string{"dockerfile-image", "dockerfile-service"} {
        def := defs[key]
        _, ok := def.Devcontainer.Features["ghcr.io/devcontainers/features/docker-in-docker:2"]
        assert.True(t, ok, "%s should have docker-in-docker feature", key)
    }
}

func TestDockerfileDefinitionsHaveDockerExtension(t *testing.T) {
    defs, err := techdef.Load()
    require.NoError(t, err)

    for _, key := range []string{"dockerfile-image", "dockerfile-service"} {
        def := defs[key]
        assert.Contains(t, def.Devcontainer.Extensions, "ms-azuretools.vscode-docker",
            "%s should have Docker VS Code extension", key)
    }
}
```

#### No path conflicts with existing composable techs

```go
func TestDockerfileServiceNoPathConflictWithGo(t *testing.T) {
    defs, _ := techdef.Load()
    techs := []*techdef.TechDef{defs["dockerfile-service"], defs["go"]}
    data := map[string]string{"ProjectName": "test"}
    err := scaffold.DetectPathConflicts(techs, data)
    assert.NoError(t, err)
}

func TestDockerfileServiceNoPathConflictWithPython(t *testing.T) {
    defs, _ := techdef.Load()
    techs := []*techdef.TechDef{defs["dockerfile-service"], defs["python"]}
    data := map[string]string{"ProjectName": "test", "package_name": "test_app"}
    err := scaffold.DetectPathConflicts(techs, data)
    assert.NoError(t, err)
}

func TestDockerfileServiceNoPathConflictWithTerraformInfra(t *testing.T) {
    defs, _ := techdef.Load()
    techs := []*techdef.TechDef{defs["dockerfile-service"], defs["terraform-infrastructure"]}
    data := map[string]string{"ProjectName": "test"}
    err := scaffold.DetectPathConflicts(techs, data)
    assert.NoError(t, err)
}

func TestDockerfileServiceNoPathConflictWithAllComposable(t *testing.T) {
    defs, _ := techdef.Load()
    techs := []*techdef.TechDef{
        defs["dockerfile-service"],
        defs["go"],
        defs["python"],
        defs["terraform-infrastructure"],
    }
    data := map[string]string{"ProjectName": "test", "package_name": "test_app"}
    err := scaffold.DetectPathConflicts(techs, data)
    assert.NoError(t, err)
}
```

### 7.4 Acceptance Tests

#### Acceptance test: Dockerfile (Image) standalone, maximum selections

```go
func TestAcceptance_DockerfileImage_MaxSelections(t *testing.T) {
    defs, err := techdef.Load()
    require.NoError(t, err)

    cfg := &config.Config{
        ProjectName:  "my-image",
        Provider:     "github",
        Technologies: []string{"dockerfile-image"},
        Licence:      "mit",
        Docs:         []string{"contributing"},
        Tooling:      []string{"editorconfig", "gitattributes"},
        RepoConfig:   []string{"issue_templates", "pr_template", "dependabot"},
        Confirmed:    true,
    }

    baseDir := t.TempDir()
    techs := []*techdef.TechDef{defs["dockerfile-image"]}
    err = scaffold.Generate(cfg, techs, baseDir)
    require.NoError(t, err)

    root := filepath.Join(baseDir, "my-image")

    // Verify exact file tree
    var files []string
    filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        rel, _ := filepath.Rel(root, path)
        if rel != "." {
            files = append(files, rel)
        }
        return nil
    })

    expected := []string{
        ".devcontainer",
        ".devcontainer/Dockerfile",
        ".devcontainer/devcontainer.json",
        ".devcontainer/setup.sh",
        ".dockerignore",
        ".editorconfig",
        ".gitattributes",
        ".github",
        ".github/ISSUE_TEMPLATE",
        ".github/ISSUE_TEMPLATE/bug_report.yaml",
        ".github/ISSUE_TEMPLATE/feature_request.yaml",
        ".github/dependabot.yml",
        ".github/pull_request_template.md",
        ".github/workflows",
        ".github/workflows/ci.yml",
        ".gitignore",
        "CONTRIBUTING.md",
        "Dockerfile",
        "LICENSE",
        "Makefile",
        "README.md",
        "scripts",
        "scripts/.gitkeep",
    }
    assert.ElementsMatch(t, expected, files)

    // Verify Dockerfile content
    dockerfile, _ := os.ReadFile(filepath.Join(root, "Dockerfile"))
    assert.Contains(t, string(dockerfile), "FROM ubuntu:24.04")
    assert.Contains(t, string(dockerfile), "my-image")

    // Verify .dockerignore content
    dockerignore, _ := os.ReadFile(filepath.Join(root, ".dockerignore"))
    assert.Contains(t, string(dockerignore), ".git")
    assert.Contains(t, string(dockerignore), ".devcontainer")

    // Verify .gitignore has Docker section
    gitignore, _ := os.ReadFile(filepath.Join(root, ".gitignore"))
    assert.Contains(t, string(gitignore), "# Dockerfile (Image)")
    assert.Contains(t, string(gitignore), ".docker/")

    // Verify CI has docker jobs
    ci, _ := os.ReadFile(filepath.Join(root, ".github/workflows/ci.yml"))
    assert.Contains(t, string(ci), "lint-docker")
    assert.Contains(t, string(ci), "test-docker")
    assert.Contains(t, string(ci), "hadolint")
    assert.Contains(t, string(ci), "docker build")
    assert.Contains(t, string(ci), "build")

    // Verify devcontainer has docker-in-docker feature
    dcJson, _ := os.ReadFile(filepath.Join(root, ".devcontainer/devcontainer.json"))
    assert.Contains(t, string(dcJson), "docker-in-docker")
    assert.Contains(t, string(dcJson), "ms-azuretools.vscode-docker")
}
```

#### Acceptance test: Dockerfile (Image) standalone, minimum selections

```go
func TestAcceptance_DockerfileImage_MinSelections(t *testing.T) {
    defs, err := techdef.Load()
    require.NoError(t, err)

    cfg := &config.Config{
        ProjectName:  "my-image",
        Provider:     "github",
        Technologies: []string{"dockerfile-image"},
        Licence:      "none",
        Docs:         []string{},
        Tooling:      []string{},
        RepoConfig:   []string{},
        Confirmed:    true,
    }

    baseDir := t.TempDir()
    techs := []*techdef.TechDef{defs["dockerfile-image"]}
    err = scaffold.Generate(cfg, techs, baseDir)
    require.NoError(t, err)

    root := filepath.Join(baseDir, "my-image")

    var files []string
    filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        rel, _ := filepath.Rel(root, path)
        if rel != "." {
            files = append(files, rel)
        }
        return nil
    })

    expected := []string{
        ".devcontainer",
        ".devcontainer/Dockerfile",
        ".devcontainer/devcontainer.json",
        ".devcontainer/setup.sh",
        ".dockerignore",
        ".github",
        ".github/workflows",
        ".github/workflows/ci.yml",
        ".gitignore",
        "Dockerfile",
        "Makefile",
        "README.md",
        "scripts",
        "scripts/.gitkeep",
    }
    assert.ElementsMatch(t, expected, files)

    // Verify no LICENSE, no CONTRIBUTING, no .editorconfig, no .gitattributes
    assert.NoFileExists(t, filepath.Join(root, "LICENSE"))
    assert.NoFileExists(t, filepath.Join(root, "CONTRIBUTING.md"))
    assert.NoFileExists(t, filepath.Join(root, ".editorconfig"))
    assert.NoFileExists(t, filepath.Join(root, ".gitattributes"))
}
```

#### Acceptance test: Go + Dockerfile (Service), maximum selections

```go
func TestAcceptance_GoDockerfileService_MaxSelections(t *testing.T) {
    defs, err := techdef.Load()
    require.NoError(t, err)

    cfg := &config.Config{
        ProjectName:  "my-project",
        Provider:     "github",
        Technologies: []string{"dockerfile-service", "go"},
        Licence:      "mit",
        Docs:         []string{"contributing"},
        Tooling:      []string{"editorconfig", "gitattributes"},
        RepoConfig:   []string{"issue_templates", "pr_template", "dependabot"},
        Confirmed:    true,
    }

    baseDir := t.TempDir()
    keys := []string{"dockerfile-service", "go"}
    sort.Strings(keys)
    techs := make([]*techdef.TechDef, len(keys))
    for i, key := range keys {
        techs[i] = defs[key]
    }
    err = scaffold.Generate(cfg, techs, baseDir)
    require.NoError(t, err)

    root := filepath.Join(baseDir, "my-project")

    var files []string
    filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        rel, _ := filepath.Rel(root, path)
        if rel != "." {
            files = append(files, rel)
        }
        return nil
    })

    expected := []string{
        ".devcontainer",
        ".devcontainer/Dockerfile",
        ".devcontainer/devcontainer.json",
        ".devcontainer/setup.sh",
        ".dockerignore",
        ".editorconfig",
        ".gitattributes",
        ".github",
        ".github/ISSUE_TEMPLATE",
        ".github/ISSUE_TEMPLATE/bug_report.yaml",
        ".github/ISSUE_TEMPLATE/feature_request.yaml",
        ".github/dependabot.yml",
        ".github/pull_request_template.md",
        ".github/workflows",
        ".github/workflows/ci.yml",
        ".gitignore",
        "CONTRIBUTING.md",
        "cmd",
        "cmd/app",
        "cmd/app/.gitkeep",
        "docker",
        "docker/.gitkeep",
        "docker/Dockerfile",
        "internal",
        "internal/.gitkeep",
        "LICENSE",
        "Makefile",
        "README.md",
    }
    assert.ElementsMatch(t, expected, files)

    // Verify composite gitignore
    gitignore, _ := os.ReadFile(filepath.Join(root, ".gitignore"))
    gitignoreStr := string(gitignore)
    assert.Contains(t, gitignoreStr, "# Dockerfile (Service)")
    assert.Contains(t, gitignoreStr, "# Go")
    dockerIdx := strings.Index(gitignoreStr, "# Dockerfile (Service)")
    goIdx := strings.Index(gitignoreStr, "# Go")
    assert.Less(t, dockerIdx, goIdx, "Docker section should appear before Go (alphabetical)")

    // Verify composite CI
    ci, _ := os.ReadFile(filepath.Join(root, ".github/workflows/ci.yml"))
    ciStr := string(ci)
    assert.Contains(t, ciStr, "lint-docker")
    assert.Contains(t, ciStr, "lint-go")
    assert.Contains(t, ciStr, "test-docker")
    assert.Contains(t, ciStr, "test-go")
    assert.Contains(t, ciStr, "hadolint docker/Dockerfile")
    assert.Contains(t, ciStr, "docker build -t test-image -f docker/Dockerfile .")
    assert.Contains(t, ciStr, "build")

    // Verify composite devcontainer
    dcJson, _ := os.ReadFile(filepath.Join(root, ".devcontainer/devcontainer.json"))
    dcStr := string(dcJson)
    assert.Contains(t, dcStr, "docker-in-docker")
    assert.Contains(t, dcStr, "ghcr.io/devcontainers/features/go:1")
    assert.Contains(t, dcStr, "golang.go")
    assert.Contains(t, dcStr, "ms-azuretools.vscode-docker")

    // Verify composite setup.sh — Docker has empty setup, so only Go block present
    setupSh, _ := os.ReadFile(filepath.Join(root, ".devcontainer/setup.sh"))
    setupStr := string(setupSh)
    assert.Contains(t, setupStr, "# === Base tooling ===")
    assert.Contains(t, setupStr, "# === Go ===")
    assert.NotContains(t, setupStr, "# === Dockerfile")
}
```

#### Acceptance test: Dockerfile (Service) standalone constraint rejection

```go
func TestAcceptance_DockerfileImageStandaloneRejection(t *testing.T) {
    defs, err := techdef.Load()
    require.NoError(t, err)

    cfg := &config.Config{
        ProjectName:  "bad-combo",
        Provider:     "github",
        Technologies: []string{"dockerfile-image", "go"},
        Licence:      "none",
        Confirmed:    true,
    }

    err = cfg.Validate(defs)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "standalone")
}
```

#### Acceptance test: Go + Python + Dockerfile (Service) + Terraform Infrastructure

```go
func TestAcceptance_AllComposable_WithDockerfileService(t *testing.T) {
    defs, err := techdef.Load()
    require.NoError(t, err)

    cfg := &config.Config{
        ProjectName:         "my-project",
        Provider:            "github",
        Technologies:        []string{"dockerfile-service", "go", "python", "terraform-infrastructure"},
        TechPromptResponses: map[string]string{"package_name": "my_project"},
        Licence:             "none",
        Docs:                []string{},
        Tooling:             []string{},
        RepoConfig:          []string{},
        Confirmed:           true,
    }

    baseDir := t.TempDir()
    keys := []string{"dockerfile-service", "go", "python", "terraform-infrastructure"}
    sort.Strings(keys)
    techs := make([]*techdef.TechDef, len(keys))
    for i, key := range keys {
        techs[i] = defs[key]
    }
    err = scaffold.Generate(cfg, techs, baseDir)
    require.NoError(t, err)

    root := filepath.Join(baseDir, "my-project")

    // Verify key directories exist
    assert.DirExists(t, filepath.Join(root, "docker"))
    assert.DirExists(t, filepath.Join(root, "cmd/app"))
    assert.DirExists(t, filepath.Join(root, "internal"))
    assert.DirExists(t, filepath.Join(root, "infrastructure"))
    assert.DirExists(t, filepath.Join(root, "src/my_project"))
    assert.DirExists(t, filepath.Join(root, "tests"))

    // Verify Dockerfile under docker/
    assert.FileExists(t, filepath.Join(root, "docker/Dockerfile"))

    // Verify .dockerignore at root (for build context)
    assert.FileExists(t, filepath.Join(root, ".dockerignore"))

    // Verify no Dockerfile at root (composable, not standalone)
    assert.NoFileExists(t, filepath.Join(root, "Dockerfile"))

    // Verify CI has all four technology jobs
    ci, _ := os.ReadFile(filepath.Join(root, ".github/workflows/ci.yml"))
    ciStr := string(ci)
    assert.Contains(t, ciStr, "lint-docker")
    assert.Contains(t, ciStr, "lint-go")
    assert.Contains(t, ciStr, "lint-python")
    assert.Contains(t, ciStr, "lint-terraform")
    assert.Contains(t, ciStr, "test-docker")
    assert.Contains(t, ciStr, "test-go")
    assert.Contains(t, ciStr, "test-python")
    assert.Contains(t, ciStr, "test-terraform")

    // Verify composite gitignore has all four sections in alphabetical order
    gitignore, _ := os.ReadFile(filepath.Join(root, ".gitignore"))
    gitignoreStr := string(gitignore)
    dockerIdx := strings.Index(gitignoreStr, "# Dockerfile (Service)")
    goIdx := strings.Index(gitignoreStr, "# Go")
    pyIdx := strings.Index(gitignoreStr, "# Python")
    tfIdx := strings.Index(gitignoreStr, "# Terraform (Infrastructure)")
    assert.Less(t, dockerIdx, goIdx)
    assert.Less(t, goIdx, pyIdx)
    assert.Less(t, pyIdx, tfIdx)

    // Verify composite setup.sh — Docker has no block, others present in order
    setupSh, _ := os.ReadFile(filepath.Join(root, ".devcontainer/setup.sh"))
    setupStr := string(setupSh)
    assert.Contains(t, setupStr, "# === Go ===")
    assert.Contains(t, setupStr, "# === Python ===")
    assert.Contains(t, setupStr, "# === Terraform (Infrastructure) ===")
    assert.NotContains(t, setupStr, "# === Dockerfile")
    goSetupIdx := strings.Index(setupStr, "# === Go ===")
    pySetupIdx := strings.Index(setupStr, "# === Python ===")
    tfSetupIdx := strings.Index(setupStr, "# === Terraform (Infrastructure) ===")
    assert.Less(t, goSetupIdx, pySetupIdx)
    assert.Less(t, pySetupIdx, tfSetupIdx)
}
```

### 7.5 Running Tests

Unchanged from Stage 2:

```bash
# Unit tests only
go test ./internal/...

# Acceptance tests only
go test ./tests/acceptance/

# All tests
go test ./...
```

---

## 8. Implementation Order

This section defines the order in which features should be implemented, following TDD. Each step begins with writing the test, then the production code.

### Step 1: Dockerfile (Image) definition

1. Create `technologies/dockerfile-image.yaml` with the content specified in Section 3.1.
2. Write tests verifying the definition loads correctly, passes validation, has expected fields (name, standalone, structure, gitignore, devcontainer, CI).
3. Verify `techdef.Load()` returns 7 technology definitions (5 from Stage 2 + 2 new).

### Step 2: Dockerfile (Service) definition

1. Create `technologies/dockerfile-service.yaml` with the content specified in Section 3.2.
2. Write tests verifying the definition loads correctly, passes validation, has expected fields.
3. Write tests verifying all structure paths are under `docker/`.

### Step 3: Constraint and conflict tests

1. Write tests verifying Dockerfile (Image) is standalone and cannot combine with other techs.
2. Write tests verifying Dockerfile (Service) is composable and can combine with Go, Python, and Terraform Infrastructure.
3. Write path conflict tests verifying Dockerfile (Service) has no conflicts with any existing composable technology.

### Step 4: Acceptance tests

1. Write acceptance test for Dockerfile (Image), maximum selections.
2. Write acceptance test for Dockerfile (Image), minimum selections.
3. Write acceptance test for Go + Dockerfile (Service), maximum selections.
4. Write acceptance test for Dockerfile (Image) standalone constraint rejection.
5. Write acceptance test for all composable technologies together (Go + Python + Dockerfile (Service) + Terraform Infrastructure).
6. Verify all acceptance tests pass.

### Step 5: CI verification

1. Push changes and verify CI passes (lint, unit tests, acceptance tests).
2. Verify all existing Stage 1 and Stage 2 acceptance tests still pass.

---

## 9. Acceptance Criteria

Stage 3a is complete when all of the following are true:

1. Both Dockerfile definitions (`dockerfile-image.yaml`, `dockerfile-service.yaml`) load, parse, and validate correctly.
2. Dockerfile (Image) is standalone — it cannot be selected alongside any other technology. The prompt prevents it and config validation enforces it.
3. Dockerfile (Service) is composable — it can be selected alongside other composable technologies (Go, Python, Terraform Infrastructure).
4. Dockerfile (Image) scaffolds files at the project root: `Dockerfile`, `.dockerignore`, `scripts/`.
5. Dockerfile (Service) scaffolds `docker/Dockerfile` under the `docker/` subdirectory and `.dockerignore` at the project root.
6. No path conflicts exist between Dockerfile (Service) and any existing composable technology (Go, Python, Terraform Infrastructure).
7. The `.gitignore` correctly includes the Docker section for both definitions.
8. The devcontainer correctly includes the `docker-in-docker:2` feature and `ms-azuretools.vscode-docker` extension for both definitions.
9. The CI config correctly generates `lint-docker` and `test-docker` jobs with Hadolint and `docker build` steps respectively.
10. For Dockerfile (Service), CI lint and test steps reference the correct path (`docker/Dockerfile`).
11. For Dockerfile (Image), CI lint and test steps reference the root path (`Dockerfile`).
12. Dockerfile (Service) has an empty `setup` field and does not contribute a block to `setup.sh`.
13. The `{{.ProjectName}}` template variable in the Dockerfile `LABEL` is correctly substituted.
14. All new unit tests pass.
15. All new acceptance tests pass.
16. All existing Stage 1 and Stage 2 tests continue to pass without modification.
17. The project's own CI pipeline passes lint, unit tests, and acceptance tests.
18. No schema, engine, config, or prompt code changes were required — only new YAML definition files and tests.