# SPEC — doozer-scaffold Stage 1: Foundations

**Version:** 2.0
**Status:** Final
**Parent PRD:** doozer-scaffold PRD v0.2-draft
**Owner:** Tony Lea

---

## 1. Purpose

This document is the complete technical specification for Stage 1 of `doozer-scaffold`. It defines every input, output, behaviour, and test expectation in sufficient detail that an implementer can build the stage without ambiguity.

Stage 1 delivers a working CLI that scaffolds a single-technology (PowerShell module) project targeting GitHub, proving the full toolchain, project structure, TDD approach, and CI pipeline. Technologies are defined declaratively via YAML definition files (see ADR-001).

---

## 2. Technology Stack

### 2.1 Language and Module Path

- Language: Go (latest stable release at time of development).
- Module path: `github.com/tonylea/doozer-scaffold`.
- Entry point: `cmd/doozer-scaffold/main.go`.

### 2.2 Project Layout (doozer-scaffold itself)

```
doozer-scaffold/
├── cmd/
│   └── doozer-scaffold/
│       └── main.go
├── internal/
│   ├── prompt/
│   │   └── prompt.go          # Interactive prompt logic
│   ├── scaffold/
│   │   └── scaffold.go        # File generation orchestration
│   ├── techdef/
│   │   └── techdef.go         # Technology definition loading and parsing
│   ├── templates/
│   │   └── templates.go       # Template loading and rendering
│   └── config/
│       └── config.go          # User selections data structure
├── technologies/               # Technology definition files (via go:embed)
│   └── powershell.yaml
├── templates/                  # Embedded template files (via go:embed)
│   ├── licences/
│   │   └── MIT.tmpl
│   ├── github/
│   │   ├── ci.yml.tmpl
│   │   ├── dependabot.yml.tmpl
│   │   ├── bug_report.yaml.tmpl
│   │   ├── feature_request.yaml.tmpl
│   │   └── pull_request_template.md.tmpl
│   ├── devcontainer/
│   │   ├── devcontainer.json.tmpl
│   │   ├── Dockerfile.tmpl
│   │   └── setup/
│   │       └── base.sh.tmpl
│   ├── docs/
│   │   └── CONTRIBUTING.md.tmpl
│   ├── editorconfig.tmpl
│   ├── gitattributes.tmpl
│   └── README.md.tmpl
├── tests/
│   └── acceptance/
│       └── acceptance_test.go
├── go.mod
├── go.sum
├── Makefile
├── .github/
│   └── workflows/
│       └── ci.yml              # doozer-scaffold's own CI
├── .gitignore
├── .golangci.yml
└── README.md
```

### 2.3 Key Dependencies

| Dependency                     | Purpose                                                                         |
| ------------------------------ | ------------------------------------------------------------------------------- |
| `github.com/charmbracelet/huh` | Interactive terminal prompts (single-select, multi-select, text input, confirm) |
| `github.com/stretchr/testify`  | Test assertions (`assert`, `require`) and test suites                           |
| `gopkg.in/yaml.v3`             | Parsing technology definition YAML files                                        |

See ADR-002, ADR-003, and ADR-004 for selection rationale.

### 2.4 Template Engine

Use Go's standard `text/template` package for all template rendering. Templates and technology definitions are embedded into the binary using `go:embed`.

```go
import "embed"

//go:embed templates/*
var templateFS embed.FS

//go:embed technologies/*
var techFS embed.FS
```

Template variables use the following naming convention:

```go
type TemplateData struct {
    ProjectName string
    Year        string // e.g. "2025", used in licence files
}
```

---

## 3. Technology Definition Format

### 3.1 Overview

Each technology is defined by a single YAML file in the `technologies/` directory. The file name (without extension) is the technology's key (e.g. `powershell.yaml` → key `powershell`). The scaffold engine reads these definitions at startup, presents them as options in the prompt, and uses their contents to drive file generation. See ADR-001 for the rationale behind this approach.

### 3.2 Schema

```yaml
# Required. Display name shown in the interactive prompt.
name: "PowerShell Module"

# Required. Files and directories to create.
# Each entry is a file path relative to the project root.
# Paths ending in / are directories (a .gitkeep file is created inside).
# Paths not ending in / are files. Their content is specified by the "content" field.
# If no "content" field is present, the file is created empty (zero bytes).
structure:
  - path: "src/classes/"
  - path: "src/private/"
  - path: "src/public/"
  - path: "tests/unit-tests/private/"
  - path: "tests/unit-tests/public/"
  - path: "tests/integration-tests/"

# Required. Lines to include in the composite .gitignore.
# The engine adds a "# {name}" header comment before these lines.
gitignore: |
  *.ps1xml
  *.nupkg
  *.snk
  output/
  logs/
  TestResults/
  code-coverage.*
  node_modules/
  package-lock.json

# Required. Devcontainer configuration contributed by this technology.
devcontainer:
  # Devcontainer features to merge into the features object.
  features:
    "ghcr.io/devcontainers/features/powershell:1": {}
  # VS Code extensions to merge into the extensions array.
  extensions:
    - "ms-vscode.powershell"
  # Shell commands to append to setup.sh after the base block.
  # The engine adds a "# === {name} ===" header comment before these lines.
  setup: |
    pwsh -NoProfile -Command "Install-Module -Name Pester -Force -Scope AllUsers"
    pwsh -NoProfile -Command "Install-Module -Name PSScriptAnalyzer -Force -Scope AllUsers"
    pwsh -NoProfile -Command "Install-Module -Name PlatyPS -Force -Scope AllUsers"
    pwsh -NoProfile -Command "Install-Module -Name BuildHelpers -Force -Scope AllUsers"
```

### 3.3 Schema Rules

**`structure` entries:**

- A path ending with `/` denotes a directory. The engine creates the directory and places an empty `.gitkeep` file inside it (zero bytes).
- A path not ending with `/` denotes a file. If a `content` field is present, the file is created with that content. If no `content` field is present, the file is created empty (zero bytes).
- Paths are always relative to the project root. Leading `/` is not permitted.
- Nested directories are created automatically (equivalent to `mkdir -p`).

Example showing both directories and files with content:

```yaml
structure:
  - path: "src/classes/"                    # Directory → creates src/classes/.gitkeep
  - path: "src/private/"                    # Directory → creates src/private/.gitkeep
  - path: "src/MyModule.psm1"              # File with content
    content: |
      # Root module file
      Get-ChildItem "$PSScriptRoot/public/*.ps1" | ForEach-Object { . $_.FullName }
  - path: "src/MyModule.psd1"              # File with content
    content: |
      @{
          RootModule = 'MyModule.psm1'
          ModuleVersion = '0.1.0'
      }
  - path: "tests/unit-tests/private/"      # Directory → creates tests/unit-tests/private/.gitkeep
  - path: "config.json"                    # File without content → created empty
```

See ADR-018 for the rationale on supporting all three entry types from Stage 1.

**`gitignore`:**

- A multi-line string. Each line becomes a line in the composite `.gitignore`.
- The engine prepends a `# {name}` comment header automatically. The definition should not include its own header.

**`devcontainer.features`:**

- A map of feature URIs to configuration objects. In most cases the configuration object is empty (`{}`).
- These are merged into the `devcontainer.json` features object alongside the base features (Node.js is always included as a base feature).

**`devcontainer.extensions`:**

- A list of VS Code extension identifiers.
- These are merged into the `devcontainer.json` extensions array.

**`devcontainer.setup`:**

- A multi-line string of shell commands.
- The engine prepends a `# === {name} ===` comment header automatically. The definition should not include its own header.
- These commands are appended to `setup.sh` after the base block.

### 3.4 Go Data Structure

```go
package techdef

import (
    "embed"
    "fmt"
    "path/filepath"
    "strings"

    "gopkg.in/yaml.v3"
)

//go:embed technologies/*
var techFS embed.FS

// TechDef represents a parsed technology definition.
type TechDef struct {
    Name         string            `yaml:"name"`
    Structure    []StructureEntry  `yaml:"structure"`
    Gitignore    string            `yaml:"gitignore"`
    Devcontainer DevcontainerDef   `yaml:"devcontainer"`
}

// StructureEntry represents a single file or directory in the technology's scaffold.
type StructureEntry struct {
    Path    string  `yaml:"path"`
    Content *string `yaml:"content,omitempty"` // nil = directory (if path ends with /) or empty file (if not)
}

// IsDir returns true if this entry represents a directory (path ends with /).
func (s StructureEntry) IsDir() bool {
    return strings.HasSuffix(s.Path, "/")
}

type DevcontainerDef struct {
    Features   map[string]interface{} `yaml:"features"`
    Extensions []string               `yaml:"extensions"`
    Setup      string                 `yaml:"setup"`
}

// Load reads all technology definitions from the embedded filesystem.
// Returns a map keyed by technology key (filename without extension).
func Load() (map[string]*TechDef, error) {
    entries, err := techFS.ReadDir("technologies")
    if err != nil {
        return nil, fmt.Errorf("reading technologies directory: %w", err)
    }

    defs := make(map[string]*TechDef)
    for _, entry := range entries {
        if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
            continue
        }

        data, err := techFS.ReadFile(filepath.Join("technologies", entry.Name()))
        if err != nil {
            return nil, fmt.Errorf("reading %s: %w", entry.Name(), err)
        }

        var def TechDef
        if err := yaml.Unmarshal(data, &def); err != nil {
            return nil, fmt.Errorf("parsing %s: %w", entry.Name(), err)
        }

        key := strings.TrimSuffix(entry.Name(), ".yaml")
        defs[key] = &def
    }

    return defs, nil
}
```

### 3.5 PowerShell Module Definition

File: `technologies/powershell.yaml`

```yaml
name: "PowerShell Module"

structure:
  - path: "src/classes/"
  - path: "src/private/"
  - path: "src/public/"
  - path: "tests/unit-tests/private/"
  - path: "tests/unit-tests/public/"
  - path: "tests/integration-tests/"

gitignore: |
  *.ps1xml
  *.nupkg
  *.snk
  output/
  logs/
  TestResults/
  code-coverage.*
  node_modules/
  package-lock.json

devcontainer:
  features:
    "ghcr.io/devcontainers/features/powershell:1": {}
  extensions:
    - "ms-vscode.powershell"
  setup: |
    pwsh -NoProfile -Command "Install-Module -Name Pester -Force -Scope AllUsers"
    pwsh -NoProfile -Command "Install-Module -Name PSScriptAnalyzer -Force -Scope AllUsers"
    pwsh -NoProfile -Command "Install-Module -Name PlatyPS -Force -Scope AllUsers"
    pwsh -NoProfile -Command "Install-Module -Name BuildHelpers -Force -Scope AllUsers"
```

### 3.6 Validation

The `techdef` package must validate each definition on load. Validation rules:

1. `name` must be non-empty.
2. `structure` must contain at least one entry.
3. Every `structure` entry must have a non-empty `path`.
4. No `path` may start with `/` or contain `..`.
5. `gitignore` must be non-empty.
6. `devcontainer` must be present. `features`, `extensions`, and `setup` may each be empty but the `devcontainer` key itself must exist.

If validation fails, the tool must exit with a clear error message identifying the invalid definition file and the reason.

```go
func (t *TechDef) Validate(key string) error {
    if t.Name == "" {
        return fmt.Errorf("technology '%s': name is required", key)
    }
    if len(t.Structure) == 0 {
        return fmt.Errorf("technology '%s': structure must contain at least one entry", key)
    }
    for i, entry := range t.Structure {
        if entry.Path == "" {
            return fmt.Errorf("technology '%s': structure[%d] has empty path", key, i)
        }
        if strings.HasPrefix(entry.Path, "/") {
            return fmt.Errorf("technology '%s': structure[%d] path must not start with /", key, i)
        }
        if strings.Contains(entry.Path, "..") {
            return fmt.Errorf("technology '%s': structure[%d] path must not contain '..'", key, i)
        }
    }
    if strings.TrimSpace(t.Gitignore) == "" {
        return fmt.Errorf("technology '%s': gitignore is required", key)
    }
    return nil
}
```

---

## 4. User Interaction Flow

The CLI is invoked with no required arguments:

```bash
doozer-scaffold
```

Or with an optional project name:

```bash
doozer-scaffold my-project
```

If no project name is provided as an argument, the tool prompts for one. If a name is given as an argument, the prompt is skipped.

### 4.1 Prompt Sequence

The prompts are presented in this exact order:

| Step | Type                       | Prompt Text                                | Options                                                                |
| ---- | -------------------------- | ------------------------------------------ | ---------------------------------------------------------------------- |
| 1    | Text input (if no CLI arg) | "Project name:"                            | Free text, required, non-empty                                         |
| 2    | Single-select              | "Remote hosting provider:"                 | `GitHub`                                                               |
| 3    | Single-select              | "Technology:"                              | Dynamically populated from loaded technology definitions               |
| 4    | Single-select              | "Licence:"                                 | `MIT`, `None`                                                          |
| 5    | Multi-select               | "Supplementary documentation:"             | `CONTRIBUTING.md`                                                      |
| 6    | Multi-select               | "Project tooling:"                         | `.editorconfig`, `.gitattributes`                                      |
| 7    | Multi-select               | "GitHub repository configuration:"         | `Issue templates`, `Pull request template`, `Dependabot configuration` |
| 8    | Confirm                    | "Generate scaffold with these selections?" | Yes / No                                                               |

**Implementation notes:**

- Step 2 has only one option in Stage 1 (GitHub). It is still presented as a selection prompt, not hard-coded. See ADR-009.
- Step 3 is dynamically populated from the loaded technology definitions. In Stage 1 only `powershell.yaml` exists, so only "PowerShell Module" appears.
- Step 5 and Step 6 are multi-select prompts. The user may select none, some, or all options.
- Step 7 is a multi-select prompt. The user may select none, some, or all options.
- Step 8 shows a summary of all selections before confirmation. If the user declines, the tool exits without writing any files.

### 4.2 Prompt Implementation with huh

Each step maps to a `huh.Form` group. The technology options are built dynamically from the loaded definitions.

```go
package prompt

import (
    "fmt"
    "sort"

    "github.com/charmbracelet/huh"
    "github.com/tonylea/doozer-scaffold/internal/config"
    "github.com/tonylea/doozer-scaffold/internal/techdef"
)

func Run(cfg *config.Config, techDefs map[string]*techdef.TechDef) error {
    groups := []*huh.Group{}

    if cfg.ProjectName == "" {
        groups = append(groups, huh.NewGroup(
            huh.NewInput().
                Title("Project name:").
                Value(&cfg.ProjectName).
                Validate(func(s string) error {
                    if s == "" {
                        return fmt.Errorf("project name is required")
                    }
                    return nil
                }),
        ))
    }

    // Build technology options dynamically from loaded definitions
    techOptions := buildTechOptions(techDefs)

    groups = append(groups,
        huh.NewGroup(
            huh.NewSelect[string]().
                Title("Remote hosting provider:").
                Options(
                    huh.NewOption("GitHub", "github"),
                ).
                Value(&cfg.Provider),
        ),
        huh.NewGroup(
            huh.NewSelect[string]().
                Title("Technology:").
                Options(techOptions...).
                Value(&cfg.Technology),
        ),
        huh.NewGroup(
            huh.NewSelect[string]().
                Title("Licence:").
                Options(
                    huh.NewOption("MIT", "mit"),
                    huh.NewOption("None", "none"),
                ).
                Value(&cfg.Licence),
        ),
        huh.NewGroup(
            huh.NewMultiSelect[string]().
                Title("Supplementary documentation:").
                Options(
                    huh.NewOption("CONTRIBUTING.md", "contributing"),
                ).
                Value(&cfg.Docs),
        ),
        huh.NewGroup(
            huh.NewMultiSelect[string]().
                Title("Project tooling:").
                Options(
                    huh.NewOption(".editorconfig", "editorconfig"),
                    huh.NewOption(".gitattributes", "gitattributes"),
                ).
                Value(&cfg.Tooling),
        ),
        huh.NewGroup(
            huh.NewMultiSelect[string]().
                Title("GitHub repository configuration:").
                Options(
                    huh.NewOption("Issue templates", "issue_templates"),
                    huh.NewOption("Pull request template", "pr_template"),
                    huh.NewOption("Dependabot configuration", "dependabot"),
                ).
                Value(&cfg.RepoConfig),
        ),
        huh.NewGroup(
            huh.NewConfirm().
                Title("Generate scaffold with these selections?").
                Affirmative("Yes").
                Negative("No").
                Value(&cfg.Confirmed),
        ),
    )

    form := huh.NewForm(groups...)
    return form.Run()
}

// buildTechOptions creates huh.Option entries from loaded technology definitions.
// Options are sorted alphabetically by display name for consistent presentation.
func buildTechOptions(techDefs map[string]*techdef.TechDef) []huh.Option[string] {
    type techEntry struct {
        key  string
        name string
    }

    entries := make([]techEntry, 0, len(techDefs))
    for key, def := range techDefs {
        entries = append(entries, techEntry{key: key, name: def.Name})
    }
    sort.Slice(entries, func(i, j int) bool {
        return entries[i].name < entries[j].name
    })

    options := make([]huh.Option[string], 0, len(entries))
    for _, e := range entries {
        options = append(options, huh.NewOption(e.name, e.key))
    }
    return options
}
```

### 4.3 Config Data Structure

```go
package config

type Config struct {
    ProjectName string
    Provider    string
    Technology  string
    Licence     string
    Docs        []string
    Tooling     []string
    RepoConfig  []string
    Confirmed   bool
}
```

### 4.4 CLI Argument Handling

Use Go's standard `flag` package or bare `os.Args`. See ADR-013.

```go
func main() {
    cfg := &config.Config{}

    // If a positional argument is provided, use it as the project name
    if len(os.Args) > 1 {
        cfg.ProjectName = os.Args[1]
    }

    // Load technology definitions
    techDefs, err := techdef.Load()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error loading technology definitions: %v\n", err)
        os.Exit(1)
    }

    // Run interactive prompts
    if err := prompt.Run(cfg, techDefs); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    if !cfg.Confirmed {
        fmt.Println("Scaffold generation cancelled.")
        os.Exit(0)
    }

    // Look up the selected technology definition
    tech, ok := techDefs[cfg.Technology]
    if !ok {
        fmt.Fprintf(os.Stderr, "Error: unknown technology '%s'\n", cfg.Technology)
        os.Exit(1)
    }

    if err := scaffold.Generate(cfg, tech, "."); err != nil {
        fmt.Fprintf(os.Stderr, "Error generating scaffold: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("Project '%s' scaffolded successfully.\n", cfg.ProjectName)
}
```

---

## 5. Output Specification

All output is written to a subdirectory named after the project within the current working directory. If the user runs `doozer-scaffold my-module` or enters `my-module` at the prompt, the output root is `./my-module/`.

### 5.1 Pre-Generation Validation

Before writing any files, the tool must check:

1. The target directory does not already exist. If it does, print an error and exit:
   ```
   Error: directory 'my-module' already exists.
   ```
2. The tool must not create any files or directories until all validations pass and the user has confirmed.

### 5.2 Atomic Generation

File generation should be all-or-nothing. If any file write fails mid-generation, the tool must clean up by removing the partially created directory and print an error:

```
Error: failed to create file 'src/public/.gitkeep': permission denied. Cleaning up.
```

Implementation approach: create all files in the target directory, and if any error occurs, remove the entire directory before exiting.

### 5.3 Universal Outputs (always generated)

These files are always produced regardless of user selections:

#### 5.3.1 README.md

```markdown
# {{.ProjectName}}
```

Template: `templates/README.md.tmpl`

The README contains only the project name as an H1 heading. Nothing else.

#### 5.3.2 .gitignore

For the PowerShell technology:

```gitignore
# PowerShell Module
*.ps1xml
*.nupkg
*.snk
output/
logs/
TestResults/
code-coverage.*
node_modules/
package-lock.json
```

#### 5.3.3 CI Configuration

Generated at: `.github/workflows/ci.yml`

This is a valid but minimal GitHub Actions workflow file with placeholder steps. It triggers on push and pull request, with a single job containing commented-out step placeholders for the user to complete.

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  ci:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      # TODO: Add linting steps
      # TODO: Add test steps
      # TODO: Add build steps
```

Template: `templates/github/ci.yml.tmpl`

### 5.4 Conditional Outputs — Licence

#### MIT (when `cfg.Licence == "mit"`)

Generated at: `LICENSE`

```
MIT License

Copyright (c) {{.Year}} {{.ProjectName}}

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

Template: `templates/licences/MIT.tmpl`

When `cfg.Licence == "none"`, no `LICENSE` file is generated.

### 5.5 Conditional Outputs — Technology Structure

The scaffold engine reads the selected technology's `structure` field and creates all specified directories and files. This is entirely data-driven — the engine does not contain any technology-specific logic.

**Engine behaviour for each structure entry:**

| Entry type           | Detection                                             | Action                                                                         |
| -------------------- | ----------------------------------------------------- | ------------------------------------------------------------------------------ |
| Directory            | `path` ends with `/`                                  | Create the directory and place an empty `.gitkeep` file (zero bytes) inside it |
| File without content | `path` does not end with `/`, no `content` field      | Create the file with zero bytes                                                |
| File with content    | `path` does not end with `/`, `content` field present | Create the file with the specified content                                     |

**Implementation:**

```go
func createStructure(targetDir string, structure []techdef.StructureEntry) error {
    for _, entry := range structure {
        fullPath := filepath.Join(targetDir, entry.Path)

        if entry.IsDir() {
            // Create directory with .gitkeep
            if err := os.MkdirAll(fullPath, 0o755); err != nil {
                return fmt.Errorf("creating directory '%s': %w", entry.Path, err)
            }
            gitkeepPath := filepath.Join(fullPath, ".gitkeep")
            if err := os.WriteFile(gitkeepPath, []byte{}, 0o644); err != nil {
                return fmt.Errorf("creating .gitkeep in '%s': %w", entry.Path, err)
            }
        } else {
            // Create parent directories
            parentDir := filepath.Dir(fullPath)
            if err := os.MkdirAll(parentDir, 0o755); err != nil {
                return fmt.Errorf("creating parent directory for '%s': %w", entry.Path, err)
            }
            // Write file content (empty if no content specified)
            content := []byte{}
            if entry.Content != nil {
                content = []byte(*entry.Content)
            }
            if err := os.WriteFile(fullPath, content, 0o644); err != nil {
                return fmt.Errorf("creating file '%s': %w", entry.Path, err)
            }
        }
    }
    return nil
}
```

For the PowerShell definition in Stage 1, this produces:

```
{project}/
├── src/
│   ├── classes/
│   │   └── .gitkeep
│   ├── private/
│   │   └── .gitkeep
│   └── public/
│       └── .gitkeep
└── tests/
    ├── unit-tests/
    │   ├── private/
    │   │   └── .gitkeep
    │   └── public/
    │       └── .gitkeep
    └── integration-tests/
        └── .gitkeep
```

All `.gitkeep` files are empty (zero bytes).

### 5.6 Conditional Outputs — Supplementary Documentation

#### CONTRIBUTING.md (when `"contributing"` is in `cfg.Docs`)

Generated at: `CONTRIBUTING.md`

```markdown
# Contributing to {{.ProjectName}}

Thank you for considering contributing to {{.ProjectName}}.

## How to Contribute

1. Fork the repository.
2. Create a feature branch from `main`.
3. Make your changes.
4. Submit a pull request.

## Reporting Issues

Please use the GitHub issue tracker to report bugs or request features.
```

Template: `templates/docs/CONTRIBUTING.md.tmpl`

### 5.7 Conditional Outputs — Project Tooling

#### .editorconfig (when `"editorconfig"` is in `cfg.Tooling`)

Generated at: `.editorconfig`

```ini
root = true

[*]
indent_style = space
indent_size = 4
end_of_line = lf
charset = utf-8
trim_trailing_whitespace = true
insert_final_newline = true

[*.md]
trim_trailing_whitespace = false

[*.{yml,yaml}]
indent_size = 2

[Makefile]
indent_style = tab
```

Template: `templates/editorconfig.tmpl`

#### .gitattributes (when `"gitattributes"` is in `cfg.Tooling`)

Generated at: `.gitattributes`

```
* text=auto eol=lf
*.ps1 text eol=lf
*.psm1 text eol=lf
*.psd1 text eol=lf
*.md text eol=lf
*.yml text eol=lf
*.yaml text eol=lf
*.json text eol=lf
```

Template: `templates/gitattributes.tmpl`

### 5.8 Conditional Outputs — GitHub Repository Configuration

#### Issue Templates (when `"issue_templates"` is in `cfg.RepoConfig`)

Two files are generated:

**`.github/ISSUE_TEMPLATE/bug_report.yaml`**

```yaml
name: Bug Report
description: Report a bug
labels: ["bug"]
body:
  - type: textarea
    id: description
    attributes:
      label: Description
      description: A clear description of the bug.
    validations:
      required: true
  - type: textarea
    id: reproduction
    attributes:
      label: Steps to Reproduce
      description: Steps to reproduce the behaviour.
    validations:
      required: true
  - type: textarea
    id: expected
    attributes:
      label: Expected Behaviour
      description: What you expected to happen.
    validations:
      required: true
  - type: textarea
    id: environment
    attributes:
      label: Environment
      description: "OS, PowerShell version, etc."
    validations:
      required: false
```

Template: `templates/github/bug_report.yaml.tmpl`

**`.github/ISSUE_TEMPLATE/feature_request.yaml`**

```yaml
name: Feature Request
description: Suggest a new feature
labels: ["enhancement"]
body:
  - type: textarea
    id: description
    attributes:
      label: Description
      description: A clear description of the feature you'd like.
    validations:
      required: true
  - type: textarea
    id: motivation
    attributes:
      label: Motivation
      description: Why is this feature needed? What problem does it solve?
    validations:
      required: false
  - type: textarea
    id: alternatives
    attributes:
      label: Alternatives Considered
      description: Any alternative solutions or features you've considered.
    validations:
      required: false
```

Template: `templates/github/feature_request.yaml.tmpl`

#### Pull Request Template (when `"pr_template"` is in `cfg.RepoConfig`)

Generated at: `.github/pull_request_template.md`

```markdown
## Description

<!-- Describe your changes -->

## Type of Change

- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Checklist

- [ ] Tests pass
- [ ] Documentation updated (if applicable)
```

Template: `templates/github/pull_request_template.md.tmpl`

#### Dependabot Configuration (when `"dependabot"` is in `cfg.RepoConfig`)

Generated at: `.github/dependabot.yml`

```yaml
version: 2
updates:
  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "weekly"
```

Template: `templates/github/dependabot.yml.tmpl`

### 5.9 Conditional Outputs — Devcontainer

The devcontainer is always generated when a technology is selected (which is always the case — the technology prompt is required). Its contents are composed from a base layer plus contributions from the selected technology's definition.

```
.devcontainer/
├── Dockerfile
├── devcontainer.json
└── setup.sh
```

Only the ARM architecture is supported in Stage 1. See ADR-008.

#### Composability Model

The scaffold engine composes the devcontainer from two sources:

1. **Base layer** (always present): the Dockerfile, base devcontainer features (Node.js), and the base setup.sh block (markdownlint, commitlint).
2. **Technology layer**: features, extensions, and setup commands from the selected technology's `devcontainer` field.

See ADR-006 and ADR-007 for rationale.

#### Dockerfile

Generated at: `.devcontainer/Dockerfile`

```dockerfile
FROM mcr.microsoft.com/devcontainers/base:ubuntu

# Base packages
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    curl \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*
```

Template: `templates/devcontainer/Dockerfile.tmpl`

#### devcontainer.json

Generated at: `.devcontainer/devcontainer.json`

The features and extensions are composed from the base layer plus the selected technology definition. For the PowerShell technology:

```json
{
    "name": "{{.ProjectName}}",
    "build": {
        "dockerfile": "Dockerfile"
    },
    "features": {
        "ghcr.io/devcontainers/features/node:1": {},
        "ghcr.io/devcontainers/features/powershell:1": {}
    },
    "customizations": {
        "vscode": {
            "extensions": [
                "ms-vscode.powershell"
            ]
        }
    },
    "postCreateCommand": "bash .devcontainer/setup.sh"
}
```

Template: `templates/devcontainer/devcontainer.json.tmpl`

The template receives composed data — the engine merges the base features with the technology's features before rendering.

**Base features (always included):**

| Feature                                 | Source        |
| --------------------------------------- | ------------- |
| `ghcr.io/devcontainers/features/node:1` | Base (always) |

**Technology-contributed features and extensions:**

| Feature / Extension                           | Source                |
| --------------------------------------------- | --------------------- |
| `ghcr.io/devcontainers/features/powershell:1` | PowerShell definition |
| `ms-vscode.powershell` extension              | PowerShell definition |

**Implementation for devcontainer.json rendering:**

Render programmatically via `encoding/json` with `json.MarshalIndent` rather than via `text/template`. See ADR-005 for rationale.

```go
type DevcontainerJSON struct {
    Name        string                 `json:"name"`
    Build       map[string]string      `json:"build"`
    Features    map[string]interface{} `json:"features"`
    Customizations map[string]interface{} `json:"customizations"`
    PostCreateCommand string            `json:"postCreateCommand"`
}

func renderDevcontainerJSON(projectName string, tech *techdef.TechDef) ([]byte, error) {
    // Base features
    features := map[string]interface{}{
        "ghcr.io/devcontainers/features/node:1": map[string]interface{}{},
    }
    // Merge technology features
    for k, v := range tech.Devcontainer.Features {
        features[k] = v
    }

    dc := DevcontainerJSON{
        Name:  projectName,
        Build: map[string]string{"dockerfile": "Dockerfile"},
        Features: features,
        Customizations: map[string]interface{}{
            "vscode": map[string]interface{}{
                "extensions": tech.Devcontainer.Extensions,
            },
        },
        PostCreateCommand: "bash .devcontainer/setup.sh",
    }

    return json.MarshalIndent(dc, "", "    ")
}
```

#### setup.sh

Generated at: `.devcontainer/setup.sh`

The script is composed from a base block followed by the selected technology's setup commands.

For the PowerShell technology:

```bash
#!/bin/bash
set -e

# === Base tooling ===
npm install -g markdownlint-cli2 @commitlint/cli @commitlint/config-conventional

# === PowerShell Module ===
pwsh -NoProfile -Command "Install-Module -Name Pester -Force -Scope AllUsers"
pwsh -NoProfile -Command "Install-Module -Name PSScriptAnalyzer -Force -Scope AllUsers"
pwsh -NoProfile -Command "Install-Module -Name PlatyPS -Force -Scope AllUsers"
pwsh -NoProfile -Command "Install-Module -Name BuildHelpers -Force -Scope AllUsers"
```

The base block is rendered from `templates/devcontainer/setup/base.sh.tmpl`:

```bash
#!/bin/bash
set -e

# === Base tooling ===
npm install -g markdownlint-cli2 @commitlint/cli @commitlint/config-conventional
```

The technology block is taken directly from the technology definition's `devcontainer.setup` field, prefixed with a `# === {name} ===` header by the engine:

```go
func renderSetupSh(baseTmpl string, tech *techdef.TechDef) string {
    var sb strings.Builder
    sb.WriteString(baseTmpl)
    if strings.TrimSpace(tech.Devcontainer.Setup) != "" {
        sb.WriteString("\n# === ")
        sb.WriteString(tech.Name)
        sb.WriteString(" ===\n")
        sb.WriteString(tech.Devcontainer.Setup)
    }
    return sb.String()
}
```

The `setup.sh` file must be generated with executable permissions (`0o755`).

### 5.10 Complete Output Tree — Maximum Selection

When all options are selected (MIT licence, CONTRIBUTING.md, both tooling files, all GitHub repo config), with the PowerShell technology:

```
my-module/
├── .devcontainer/
│   ├── devcontainer.json
│   ├── Dockerfile
│   └── setup.sh
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
├── LICENSE
├── README.md
├── src/
│   ├── classes/
│   │   └── .gitkeep
│   ├── private/
│   │   └── .gitkeep
│   └── public/
│       └── .gitkeep
└── tests/
    ├── integration-tests/
    │   └── .gitkeep
    └── unit-tests/
        ├── private/
        │   └── .gitkeep
        └── public/
            └── .gitkeep
```

### 5.11 Complete Output Tree — Minimum Selection

When minimum options are selected (no licence, no docs, no tooling, no repo config), with the PowerShell technology:

```
my-module/
├── .devcontainer/
│   ├── devcontainer.json
│   ├── Dockerfile
│   └── setup.sh
├── .github/
│   └── workflows/
│       └── ci.yml
├── .gitignore
├── README.md
├── src/
│   ├── classes/
│   │   └── .gitkeep
│   ├── private/
│   │   └── .gitkeep
│   └── public/
│       └── .gitkeep
└── tests/
    ├── integration-tests/
    │   └── .gitkeep
    └── unit-tests/
        ├── private/
        │   └── .gitkeep
        └── public/
            └── .gitkeep
```

---

## 6. Scaffold Engine — Generate Function

The `Generate` function accepts the config, the resolved technology definition, and a base directory. See ADR-010.

```go
// Generate creates the scaffold in a subdirectory of baseDir named after cfg.ProjectName.
func Generate(cfg *config.Config, tech *techdef.TechDef, baseDir string) error {
    targetDir := filepath.Join(baseDir, cfg.ProjectName)

    // Check target doesn't already exist
    if _, err := os.Stat(targetDir); err == nil {
        return fmt.Errorf("directory '%s' already exists", cfg.ProjectName)
    }

    // Create directory structure and files
    // On any error, clean up targetDir and return the error
    if err := createScaffold(cfg, tech, targetDir); err != nil {
        os.RemoveAll(targetDir)
        return fmt.Errorf("scaffold generation failed: %w", err)
    }

    return nil
}
```

When called from `main.go`, `baseDir` is `"."` (the current working directory):

```go
if err := scaffold.Generate(cfg, tech, "."); err != nil {
    // ...
}
```

The `createScaffold` function orchestrates file generation in this order:

1. Create the target directory.
2. Generate universal outputs (README.md, .gitignore, CI config).
3. Generate technology structure (from the definition's `structure` field).
4. Generate devcontainer (from base templates + technology definition).
5. Generate conditional outputs (licence, docs, tooling, repo config) based on user selections.

Each step uses the technology definition data — the engine contains no technology-specific logic.

---

## 7. Testing Strategy

### 7.1 Principles

All development follows strict TDD. The workflow is:

1. Write a failing test.
2. Write the minimum production code to make it pass.
3. Refactor.
4. Repeat.

No production code is written without a failing test driving it.

### 7.2 Test Framework

Use `github.com/stretchr/testify` for assertions. Use the standard `testing` package for test structure.

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)
```

Use `assert` for checks where the test can continue after a failure. Use `require` for checks where a failure should stop the test immediately (e.g. "the directory must exist before we check its contents").

### 7.3 Unit Tests

Unit tests live alongside the code they test, following standard Go conventions:

```
internal/
├── scaffold/
│   ├── scaffold.go
│   └── scaffold_test.go
├── techdef/
│   ├── techdef.go
│   └── techdef_test.go
├── templates/
│   ├── templates.go
│   └── templates_test.go
└── config/
    ├── config.go
    └── config_test.go
```

#### Unit test examples

**Testing technology definition loading:**

```go
func TestLoadTechDefs(t *testing.T) {
    defs, err := techdef.Load()
    require.NoError(t, err)
    require.Contains(t, defs, "powershell")

    ps := defs["powershell"]
    assert.Equal(t, "PowerShell Module", ps.Name)
    assert.NotEmpty(t, ps.Structure)
    assert.NotEmpty(t, ps.Gitignore)
    assert.NotEmpty(t, ps.Devcontainer.Features)
    assert.NotEmpty(t, ps.Devcontainer.Extensions)
    assert.NotEmpty(t, ps.Devcontainer.Setup)
}
```

**Testing technology definition validation:**

```go
func TestTechDefValidation_MissingName(t *testing.T) {
    def := &techdef.TechDef{
        Name:      "",
        Structure: []techdef.StructureEntry{{Path: "src/"}},
        Gitignore: "*.log",
    }
    err := def.Validate("test")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "name is required")
}

func TestTechDefValidation_AbsolutePath(t *testing.T) {
    def := &techdef.TechDef{
        Name:      "Test",
        Structure: []techdef.StructureEntry{{Path: "/etc/passwd"}},
        Gitignore: "*.log",
    }
    err := def.Validate("test")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "must not start with /")
}

func TestTechDefValidation_PathTraversal(t *testing.T) {
    def := &techdef.TechDef{
        Name:      "Test",
        Structure: []techdef.StructureEntry{{Path: "../escape/"}},
        Gitignore: "*.log",
    }
    err := def.Validate("test")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "must not contain '..'")
}
```

**Testing structure entry type detection:**

```go
func TestStructureEntryIsDir(t *testing.T) {
    dirEntry := techdef.StructureEntry{Path: "src/public/"}
    fileEntry := techdef.StructureEntry{Path: "src/MyModule.psm1"}

    assert.True(t, dirEntry.IsDir())
    assert.False(t, fileEntry.IsDir())
}
```

**Testing structure creation with directories:**

```go
func TestCreateStructure_Directories(t *testing.T) {
    dir := t.TempDir()

    structure := []techdef.StructureEntry{
        {Path: "src/classes/"},
        {Path: "src/private/"},
    }

    err := scaffold.CreateStructure(dir, structure)
    require.NoError(t, err)

    assert.DirExists(t, filepath.Join(dir, "src/classes"))
    assert.FileExists(t, filepath.Join(dir, "src/classes/.gitkeep"))
    assert.DirExists(t, filepath.Join(dir, "src/private"))
    assert.FileExists(t, filepath.Join(dir, "src/private/.gitkeep"))

    // .gitkeep must be empty
    content, err := os.ReadFile(filepath.Join(dir, "src/classes/.gitkeep"))
    require.NoError(t, err)
    assert.Empty(t, content)
}
```

**Testing structure creation with files:**

```go
func TestCreateStructure_FileWithContent(t *testing.T) {
    dir := t.TempDir()

    content := "# Hello\n"
    structure := []techdef.StructureEntry{
        {Path: "src/README.md", Content: &content},
    }

    err := scaffold.CreateStructure(dir, structure)
    require.NoError(t, err)

    assert.FileExists(t, filepath.Join(dir, "src/README.md"))
    actual, err := os.ReadFile(filepath.Join(dir, "src/README.md"))
    require.NoError(t, err)
    assert.Equal(t, "# Hello\n", string(actual))
}

func TestCreateStructure_EmptyFile(t *testing.T) {
    dir := t.TempDir()

    structure := []techdef.StructureEntry{
        {Path: "config.json"},
    }

    err := scaffold.CreateStructure(dir, structure)
    require.NoError(t, err)

    assert.FileExists(t, filepath.Join(dir, "config.json"))
    content, err := os.ReadFile(filepath.Join(dir, "config.json"))
    require.NoError(t, err)
    assert.Empty(t, content)
}
```

**Testing gitignore composition:**

```go
func TestComposeGitignore(t *testing.T) {
    tech := &techdef.TechDef{
        Name:      "PowerShell Module",
        Gitignore: "*.ps1xml\n*.nupkg\n",
    }

    result := scaffold.ComposeGitignore(tech)
    assert.Contains(t, result, "# PowerShell Module")
    assert.Contains(t, result, "*.ps1xml")
    assert.Contains(t, result, "*.nupkg")
}
```

**Testing template rendering:**

```go
func TestRenderReadme(t *testing.T) {
    data := templates.TemplateData{
        ProjectName: "my-module",
    }
    result, err := templates.Render("README.md.tmpl", data)
    require.NoError(t, err)
    assert.Equal(t, "# my-module\n", result)
}
```

**Testing config validation:**

```go
func TestConfigRequiresProjectName(t *testing.T) {
    cfg := &config.Config{
        ProjectName: "",
    }
    err := cfg.Validate()
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "project name")
}
```

**Testing directory-already-exists check:**

```go
func TestRejectsExistingDirectory(t *testing.T) {
    dir := t.TempDir()
    target := filepath.Join(dir, "my-module")
    require.NoError(t, os.Mkdir(target, 0o755))

    tech := loadPowerShellDef(t) // helper to load the powershell definition

    cfg := &config.Config{
        ProjectName: "my-module",
        Provider:    "github",
        Technology:  "powershell",
        Licence:     "none",
        Confirmed:   true,
    }

    err := scaffold.Generate(cfg, tech, dir)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "already exists")
}
```

**Testing cleanup on failure:**

```go
func TestCleansUpOnFailure(t *testing.T) {
    dir := t.TempDir()
    target := filepath.Join(dir, "my-module")

    tech := loadPowerShellDef(t)

    cfg := &config.Config{
        ProjectName: "my-module",
        Provider:    "github",
        Technology:  "powershell",
        Licence:     "mit",
        Confirmed:   true,
    }

    // Create the target dir and make a subdirectory read-only to force failure
    require.NoError(t, os.MkdirAll(filepath.Join(target, "src"), 0o755))
    require.NoError(t, os.Chmod(filepath.Join(target, "src"), 0o444))

    err := scaffold.Generate(cfg, tech, dir)
    assert.Error(t, err)
    // After failure, the target directory should have been cleaned up
    assert.NoDirExists(t, target)
}
```

### 7.4 Acceptance Tests

Acceptance tests live in `tests/acceptance/` and verify the complete scaffold output. They invoke the scaffold generation programmatically and assert on the resulting filesystem. See ADR-011 for the rationale.

The acceptance tests use the `scaffold.Generate` function directly, passing a `config.Config` struct, a loaded `TechDef`, and a base directory (using `t.TempDir()` for isolation). They then walk the output directory and assert:

1. Exactly the expected set of files and directories exists (no more, no fewer).
2. File contents match expected values.

#### Test helper

```go
// loadPowerShellDef loads the powershell technology definition for use in tests.
func loadPowerShellDef(t *testing.T) *techdef.TechDef {
    t.Helper()
    defs, err := techdef.Load()
    require.NoError(t, err)
    require.Contains(t, defs, "powershell")
    return defs["powershell"]
}

// collectFiles walks a directory and returns all file paths relative to root.
func collectFiles(t *testing.T, root string) []string {
    t.Helper()
    var files []string
    err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
        if err != nil {
            return err
        }
        if !d.IsDir() {
            rel, _ := filepath.Rel(root, path)
            files = append(files, rel)
        }
        return nil
    })
    require.NoError(t, err)
    return files
}
```

#### Acceptance test: maximum selections

```go
func TestMaximumSelections(t *testing.T) {
    baseDir := t.TempDir()
    tech := loadPowerShellDef(t)

    cfg := &config.Config{
        ProjectName: "my-module",
        Provider:    "github",
        Technology:  "powershell",
        Licence:     "mit",
        Docs:        []string{"contributing"},
        Tooling:     []string{"editorconfig", "gitattributes"},
        RepoConfig:  []string{"issue_templates", "pr_template", "dependabot"},
        Confirmed:   true,
    }

    err := scaffold.Generate(cfg, tech, baseDir)
    require.NoError(t, err)

    root := filepath.Join(baseDir, "my-module")

    expectedFiles := []string{
        ".devcontainer/devcontainer.json",
        ".devcontainer/Dockerfile",
        ".devcontainer/setup.sh",
        ".editorconfig",
        ".gitattributes",
        ".github/dependabot.yml",
        ".github/ISSUE_TEMPLATE/bug_report.yaml",
        ".github/ISSUE_TEMPLATE/feature_request.yaml",
        ".github/pull_request_template.md",
        ".github/workflows/ci.yml",
        ".gitignore",
        "CONTRIBUTING.md",
        "LICENSE",
        "README.md",
        "src/classes/.gitkeep",
        "src/private/.gitkeep",
        "src/public/.gitkeep",
        "tests/integration-tests/.gitkeep",
        "tests/unit-tests/private/.gitkeep",
        "tests/unit-tests/public/.gitkeep",
    }

    actualFiles := collectFiles(t, root)
    assert.ElementsMatch(t, expectedFiles, actualFiles,
        "scaffold output does not match expected file tree")

    // Assert specific file contents
    readme, err := os.ReadFile(filepath.Join(root, "README.md"))
    require.NoError(t, err)
    assert.Equal(t, "# my-module\n", string(readme))

    licence, err := os.ReadFile(filepath.Join(root, "LICENSE"))
    require.NoError(t, err)
    assert.Contains(t, string(licence), "MIT License")
    assert.Contains(t, string(licence), "my-module")

    gitignore, err := os.ReadFile(filepath.Join(root, ".gitignore"))
    require.NoError(t, err)
    assert.Contains(t, string(gitignore), "# PowerShell Module")
    assert.Contains(t, string(gitignore), "*.ps1xml")

    contributing, err := os.ReadFile(filepath.Join(root, "CONTRIBUTING.md"))
    require.NoError(t, err)
    assert.Contains(t, string(contributing), "Contributing to my-module")

    // Assert devcontainer contents
    dcJSON, err := os.ReadFile(filepath.Join(root, ".devcontainer/devcontainer.json"))
    require.NoError(t, err)
    assert.Contains(t, string(dcJSON), "my-module")
    assert.Contains(t, string(dcJSON), "ghcr.io/devcontainers/features/node:1")
    assert.Contains(t, string(dcJSON), "ghcr.io/devcontainers/features/powershell:1")
    assert.Contains(t, string(dcJSON), "ms-vscode.powershell")

    setupSh, err := os.ReadFile(filepath.Join(root, ".devcontainer/setup.sh"))
    require.NoError(t, err)
    assert.Contains(t, string(setupSh), "# === Base tooling ===")
    assert.Contains(t, string(setupSh), "markdownlint-cli2")
    assert.Contains(t, string(setupSh), "# === PowerShell Module ===")
    assert.Contains(t, string(setupSh), "Install-Module -Name Pester")

    // Assert setup.sh is executable
    info, err := os.Stat(filepath.Join(root, ".devcontainer/setup.sh"))
    require.NoError(t, err)
    assert.True(t, info.Mode().Perm()&0o111 != 0, "setup.sh must be executable")
}
```

#### Acceptance test: minimum selections

```go
func TestMinimumSelections(t *testing.T) {
    baseDir := t.TempDir()
    tech := loadPowerShellDef(t)

    cfg := &config.Config{
        ProjectName: "bare-project",
        Provider:    "github",
        Technology:  "powershell",
        Licence:     "none",
        Docs:        []string{},
        Tooling:     []string{},
        RepoConfig:  []string{},
        Confirmed:   true,
    }

    err := scaffold.Generate(cfg, tech, baseDir)
    require.NoError(t, err)

    root := filepath.Join(baseDir, "bare-project")

    expectedFiles := []string{
        ".devcontainer/devcontainer.json",
        ".devcontainer/Dockerfile",
        ".devcontainer/setup.sh",
        ".github/workflows/ci.yml",
        ".gitignore",
        "README.md",
        "src/classes/.gitkeep",
        "src/private/.gitkeep",
        "src/public/.gitkeep",
        "tests/integration-tests/.gitkeep",
        "tests/unit-tests/private/.gitkeep",
        "tests/unit-tests/public/.gitkeep",
    }

    actualFiles := collectFiles(t, root)
    assert.ElementsMatch(t, expectedFiles, actualFiles,
        "scaffold output does not match expected file tree")

    // Assert conditional files do not exist
    assert.NoFileExists(t, filepath.Join(root, "LICENSE"))
    assert.NoFileExists(t, filepath.Join(root, "CONTRIBUTING.md"))
    assert.NoFileExists(t, filepath.Join(root, ".editorconfig"))
    assert.NoFileExists(t, filepath.Join(root, ".gitattributes"))

    readme, err := os.ReadFile(filepath.Join(root, "README.md"))
    require.NoError(t, err)
    assert.Equal(t, "# bare-project\n", string(readme))
}
```

#### Acceptance test: confirmation declined

```go
func TestConfirmationDeclined(t *testing.T) {
    baseDir := t.TempDir()

    cfg := &config.Config{
        ProjectName: "declined-project",
        Provider:    "github",
        Technology:  "powershell",
        Licence:     "mit",
        Confirmed:   false,
    }

    // Generate should not be called when Confirmed is false.
    // This test verifies the calling code's behaviour, not Generate itself.
    // The main function is responsible for checking cfg.Confirmed.
    assert.False(t, cfg.Confirmed)
    assert.NoDirExists(t, filepath.Join(baseDir, "declined-project"))
}
```

### 7.5 Running Tests

```bash
# Unit tests only
go test ./internal/...

# Acceptance tests only
go test ./tests/acceptance/

# All tests
go test ./...
```

---

## 8. CI Pipeline (doozer-scaffold's own)

The tool's own GitHub Actions workflow at `.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: stable
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v6

  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: stable
      - name: Unit tests
        run: go test ./internal/... -v
      - name: Acceptance tests
        run: go test ./tests/acceptance/ -v

  build:
    runs-on: ubuntu-latest
    needs: [lint, test]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: stable
      - name: Build
        run: go build -o doozer-scaffold ./cmd/doozer-scaffold
```

### 8.1 golangci-lint Configuration

A `.golangci.yml` at the project root:

```yaml
run:
  timeout: 5m

linters:
  enable:
    - errcheck
    - govet
    - staticcheck
    - unused
    - gosimple
    - ineffassign
```

---

## 9. Makefile

A `Makefile` at the project root for local development convenience:

```makefile
.PHONY: build test lint clean

build:
	go build -o bin/doozer-scaffold ./cmd/doozer-scaffold

test:
	go test ./... -v

test-unit:
	go test ./internal/... -v

test-acceptance:
	go test ./tests/acceptance/ -v

lint:
	golangci-lint run

clean:
	rm -rf bin/
```

---

## 10. Implementation Order

This section defines the order in which features should be implemented, following TDD. Each step begins with writing the test, then the production code.

### Step 1: Project skeleton and build

1. Initialise the Go module: `go mod init github.com/tonylea/doozer-scaffold`.
2. Create `cmd/doozer-scaffold/main.go` with a minimal `main()` that prints a placeholder message.
3. Create the `internal/` package structure with empty files.
4. Verify `go build ./cmd/doozer-scaffold` succeeds.
5. Create the `Makefile`.

### Step 2: Technology definition loading and validation

1. Create the `technologies/powershell.yaml` definition file.
2. Write tests for `techdef.Load()` — verify it reads and parses the definition correctly.
3. Write tests for `TechDef.Validate()` — verify all validation rules (missing name, empty structure, absolute paths, path traversal).
4. Write tests for `StructureEntry.IsDir()`.
5. Implement the `techdef` package.

### Step 3: Config and validation

1. Write tests for `config.Config` validation (project name required, etc.).
2. Implement the `Config` struct and `Validate()` method.

### Step 4: Template loading and rendering

1. Write tests for loading and rendering each template (README, LICENSE, CI config, etc.).
2. Create all template files in the `templates/` directory.
3. Implement the template loading and rendering functions using `go:embed`.

### Step 5: Scaffold generation

1. Write tests for `scaffold.Generate` — starting with the minimum-selection case, then maximum-selection.
2. Write tests for structure creation (directories with `.gitkeep`, files with content, empty files).
3. Write tests for gitignore composition.
4. Write tests for devcontainer composition (JSON rendering, setup.sh assembly).
5. Write tests for error conditions (directory exists, cleanup on failure).
6. Implement `scaffold.Generate` and its internal file-creation logic. The engine must be entirely data-driven — no technology-specific `if` statements.

### Step 6: Interactive prompts

1. Implement the prompt flow using `huh`, with technology options populated dynamically from loaded definitions.
2. Wire prompts into `main.go`.
3. Manual testing of the interactive flow.

### Step 7: Acceptance tests

1. Write the full acceptance tests (maximum and minimum selection scenarios).
2. Verify they pass against the implementation from Steps 2–5.
3. Add any edge-case acceptance tests identified during implementation.

### Step 8: CI and linting

1. Create `.github/workflows/ci.yml` for the project's own CI.
2. Create `.golangci.yml`.
3. Fix any linting issues.
4. Push and verify CI passes.

---

## 11. Acceptance Criteria

Stage 1 is complete when all of the following are true:

1. Running `doozer-scaffold` presents the interactive prompt flow described in Section 4.
2. Running `doozer-scaffold my-module` skips the project name prompt and uses `my-module`.
3. The technology prompt is dynamically populated from YAML definitions in `technologies/`.
4. Selecting all options and confirming produces the exact file tree in Section 5.10.
5. Selecting minimum options and confirming produces the exact file tree in Section 5.11.
6. All generated files contain the correct content as specified in Sections 5.3–5.9.
7. Template variables (`ProjectName`, `Year`) are correctly substituted in all templated files.
8. The technology's directory structure and files are created exactly as defined in `powershell.yaml`.
9. The `.gitignore` is composed from the technology definition's `gitignore` field with the correct header.
10. The devcontainer is composed from the base layer plus the technology definition's `devcontainer` field.
11. The `setup.sh` is executable and correctly composes base + technology setup blocks.
12. Attempting to scaffold into an existing directory produces a clear error and no files.
13. Declining confirmation produces no files and a clear message.
14. If generation fails partway through, the partially created directory is removed.
15. Technology definition validation catches invalid definitions (missing name, bad paths, etc.).
16. All unit tests pass.
17. All acceptance tests pass.
18. The project's own CI pipeline runs lint, unit tests, and acceptance tests successfully.
19. `go build` produces a working binary.
20. Adding a hypothetical new technology requires only a new YAML file in `technologies/` and corresponding acceptance tests — no changes to the scaffold engine or prompt code.