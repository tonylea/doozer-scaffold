# SPEC — doozer-scaffold Stage 2: Multi-Technology Support

**Version:** 1.0  
**Status:** Draft  
**Parent PRD:** doozer-scaffold PRD v0.2-draft  
**Related ADR:** ADR Stage 1, ADR Stage 2  
**Owner:** Tony Lea

---

## 1. Purpose

This document is the complete technical specification for Stage 2 of `doozer-scaffold`. It defines every input, output, behaviour, and test expectation in sufficient detail that an implementer can build the stage without ambiguity.

Stage 2 extends the tool to support selecting multiple technologies simultaneously, adds four new technology definitions (Go, Terraform Module, Terraform Infrastructure, Python), introduces the standalone/composable technology model, adds technology-driven prompts, composes per-technology CI jobs, generates a universal Makefile, and produces correctly merged output across all selections.

---

## 2. Summary of Changes from Stage 1

### 2.1 Schema Changes

| Change                                          | Detail                                                                               |
| ----------------------------------------------- | ------------------------------------------------------------------------------------ |
| New YAML field: `standalone`                    | Boolean. Default `false`. When `true`, technology cannot be combined with others.    |
| New YAML field: `prompts`                       | List of prompt definitions. Allows technologies to request additional user input.    |
| New YAML field: `ci`                            | CI job definition. Each technology contributes a CI job with setup action and steps. |
| Updated definition: `powershell.yaml`           | Add `standalone: true`. Add `ci` field.                                              |
| New definition: `go.yaml`                       | Go CLI/library. Composable.                                                          |
| New definition: `terraform-module.yaml`         | Terraform registry module. Standalone.                                               |
| New definition: `terraform-infrastructure.yaml` | Terraform as supporting infrastructure. Composable.                                  |
| New definition: `python.yaml`                   | Python package. Composable. Uses `prompts` for package name.                         |

### 2.2 Config Changes

| Change                                             | Detail                                                            |
| -------------------------------------------------- | ----------------------------------------------------------------- |
| `Technology string` → `Technologies []string`      | Multi-select replaces single-select.                              |
| New field: `TechPromptResponses map[string]string` | Stores responses to technology-driven prompts.                    |
| Validation extended                                | At least one technology required. Standalone constraint enforced. |

### 2.3 Prompt Changes

| Change                                          | Detail                                                                  |
| ----------------------------------------------- | ----------------------------------------------------------------------- |
| Technology prompt: single-select → multi-select | User can select multiple composable technologies.                       |
| Standalone enforcement                          | Selecting a standalone technology disables all other selections.        |
| Technology-driven prompts                       | Additional prompts inserted after technology selection, before licence. |

### 2.4 Engine Changes

| Change                                   | Detail                                                                                       |
| ---------------------------------------- | -------------------------------------------------------------------------------------------- |
| `Generate` accepts `[]*techdef.TechDef`  | Replaces single `*techdef.TechDef`.                                                          |
| Template processing in structure entries | Paths and content fields support template variables from prompt responses.                   |
| Composite gitignore                      | Sections from all selected technologies, alphabetical by name.                               |
| Composite devcontainer                   | Features, extensions, and setup.sh merged from all selected technologies.                    |
| Composite CI config                      | Three-stage pipeline (lint → test → build) with per-technology jobs in lint and test stages. |
| Composite structure                      | All structure entries from all selected technologies created.                                |
| Path conflict detection                  | Error if two technologies define a file at the same path.                                    |
| Universal Makefile                       | Empty Makefile generated for every scaffold.                                                 |

---

## 3. Technology Definition Schema — Updated

### 3.1 Full Schema (Stage 2)

```yaml
# Required. Display name shown in the interactive prompt.
name: "Python"

# Optional. Default: false.
# When true, this technology cannot be combined with any other technology.
standalone: false

# Optional. Additional prompts to present to the user after technology selection.
prompts:
  - key: "package_name"          # Unique key, used as template variable name
    title: "Python package name:" # Prompt text shown to user
    type: "text"                  # text | select | multi_select
    default_from: "project_name"  # Derive default from project name (with sanitisation)
    # For select/multi_select types:
    # options:
    #   - label: "Display text"
    #     value: "stored_value"

# Required. Files and directories to create.
# Paths and content fields support template variables: {{.ProjectName}}, {{.Year}},
# and any keys defined in the prompts section (e.g. {{.package_name}}).
structure:
  - path: "src/{{.package_name}}/"
  - path: "src/{{.package_name}}/__init__.py"
    content: |
      """{{.package_name}} package."""

# Required. Lines to include in the composite .gitignore.
gitignore: |
  __pycache__/
  *.py[cod]

# Required. Devcontainer configuration contributed by this technology.
devcontainer:
  features:
    "ghcr.io/devcontainers/features/python:1": {}
  extensions:
    - "ms-python.python"
  setup: |
    pip install uv

# Optional. CI job definition contributed by this technology.
ci:
  job_name: "python"
  setup_steps:
    - name: "Set up Python"
      uses: "actions/setup-python@v5"
      with:
        python-version: "3.12"
  lint_steps:
    - name: "Lint"
      run: "pip install ruff && ruff check ."
  test_steps:
    - name: "Test"
      run: "pip install pytest && pytest"
```

### 3.2 Schema Rules — New Fields

**`standalone`:**

- Boolean. Default `false` if absent.
- When `true`, the prompt prevents this technology from being selected alongside any other technology.
- Config validation also enforces this as a defence-in-depth check.

**`prompts`:**

- A list of prompt definitions. Each entry has:
  - `key` (required): string. Used as the template variable name. Must be a valid Go template identifier (alphanumeric and underscores, starting with a letter). Must be unique across all selected technologies.
  - `title` (required): string. The prompt text shown to the user.
  - `type` (required): one of `text`, `select`, `multi_select`.
  - `default_from` (optional): string. Currently the only supported value is `"project_name"`, which derives a default from the project name with sanitisation appropriate to the context (see Section 3.3).
  - `options` (required for `select` and `multi_select`): list of objects, each with `label` (display text) and `value` (stored value).
- For `text` type: the prompt is a free-text input. If `default_from` is set, the input is pre-populated with the derived default. The user can accept or change it.
- For `select` type: the prompt is a single-select from the provided options.
- For `multi_select` type: the prompt is a multi-select from the provided options.
- Prompt responses are stored in `Config.TechPromptResponses` keyed by the prompt's `key`.
- If no technologies define prompts, no additional prompt steps are added to the flow.

**`ci`:**

- An object defining this technology's CI job contribution. The engine produces two jobs per technology: `lint-{job_name}` and `test-{job_name}`, plus a single shared `build` job.
- `job_name` (required): string. Used as a suffix in the GitHub Actions job IDs (`lint-{job_name}`, `test-{job_name}`).
- `setup_steps` (optional): list of objects, each with:
  - `name` (required): string. Step display name.
  - `uses` (optional): string. A GitHub Actions action reference (e.g. `actions/setup-go@v5`).
  - `with` (optional): map of string to string. Parameters for the action.
  - `run` (optional): string. Shell command to execute.
  - Each step must have either `uses` or `run` (not both, not neither).
- `lint_steps` (required): list of objects, each with:
  - `name` (required): string. Step display name.
  - `run` (required): string. Shell command to execute.
- `test_steps` (required): list of objects, each with:
  - `name` (required): string. Step display name.
  - `run` (required): string. Shell command to execute.
- If `ci` is absent, the technology does not contribute any CI jobs.
- Setup steps are prepended to both the lint and test jobs (after checkout). This ensures both jobs have the same environment.

### 3.3 Default Derivation and Sanitisation

When `default_from: "project_name"` is specified, the engine derives a default value from the project name. The derivation applies sanitisation rules appropriate for common identifier constraints:

1. Replace hyphens with underscores.
2. Replace any character that is not alphanumeric or underscore with underscore.
3. Strip leading digits and underscores (identifiers must start with a letter).
4. If the result is empty after sanitisation, fall back to `"app"`.

```go
func SanitiseForIdentifier(projectName string) string {
    result := strings.ReplaceAll(projectName, "-", "_")
    cleaned := strings.Map(func(r rune) rune {
        if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
            return r
        }
        return '_'
    }, result)
    cleaned = strings.TrimLeftFunc(cleaned, func(r rune) bool {
        return !unicode.IsLetter(r)
    })
    if cleaned == "" {
        return "app"
    }
    return strings.ToLower(cleaned)
}
```

Examples:

| Project name    | Derived default |
| --------------- | --------------- |
| `my-project`    | `my_project`    |
| `MyApp`         | `myapp`         |
| `123-bad-start` | `bad_start`     |
| `---`           | `app`           |

### 3.4 Go Data Structure — Updated

```go
type TechDef struct {
    Name         string            `yaml:"name"`
    Standalone   bool              `yaml:"standalone"`
    Prompts      []PromptDef       `yaml:"prompts,omitempty"`
    Structure    []StructureEntry  `yaml:"structure"`
    Gitignore    string            `yaml:"gitignore"`
    Devcontainer DevcontainerDef   `yaml:"devcontainer"`
    CI           *CIDef            `yaml:"ci,omitempty"`
}

type PromptDef struct {
    Key         string       `yaml:"key"`
    Title       string       `yaml:"title"`
    Type        string       `yaml:"type"`        // "text", "select", "multi_select"
    DefaultFrom string       `yaml:"default_from,omitempty"`
    Options     []OptionDef  `yaml:"options,omitempty"`
}

type OptionDef struct {
    Label string `yaml:"label"`
    Value string `yaml:"value"`
}

type CIDef struct {
    JobName    string        `yaml:"job_name"`
    SetupSteps []CISetupStep `yaml:"setup_steps,omitempty"`
    LintSteps  []CIStep      `yaml:"lint_steps"`
    TestSteps  []CIStep      `yaml:"test_steps"`
}

type CISetupStep struct {
    Name string            `yaml:"name"`
    Uses string            `yaml:"uses,omitempty"`
    With map[string]string `yaml:"with,omitempty"`
    Run  string            `yaml:"run,omitempty"`
}

type CIStep struct {
    Name string `yaml:"name"`
    Run  string `yaml:"run"`
}

// StructureEntry, DevcontainerDef unchanged from Stage 1.
```

### 3.5 Validation — Updated

Add to the existing validation rules from Stage 1:

7. `standalone` is a boolean and requires no validation beyond type parsing. Absence defaults to `false`.
8. Each `prompts` entry must have a non-empty `key`, non-empty `title`, and `type` must be one of `text`, `select`, `multi_select`.
9. `prompts[].key` must match `^[a-zA-Z][a-zA-Z0-9_]*$` (valid Go template identifier).
10. If `type` is `select` or `multi_select`, `options` must be non-empty, and each option must have a non-empty `label` and `value`.
11. If `ci` is present: `ci.job_name` must be non-empty; `ci.lint_steps` must contain at least one entry; `ci.test_steps` must contain at least one entry. Each lint and test step must have non-empty `name` and `run`. Each `setup_steps` entry must have a non-empty `name` and either `uses` or `run` (not both, not neither).

```go
func (t *TechDef) Validate(key string) error {
    // ... existing validations from Stage 1 ...

    // Prompt validation
    keyPattern := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)
    for i, p := range t.Prompts {
        if p.Key == "" {
            return fmt.Errorf("technology '%s': prompts[%d] key is required", key, i)
        }
        if !keyPattern.MatchString(p.Key) {
            return fmt.Errorf("technology '%s': prompts[%d] key '%s' is not a valid identifier", key, i, p.Key)
        }
        if p.Title == "" {
            return fmt.Errorf("technology '%s': prompts[%d] title is required", key, i)
        }
        if p.Type != "text" && p.Type != "select" && p.Type != "multi_select" {
            return fmt.Errorf("technology '%s': prompts[%d] type must be text, select, or multi_select", key, i)
        }
        if (p.Type == "select" || p.Type == "multi_select") && len(p.Options) == 0 {
            return fmt.Errorf("technology '%s': prompts[%d] options required for type '%s'", key, i, p.Type)
        }
    }

    // CI validation
    if t.CI != nil {
        if t.CI.JobName == "" {
            return fmt.Errorf("technology '%s': ci.job_name is required", key)
        }
        for i, step := range t.CI.SetupSteps {
            if step.Name == "" {
                return fmt.Errorf("technology '%s': ci.setup_steps[%d] name is required", key, i)
            }
            if step.Uses == "" && step.Run == "" {
                return fmt.Errorf("technology '%s': ci.setup_steps[%d] must have either 'uses' or 'run'", key, i)
            }
            if step.Uses != "" && step.Run != "" {
                return fmt.Errorf("technology '%s': ci.setup_steps[%d] must not have both 'uses' and 'run'", key, i)
            }
        }
        if len(t.CI.LintSteps) == 0 {
            return fmt.Errorf("technology '%s': ci.lint_steps must contain at least one step", key)
        }
        for i, step := range t.CI.LintSteps {
            if step.Name == "" {
                return fmt.Errorf("technology '%s': ci.lint_steps[%d] name is required", key, i)
            }
            if step.Run == "" {
                return fmt.Errorf("technology '%s': ci.lint_steps[%d] run is required", key, i)
            }
        }
        if len(t.CI.TestSteps) == 0 {
            return fmt.Errorf("technology '%s': ci.test_steps must contain at least one step", key)
        }
        for i, step := range t.CI.TestSteps {
            if step.Name == "" {
                return fmt.Errorf("technology '%s': ci.test_steps[%d] name is required", key, i)
            }
            if step.Run == "" {
                return fmt.Errorf("technology '%s': ci.test_steps[%d] run is required", key, i)
            }
        }
    }

    return nil
}
```

---

## 4. Technology Definitions

### 4.1 PowerShell Module — Updated

File: `technologies/powershell.yaml`

Changes from Stage 1: add `standalone: true`, add `ci` field.

```yaml
name: "PowerShell Module"
standalone: true

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

ci:
  job_name: "powershell"
  setup_steps:
    - name: "Install PowerShell"
      run: |
        sudo apt-get update
        sudo apt-get install -y wget apt-transport-https software-properties-common
        wget -q "https://packages.microsoft.com/config/ubuntu/$(lsb_release -rs)/packages-microsoft-prod.deb"
        sudo dpkg -i packages-microsoft-prod.deb
        sudo apt-get update
        sudo apt-get install -y powershell
  lint_steps:
    - name: "Lint"
      run: |
        pwsh -NoProfile -Command "
          Install-Module -Name PSScriptAnalyzer -Force -Scope CurrentUser
          Invoke-ScriptAnalyzer -Path ./src -Recurse -EnableExit
        "
  test_steps:
    - name: "Test"
      run: |
        pwsh -NoProfile -Command "
          Install-Module -Name Pester -Force -Scope CurrentUser
          Invoke-Pester -CI
        "
```

### 4.2 Go

File: `technologies/go.yaml`

```yaml
name: "Go"
standalone: false

structure:
  - path: "cmd/app/"
  - path: "internal/"

gitignore: |
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

devcontainer:
  features:
    "ghcr.io/devcontainers/features/go:1": {}
  extensions:
    - "golang.go"
  setup: |
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

ci:
  job_name: "go"
  setup_steps:
    - name: "Set up Go"
      uses: "actions/setup-go@v5"
      with:
        go-version: "stable"
  lint_steps:
    - name: "Lint"
      run: |
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
        golangci-lint run
  test_steps:
    - name: "Test"
      run: "go test ./... -v"
```

This produces:

```
{project}/
├── cmd/
│   └── app/
│       └── .gitkeep
└── internal/
    └── .gitkeep
```

### 4.3 Terraform Module

File: `technologies/terraform-module.yaml`

```yaml
name: "Terraform Module"
standalone: true

structure:
  - path: "modules/"
  - path: "examples/"
  - path: "main.tf"
    content: |
      # Main Terraform configuration
  - path: "variables.tf"
    content: |
      # Input variables
  - path: "outputs.tf"
    content: |
      # Output values
  - path: "versions.tf"
    content: |
      terraform {
        required_version = ">= 1.0"
      }

gitignore: |
  # Terraform
  .terraform/
  .terraform.lock.hcl
  *.tfstate
  *.tfstate.*
  *.tfplan
  *.tfvars
  !*.tfvars.example
  crash.log
  crash.*.log
  override.tf
  override.tf.json
  *_override.tf
  *_override.tf.json

devcontainer:
  features:
    "ghcr.io/devcontainers-contrib/features/terraform-asdf:1": {}
  extensions:
    - "hashicorp.terraform"
  setup: |
    curl -s https://raw.githubusercontent.com/terraform-linters/tflint/master/install_linux.sh | bash
    tflint --init || true

ci:
  job_name: "terraform"
  setup_steps:
    - name: "Install Terraform"
      run: |
        sudo apt-get update && sudo apt-get install -y unzip
        curl -sL https://releases.hashicorp.com/terraform/1.9.0/terraform_1.9.0_linux_amd64.zip -o tf.zip
        unzip tf.zip && sudo mv terraform /usr/local/bin/
    - name: "Install TFLint"
      run: |
        curl -s https://raw.githubusercontent.com/terraform-linters/tflint/master/install_linux.sh | bash
        tflint --init || true
  lint_steps:
    - name: "Format check"
      run: "terraform fmt -check -recursive"
    - name: "Lint"
      run: "tflint"
  test_steps:
    - name: "Validate"
      run: |
        terraform init -backend=false
        terraform validate
```

This produces:

```
{project}/
├── examples/
│   └── .gitkeep
├── main.tf
├── modules/
│   └── .gitkeep
├── outputs.tf
├── variables.tf
└── versions.tf
```

### 4.4 Terraform Infrastructure

File: `technologies/terraform-infrastructure.yaml`

```yaml
name: "Terraform (Infrastructure)"
standalone: false

structure:
  - path: "infrastructure/"
  - path: "infrastructure/main.tf"
    content: |
      # Main Terraform configuration
  - path: "infrastructure/variables.tf"
    content: |
      # Input variables
  - path: "infrastructure/outputs.tf"
    content: |
      # Output values
  - path: "infrastructure/versions.tf"
    content: |
      terraform {
        required_version = ">= 1.0"
      }

gitignore: |
  # Terraform
  .terraform/
  .terraform.lock.hcl
  *.tfstate
  *.tfstate.*
  *.tfplan
  *.tfvars
  !*.tfvars.example
  crash.log
  crash.*.log
  override.tf
  override.tf.json
  *_override.tf
  *_override.tf.json

devcontainer:
  features:
    "ghcr.io/devcontainers-contrib/features/terraform-asdf:1": {}
  extensions:
    - "hashicorp.terraform"
  setup: |
    curl -s https://raw.githubusercontent.com/terraform-linters/tflint/master/install_linux.sh | bash
    tflint --init || true

ci:
  job_name: "terraform"
  setup_steps:
    - name: "Install Terraform"
      run: |
        sudo apt-get update && sudo apt-get install -y unzip
        curl -sL https://releases.hashicorp.com/terraform/1.9.0/terraform_1.9.0_linux_amd64.zip -o tf.zip
        unzip tf.zip && sudo mv terraform /usr/local/bin/
    - name: "Install TFLint"
      run: |
        curl -s https://raw.githubusercontent.com/terraform-linters/tflint/master/install_linux.sh | bash
        tflint --init || true
  lint_steps:
    - name: "Format check"
      run: "terraform -chdir=infrastructure fmt -check -recursive"
    - name: "Lint"
      run: "tflint --chdir=infrastructure"
  test_steps:
    - name: "Validate"
      run: |
        terraform -chdir=infrastructure init -backend=false
        terraform -chdir=infrastructure validate
```

This produces:

```
{project}/
└── infrastructure/
    ├── .gitkeep
    ├── main.tf
    ├── outputs.tf
    ├── variables.tf
    └── versions.tf
```

Note: The `infrastructure/` directory entry creates a `.gitkeep`. The file entries also create files in the same directory. Both are created — `.gitkeep` from the directory entry, `.tf` files from the file entries. This is not a conflict — they are different filenames in the same directory.

### 4.5 Python

File: `technologies/python.yaml`

```yaml
name: "Python"
standalone: false

prompts:
  - key: "package_name"
    title: "Python package name:"
    type: "text"
    default_from: "project_name"

structure:
  - path: "src/{{.package_name}}/"
  - path: "src/{{.package_name}}/__init__.py"
    content: |
      """{{.package_name}} package."""
  - path: "tests/"
  - path: "tests/__init__.py"
    content: ""
  - path: "pyproject.toml"
    content: |
      [project]
      name = "{{.package_name}}"
      version = "0.1.0"
      description = ""
      requires-python = ">=3.12"
      dependencies = []

      [project.optional-dependencies]
      dev = [
          "pytest>=8.0",
          "ruff>=0.4",
      ]

      [tool.ruff]
      line-length = 120
      target-version = "py312"

      [tool.ruff.lint]
      select = ["E", "F", "I", "N", "W", "UP"]

      [tool.pytest.ini_options]
      testpaths = ["tests"]

      [build-system]
      requires = ["hatchling"]
      build-backend = "hatchling.backends"

gitignore: |
  # Python
  __pycache__/
  *.py[cod]
  *$py.class
  *.so
  .Python
  build/
  dist/
  *.egg-info/
  *.egg
  .eggs/
  # Virtual environments
  .venv/
  venv/
  ENV/
  # Testing
  .pytest_cache/
  .coverage
  htmlcov/
  # Ruff
  .ruff_cache/
  # Distribution
  *.whl

devcontainer:
  features:
    "ghcr.io/devcontainers/features/python:1": {}
  extensions:
    - "charliermarsh.ruff"
    - "ms-python.python"
  setup: |
    pip install --break-system-packages uv
    if [ -f pyproject.toml ]; then uv pip install --system -e ".[dev]"; fi

ci:
  job_name: "python"
  setup_steps:
    - name: "Set up Python"
      uses: "actions/setup-python@v5"
      with:
        python-version: "3.12"
  lint_steps:
    - name: "Lint"
      run: |
        pip install ruff
        ruff check .
  test_steps:
    - name: "Test"
      run: |
        pip install pytest
        pip install -e ".[dev]" || true
        pytest
```

This produces (assuming `package_name` response is `my_project`):

```
{project}/
├── pyproject.toml
├── src/
│   └── my_project/
│       ├── .gitkeep
│       └── __init__.py
└── tests/
    ├── .gitkeep
    └── __init__.py
```

---

## 5. User Interaction Flow — Updated

### 5.1 Prompt Sequence

The prompts are presented in this exact order:

| Step | Type                       | Prompt Text                                | Options                                                                |
| ---- | -------------------------- | ------------------------------------------ | ---------------------------------------------------------------------- |
| 1    | Text input (if no CLI arg) | "Project name:"                            | Free text, required, non-empty                                         |
| 2    | Single-select              | "Remote hosting provider:"                 | `GitHub`                                                               |
| 3    | **Multi-select**           | **"Technologies:"**                        | **Dynamically populated from loaded technology definitions**           |
| 3a   | **Dynamic**                | **Technology-driven prompts (if any)**     | **Populated from selected technologies' `prompts` fields**             |
| 4    | Single-select              | "Licence:"                                 | `MIT`, `None`                                                          |
| 5    | Multi-select               | "Supplementary documentation:"             | `CONTRIBUTING.md`                                                      |
| 6    | Multi-select               | "Project tooling:"                         | `.editorconfig`, `.gitattributes`                                      |
| 7    | Multi-select               | "GitHub repository configuration:"         | `Issue templates`, `Pull request template`, `Dependabot configuration` |
| 8    | Confirm                    | "Generate scaffold with these selections?" | Yes / No                                                               |

**Changes from Stage 1:**

- Step 3: `Select` → `MultiSelect`. Title changes from "Technology:" to "Technologies:".
- Step 3: At least one technology must be selected. The prompt validates this.
- Step 3a (new): Technology-driven prompts. Collected from all selected technologies' `prompts` fields. Presented in alphabetical order by technology key, then in definition order within each technology.

### 5.2 Standalone Enforcement in the Prompt

The standalone constraint is enforced via post-selection validation on the multi-select prompt.

```go
huh.NewMultiSelect[string]().
    Title("Technologies:").
    Options(techOptions...).
    Value(&cfg.Technologies).
    Validate(func(selected []string) error {
        if len(selected) == 0 {
            return fmt.Errorf("at least one technology must be selected")
        }
        if len(selected) > 1 {
            for _, key := range selected {
                if techDefs[key].Standalone {
                    return fmt.Errorf("'%s' is a standalone technology and cannot be combined with others", techDefs[key].Name)
                }
            }
        }
        return nil
    }),
```

### 5.3 Technology-Driven Prompts

After the user selects technologies, the engine collects all `prompts` from the selected technology definitions and presents them. Technologies are processed in alphabetical order by key.

```go
func buildTechPromptGroups(cfg *config.Config, techDefs map[string]*techdef.TechDef) []*huh.Group {
    if cfg.TechPromptResponses == nil {
        cfg.TechPromptResponses = make(map[string]string)
    }

    keys := make([]string, len(cfg.Technologies))
    copy(keys, cfg.Technologies)
    sort.Strings(keys)

    var groups []*huh.Group
    for _, key := range keys {
        def := techDefs[key]
        for _, p := range def.Prompts {
            promptKey := p.Key
            switch p.Type {
            case "text":
                defaultVal := ""
                if p.DefaultFrom == "project_name" {
                    defaultVal = SanitiseForIdentifier(cfg.ProjectName)
                }
                cfg.TechPromptResponses[promptKey] = defaultVal
                groups = append(groups, huh.NewGroup(
                    huh.NewInput().
                        Title(p.Title).
                        Value(&cfg.TechPromptResponses[promptKey]).
                        Validate(func(s string) error {
                            if s == "" {
                                return fmt.Errorf("%s is required", p.Title)
                            }
                            return nil
                        }),
                ))
            case "select":
                options := make([]huh.Option[string], len(p.Options))
                for i, o := range p.Options {
                    options[i] = huh.NewOption(o.Label, o.Value)
                }
                groups = append(groups, huh.NewGroup(
                    huh.NewSelect[string]().
                        Title(p.Title).
                        Options(options...).
                        Value(&cfg.TechPromptResponses[promptKey]),
                ))
            case "multi_select":
                // For multi_select, store as comma-separated values
                var selected []string
                options := make([]huh.Option[string], len(p.Options))
                for i, o := range p.Options {
                    options[i] = huh.NewOption(o.Label, o.Value)
                }
                groups = append(groups, huh.NewGroup(
                    huh.NewMultiSelect[string]().
                        Title(p.Title).
                        Options(options...).
                        Value(&selected),
                ))
                // Note: selected values joined to cfg.TechPromptResponses after form runs
            }
        }
    }
    return groups
}
```

Note on `huh` form structure: the technology-driven prompt groups are inserted into the form between the technology selection group and the licence group. The form is built in a single `huh.NewForm(groups...)` call. The dynamic prompts are known after the technology selection step, so the form is built in two phases: first the technology selection form runs, then the remaining form (including tech prompts) is built and run.

**Implementation approach — two-phase form:**

```go
// Phase 1: Project name, provider, technology selection
phase1Groups := buildPhase1Groups(cfg, techDefs)
if err := huh.NewForm(phase1Groups...).Run(); err != nil {
    return err
}

// Phase 2: Tech prompts, licence, docs, tooling, repo config, confirm
techPromptGroups := buildTechPromptGroups(cfg, techDefs)
phase2Groups := append(techPromptGroups, buildPhase2Groups(cfg)...)
if err := huh.NewForm(phase2Groups...).Run(); err != nil {
    return err
}
```

### 5.4 Config Data Structure — Updated

```go
package config

type Config struct {
    ProjectName        string
    Provider           string
    Technologies       []string          // Changed from Technology string
    TechPromptResponses map[string]string // NEW — responses to technology-driven prompts
    Licence            string
    Docs               []string
    Tooling            []string
    RepoConfig         []string
    Confirmed          bool
}
```

### 5.5 Config Validation — Updated

```go
func (c *Config) Validate(techDefs map[string]*techdef.TechDef) error {
    if strings.TrimSpace(c.ProjectName) == "" {
        return fmt.Errorf("project name is required")
    }
    if len(c.Technologies) == 0 {
        return fmt.Errorf("at least one technology must be selected")
    }
    // Standalone constraint
    if len(c.Technologies) > 1 {
        for _, key := range c.Technologies {
            def, ok := techDefs[key]
            if !ok {
                return fmt.Errorf("unknown technology '%s'", key)
            }
            if def.Standalone {
                return fmt.Errorf("technology '%s' is standalone and cannot be combined with others", def.Name)
            }
        }
    }
    // Validate all technology keys exist
    for _, key := range c.Technologies {
        if _, ok := techDefs[key]; !ok {
            return fmt.Errorf("unknown technology '%s'", key)
        }
    }
    return nil
}
```

---

## 6. Scaffold Engine — Updated

### 6.1 Generate Function Signature

```go
// Generate creates the scaffold in a subdirectory of baseDir named after cfg.ProjectName.
// techs is the ordered list of selected technology definitions (sorted by key).
func Generate(cfg *config.Config, techs []*techdef.TechDef, baseDir string) error
```

The caller resolves the selected technology keys to `*techdef.TechDef` pointers and passes them in. The `techs` slice must be sorted alphabetically by technology key before being passed to `Generate`. This ensures deterministic output.

### 6.2 Template Data for Structure Entries

Structure entry paths and content fields are processed through `text/template` before file creation. The template data includes standard fields plus all technology-driven prompt responses:

```go
func buildTemplateData(cfg *config.Config) map[string]string {
    data := map[string]string{
        "ProjectName": cfg.ProjectName,
        "Year":        strconv.Itoa(time.Now().Year()),
    }
    for k, v := range cfg.TechPromptResponses {
        data[k] = v
    }
    return data
}

func resolveTemplate(tmplStr string, data map[string]string) (string, error) {
    t, err := template.New("").Parse(tmplStr)
    if err != nil {
        return "", err
    }
    var buf bytes.Buffer
    if err := t.Execute(&buf, data); err != nil {
        return "", err
    }
    return buf.String(), nil
}
```

Each structure entry's `Path` and `Content` are resolved before use:

```go
for _, entry := range tech.Structure {
    resolvedPath, err := resolveTemplate(entry.Path, templateData)
    // ...
    if entry.Content != nil {
        resolvedContent, err := resolveTemplate(*entry.Content, templateData)
        // ...
    }
}
```

### 6.3 Path Conflict Detection

Before creating any files, the engine checks for path conflicts across all selected technologies. Path conflict detection runs **after** template resolution so that templated paths (e.g. `src/{{.package_name}}/`) are checked with their actual values.

```go
func DetectPathConflicts(techs []*techdef.TechDef, templateData map[string]string) error {
    seen := make(map[string]string) // resolved path → technology name
    for _, tech := range techs {
        for _, entry := range tech.Structure {
            resolvedPath, err := resolveTemplate(entry.Path, templateData)
            if err != nil {
                return fmt.Errorf("resolving path in '%s': %w", tech.Name, err)
            }
            if strings.HasSuffix(resolvedPath, "/") {
                continue // Directory overlaps are fine
            }
            if existing, ok := seen[resolvedPath]; ok {
                return fmt.Errorf(
                    "path conflict: '%s' is defined by both '%s' and '%s'",
                    resolvedPath, existing, tech.Name,
                )
            }
            seen[resolvedPath] = tech.Name
        }
    }
    return nil
}
```

### 6.4 Composite Gitignore

Unchanged from original SPEC draft. Sections in alphabetical order by technology name:

```go
func ComposeGitignore(techs []*techdef.TechDef) string {
    var sb strings.Builder
    for i, tech := range techs {
        if i > 0 {
            sb.WriteString("\n")
        }
        sb.WriteString("# ")
        sb.WriteString(tech.Name)
        sb.WriteString("\n")
        sb.WriteString(tech.Gitignore)
    }
    return sb.String()
}
```

### 6.5 Composite Devcontainer

Unchanged from original SPEC draft. Features merged, extensions deduplicated and sorted, setup.sh blocks in alphabetical order by technology key. See original Sections 6.5 from the prior draft — the logic is identical.

#### devcontainer.json

Features are merged from all selected technologies plus the base. Extensions are merged, deduplicated, and sorted alphabetically.

```go
func renderDevcontainerJSON(projectName string, techs []*techdef.TechDef) ([]byte, error) {
    // Base features
    features := map[string]interface{}{
        "ghcr.io/devcontainers/features/node:1": map[string]interface{}{},
    }
    // Merge technology features
    for _, tech := range techs {
        for k, v := range tech.Devcontainer.Features {
            features[k] = v
        }
    }
    // Merge and deduplicate extensions, then sort
    extSet := make(map[string]bool)
    for _, tech := range techs {
        for _, ext := range tech.Devcontainer.Extensions {
            extSet[ext] = true
        }
    }
    extensions := make([]string, 0, len(extSet))
    for ext := range extSet {
        extensions = append(extensions, ext)
    }
    sort.Strings(extensions)

    dc := DevcontainerJSON{
        Name:     projectName,
        Build:    map[string]string{"dockerfile": "Dockerfile"},
        Features: features,
        Customizations: map[string]interface{}{
            "vscode": map[string]interface{}{
                "extensions": extensions,
            },
        },
        PostCreateCommand: "bash .devcontainer/setup.sh",
    }
    return json.MarshalIndent(dc, "", "    ")
}
```

#### setup.sh

Base block first, then technology blocks in alphabetical order by technology key:

```go
func renderSetupSh(baseTmpl string, techs []*techdef.TechDef) string {
    var sb strings.Builder
    sb.WriteString(baseTmpl)
    for _, tech := range techs {
        if strings.TrimSpace(tech.Devcontainer.Setup) != "" {
            sb.WriteString("\n# === ")
            sb.WriteString(tech.Name)
            sb.WriteString(" ===\n")
            sb.WriteString(tech.Devcontainer.Setup)
        }
    }
    return sb.String()
}
```

### 6.6 Duplicate Prompt Key Detection

Before presenting technology-driven prompts, the engine checks for duplicate prompt keys across all selected technologies:

```go
func DetectPromptKeyConflicts(techs []*techdef.TechDef) error {
    seen := make(map[string]string) // key → technology name
    for _, tech := range techs {
        for _, p := range tech.Prompts {
            if existing, ok := seen[p.Key]; ok {
                return fmt.Errorf(
                    "prompt key conflict: '%s' is defined by both '%s' and '%s'",
                    p.Key, existing, tech.Name,
                )
            }
            seen[p.Key] = tech.Name
        }
    }
    return nil
}
```

### 6.7 Composite CI Configuration

The CI workflow file is rendered programmatically (not via `text/template`) to avoid fragile YAML-in-YAML templating. The engine composes a three-stage pipeline: lint, test, and build.

**Pipeline structure:**

- **Lint stage:** One job per technology (`lint-{job_name}`), running in parallel. No dependencies.
- **Test stage:** One job per technology (`test-{job_name}`), running in parallel. Each test job depends on **all** lint jobs — the entire lint stage must be green.
- **Build stage:** A single `build` job with placeholder TODO. Depends on **all** test jobs.

Each lint and test job has: checkout → setup steps (if any) → lint or test steps.

```go
func RenderCIConfig(techs []*techdef.TechDef) ([]byte, error) {
    // Collect all technologies that contribute CI
    var ciTechs []*techdef.TechDef
    for _, tech := range techs {
        if tech.CI != nil {
            ciTechs = append(ciTechs, tech)
        }
    }

    // If no technologies contribute CI, fall back to placeholder
    if len(ciTechs) == 0 {
        return renderPlaceholderCI()
    }

    // Build lint job names for the test stage needs
    lintJobNames := make([]string, len(ciTechs))
    for i, tech := range ciTechs {
        lintJobNames[i] = "lint-" + tech.CI.JobName
    }

    // Build test job names for the build stage needs
    testJobNames := make([]string, len(ciTechs))
    for i, tech := range ciTechs {
        testJobNames[i] = "test-" + tech.CI.JobName
    }

    jobs := make(map[string]interface{})

    for _, tech := range ciTechs {
        // Build common setup steps (checkout + tech setup)
        setupSteps := []map[string]interface{}{
            {"uses": "actions/checkout@v4"},
        }
        for _, s := range tech.CI.SetupSteps {
            step := map[string]interface{}{"name": s.Name}
            if s.Uses != "" {
                step["uses"] = s.Uses
                if len(s.With) > 0 {
                    step["with"] = s.With
                }
            }
            if s.Run != "" {
                step["run"] = s.Run
            }
            setupSteps = append(setupSteps, step)
        }

        // Lint job
        lintSteps := make([]map[string]interface{}, len(setupSteps))
        copy(lintSteps, setupSteps)
        for _, s := range tech.CI.LintSteps {
            lintSteps = append(lintSteps, map[string]interface{}{
                "name": s.Name,
                "run":  s.Run,
            })
        }
        jobs["lint-"+tech.CI.JobName] = map[string]interface{}{
            "runs-on": "ubuntu-latest",
            "steps":   lintSteps,
        }

        // Test job — depends on ALL lint jobs
        testSteps := make([]map[string]interface{}, len(setupSteps))
        copy(testSteps, setupSteps)
        for _, s := range tech.CI.TestSteps {
            testSteps = append(testSteps, map[string]interface{}{
                "name": s.Name,
                "run":  s.Run,
            })
        }
        jobs["test-"+tech.CI.JobName] = map[string]interface{}{
            "runs-on": "ubuntu-latest",
            "needs":   lintJobNames,
            "steps":   testSteps,
        }
    }

    // Build job — depends on ALL test jobs
    jobs["build"] = map[string]interface{}{
        "runs-on": "ubuntu-latest",
        "needs":   testJobNames,
        "steps": []map[string]interface{}{
            {"uses": "actions/checkout@v4"},
            {"name": "Build", "run": "echo 'TODO: Add build steps'"},
        },
    }

    workflow := map[string]interface{}{
        "name": "CI",
        "on": map[string]interface{}{
            "push":         map[string]interface{}{"branches": []string{"main"}},
            "pull_request": map[string]interface{}{"branches": []string{"main"}},
        },
        "jobs": jobs,
    }

    return yaml.Marshal(workflow)
}
```

Note: The `yaml.Marshal` output may need post-processing to produce idiomatic GitHub Actions YAML (e.g. `on` key handling, multiline `run` strings with `|`). The implementation should verify the output is valid and readable. An alternative is to use a string builder with controlled formatting — the implementer should choose the approach that produces the cleanest output.

**Example output for Go + Terraform Infrastructure:**

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
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

  lint-terraform:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install Terraform
        run: |
          sudo apt-get update && sudo apt-get install -y unzip
          curl -sL https://releases.hashicorp.com/terraform/1.9.0/terraform_1.9.0_linux_amd64.zip -o tf.zip
          unzip tf.zip && sudo mv terraform /usr/local/bin/
      - name: Install TFLint
        run: |
          curl -s https://raw.githubusercontent.com/terraform-linters/tflint/master/install_linux.sh | bash
          tflint --init || true
      - name: Format check
        run: terraform -chdir=infrastructure fmt -check -recursive
      - name: Lint
        run: tflint --chdir=infrastructure

  test-go:
    runs-on: ubuntu-latest
    needs: [lint-go, lint-terraform]
    steps:
      - uses: actions/checkout@v4
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: stable
      - name: Test
        run: go test ./... -v

  test-terraform:
    runs-on: ubuntu-latest
    needs: [lint-go, lint-terraform]
    steps:
      - uses: actions/checkout@v4
      - name: Install Terraform
        run: |
          sudo apt-get update && sudo apt-get install -y unzip
          curl -sL https://releases.hashicorp.com/terraform/1.9.0/terraform_1.9.0_linux_amd64.zip -o tf.zip
          unzip tf.zip && sudo mv terraform /usr/local/bin/
      - name: Install TFLint
        run: |
          curl -s https://raw.githubusercontent.com/terraform-linters/tflint/master/install_linux.sh | bash
          tflint --init || true
      - name: Validate
        run: |
          terraform -chdir=infrastructure init -backend=false
          terraform -chdir=infrastructure validate

  build:
    runs-on: ubuntu-latest
    needs: [test-go, test-terraform]
    steps:
      - uses: actions/checkout@v4
      - name: Build
        run: echo 'TODO: Add build steps'
```

### 6.8 Universal Makefile

The `Makefile` is added to the universal outputs. It is always generated, always empty (zero bytes):

```go
// In createScaffold, alongside README.md and .gitignore:
if err := os.WriteFile(filepath.Join(targetDir, "Makefile"), []byte{}, 0o644); err != nil {
    return fmt.Errorf("creating Makefile: %w", err)
}
```

### 6.9 Generation Orchestration — Updated

The `createScaffold` function orchestration order:

1. Build template data from config (project name, year, tech prompt responses).
2. Detect path conflicts across all selected technologies (after template resolution).
3. Detect prompt key conflicts across all selected technologies.
4. Create the target directory.
5. Generate universal outputs (README.md, Makefile, composite .gitignore, composite CI config).
6. Generate structure from all selected technologies (with template variable substitution).
7. Generate composite devcontainer (merged features, extensions, setup.sh).
8. Generate conditional outputs (licence, docs, tooling, repo config) based on user selections.

---

## 7. Output Specification

### 7.1 Universal Outputs

Updated from Stage 1:

- `README.md` — `# {{.ProjectName}}` (unchanged)
- `Makefile` — empty file, zero bytes (NEW)
- `.gitignore` — composite from all selected technologies
- `.github/workflows/ci.yml` — composed from technology CI contributions

### 7.2 Complete Output Tree — Go + Terraform Infrastructure, Maximum Selections

When selecting Go + Terraform Infrastructure with all optional selections (MIT licence, CONTRIBUTING.md, both tooling files, all GitHub repo config):

```
my-project/
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
├── cmd/
│   └── app/
│       └── .gitkeep
├── CONTRIBUTING.md
├── infrastructure/
│   ├── .gitkeep
│   ├── main.tf
│   ├── outputs.tf
│   ├── variables.tf
│   └── versions.tf
├── internal/
│   └── .gitkeep
├── LICENSE
├── Makefile
└── README.md
```

### 7.3 Complete Output Tree — Go + Python + Terraform Infrastructure, Minimum Selections

When selecting Go + Python + Terraform Infrastructure with minimum optional selections. Assume `package_name` prompt response is `my_project`:

```
my-project/
├── .devcontainer/
│   ├── devcontainer.json
│   ├── Dockerfile
│   └── setup.sh
├── .github/
│   └── workflows/
│       └── ci.yml
├── .gitignore
├── cmd/
│   └── app/
│       └── .gitkeep
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

### 7.4 Complete Output Tree — Terraform Module (Standalone), Maximum Selections

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
├── examples/
│   └── .gitkeep
├── LICENSE
├── main.tf
├── Makefile
├── modules/
│   └── .gitkeep
├── outputs.tf
├── README.md
├── variables.tf
└── versions.tf
```

### 7.5 Complete Output Tree — Python (Single), Minimum Selections

Assume `package_name` prompt response is `my_app`:

```
my-app/
├── .devcontainer/
│   ├── devcontainer.json
│   ├── Dockerfile
│   └── setup.sh
├── .github/
│   └── workflows/
│       └── ci.yml
├── .gitignore
├── Makefile
├── pyproject.toml
├── README.md
├── src/
│   └── my_app/
│       ├── .gitkeep
│       └── __init__.py
└── tests/
    ├── .gitkeep
    └── __init__.py
```

### 7.6 Complete Output Tree — PowerShell Module (Standalone), Maximum Selections

Unchanged from Stage 1 Section 5.10, except: `Makefile` is now present, and `ci.yml` contains PowerShell-specific lint and test jobs instead of placeholder TODOs.

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
├── Makefile
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

## 8. Testing Strategy

### 8.1 Principles

Unchanged from Stage 1. All development follows strict TDD.

### 8.2 Updating Existing Tests

All Stage 1 tests that reference `cfg.Technology` (singular) must be updated to `cfg.Technologies` (slice). All calls to `scaffold.Generate` must be updated to pass `[]*techdef.TechDef` instead of `*techdef.TechDef`. All calls to `cfg.Validate()` must be updated to pass `techDefs`. Existing expected file trees must include `Makefile`. Existing CI content assertions must be updated to reflect the new composed CI output.

### 8.3 New Unit Tests

#### Standalone field parsing

```go
func TestTechDefStandaloneField(t *testing.T) {
    defs, err := techdef.Load()
    require.NoError(t, err)

    assert.True(t, defs["powershell"].Standalone)
    assert.True(t, defs["terraform-module"].Standalone)
    assert.False(t, defs["go"].Standalone)
    assert.False(t, defs["terraform-infrastructure"].Standalone)
    assert.False(t, defs["python"].Standalone)
}
```

#### Config validation — standalone constraint

```go
func TestConfigRejectsStandaloneWithOtherTechs(t *testing.T) {
    techDefs, _ := techdef.Load()
    cfg := &config.Config{
        ProjectName:  "test",
        Technologies: []string{"powershell", "go"},
    }
    err := cfg.Validate(techDefs)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "standalone")
}

func TestConfigAllowsMultipleComposableTechs(t *testing.T) {
    techDefs, _ := techdef.Load()
    cfg := &config.Config{
        ProjectName:  "test",
        Technologies: []string{"go", "terraform-infrastructure"},
    }
    err := cfg.Validate(techDefs)
    assert.NoError(t, err)
}

func TestConfigAllowsSingleStandaloneTech(t *testing.T) {
    techDefs, _ := techdef.Load()
    cfg := &config.Config{
        ProjectName:  "test",
        Technologies: []string{"terraform-module"},
    }
    err := cfg.Validate(techDefs)
    assert.NoError(t, err)
}

func TestConfigRequiresAtLeastOneTech(t *testing.T) {
    techDefs, _ := techdef.Load()
    cfg := &config.Config{
        ProjectName:  "test",
        Technologies: []string{},
    }
    err := cfg.Validate(techDefs)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "at least one")
}
```

#### Sanitisation

```go
func TestSanitiseForIdentifier(t *testing.T) {
    tests := []struct {
        input    string
        expected string
    }{
        {"my-project", "my_project"},
        {"MyApp", "myapp"},
        {"123-bad-start", "bad_start"},
        {"---", "app"},
        {"hello_world", "hello_world"},
        {"UPPER", "upper"},
        {"a", "a"},
        {"", "app"},
    }
    for _, tc := range tests {
        t.Run(tc.input, func(t *testing.T) {
            assert.Equal(t, tc.expected, prompt.SanitiseForIdentifier(tc.input))
        })
    }
}
```

#### Technology-driven prompt definitions

```go
func TestPythonHasPackageNamePrompt(t *testing.T) {
    defs, err := techdef.Load()
    require.NoError(t, err)

    pyDef := defs["python"]
    require.Len(t, pyDef.Prompts, 1)

    p := pyDef.Prompts[0]
    assert.Equal(t, "package_name", p.Key)
    assert.Equal(t, "text", p.Type)
    assert.Equal(t, "project_name", p.DefaultFrom)
}

func TestGoHasNoPrompts(t *testing.T) {
    defs, err := techdef.Load()
    require.NoError(t, err)
    assert.Empty(t, defs["go"].Prompts)
}
```

#### Prompt validation

```go
func TestPromptValidation_InvalidKey(t *testing.T) {
    def := &techdef.TechDef{
        Name:      "Test",
        Structure: []techdef.StructureEntry{{Path: "src/"}},
        Gitignore: "*.log",
        Prompts: []techdef.PromptDef{
            {Key: "123bad", Title: "Bad:", Type: "text"},
        },
    }
    err := def.Validate("test")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "not a valid identifier")
}

func TestPromptValidation_SelectWithoutOptions(t *testing.T) {
    def := &techdef.TechDef{
        Name:      "Test",
        Structure: []techdef.StructureEntry{{Path: "src/"}},
        Gitignore: "*.log",
        Prompts: []techdef.PromptDef{
            {Key: "choice", Title: "Pick:", Type: "select"},
        },
    }
    err := def.Validate("test")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "options required")
}
```

#### Template resolution in structure entries

```go
func TestResolveTemplatePath(t *testing.T) {
    data := map[string]string{
        "ProjectName":  "my-project",
        "package_name": "my_project",
    }
    resolved, err := scaffold.ResolveTemplate("src/{{.package_name}}/", data)
    require.NoError(t, err)
    assert.Equal(t, "src/my_project/", resolved)
}

func TestResolveTemplateContent(t *testing.T) {
    data := map[string]string{
        "package_name": "my_project",
    }
    resolved, err := scaffold.ResolveTemplate(`"""{{.package_name}} package."""`, data)
    require.NoError(t, err)
    assert.Equal(t, `"""my_project package."""`, resolved)
}
```

#### Path conflict detection

```go
func TestDetectsFilePathConflict(t *testing.T) {
    content := "hello"
    tech1 := &techdef.TechDef{
        Name:      "Alpha",
        Structure: []techdef.StructureEntry{{Path: "src/main.go", Content: &content}},
    }
    tech2 := &techdef.TechDef{
        Name:      "Beta",
        Structure: []techdef.StructureEntry{{Path: "src/main.go", Content: &content}},
    }
    data := map[string]string{"ProjectName": "test"}
    err := scaffold.DetectPathConflicts([]*techdef.TechDef{tech1, tech2}, data)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "path conflict")
    assert.Contains(t, err.Error(), "src/main.go")
}

func TestAllowsSharedDirectoryPaths(t *testing.T) {
    tech1 := &techdef.TechDef{
        Name:      "Alpha",
        Structure: []techdef.StructureEntry{{Path: "src/"}},
    }
    tech2 := &techdef.TechDef{
        Name:      "Beta",
        Structure: []techdef.StructureEntry{{Path: "src/"}},
    }
    data := map[string]string{"ProjectName": "test"}
    err := scaffold.DetectPathConflicts([]*techdef.TechDef{tech1, tech2}, data)
    assert.NoError(t, err)
}
```

#### Duplicate prompt key detection

```go
func TestDetectsPromptKeyConflict(t *testing.T) {
    tech1 := &techdef.TechDef{
        Name:    "Alpha",
        Prompts: []techdef.PromptDef{{Key: "name", Title: "Name:", Type: "text"}},
    }
    tech2 := &techdef.TechDef{
        Name:    "Beta",
        Prompts: []techdef.PromptDef{{Key: "name", Title: "Name:", Type: "text"}},
    }
    err := scaffold.DetectPromptKeyConflicts([]*techdef.TechDef{tech1, tech2})
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "prompt key conflict")
}
```

#### Composite gitignore with multiple technologies

```go
func TestComposeGitignoreMultipleTechs(t *testing.T) {
    techs := []*techdef.TechDef{
        {Name: "Go", Gitignore: "*.exe\nbin/\n"},
        {Name: "Terraform (Infrastructure)", Gitignore: ".terraform/\n*.tfstate\n"},
    }
    result := scaffold.ComposeGitignore(techs)
    assert.Contains(t, result, "# Go")
    assert.Contains(t, result, "# Terraform (Infrastructure)")
    goIdx := strings.Index(result, "# Go")
    tfIdx := strings.Index(result, "# Terraform (Infrastructure)")
    assert.Less(t, goIdx, tfIdx)
}
```

#### CI composition

```go
func TestCIComposition_SingleTech(t *testing.T) {
    techs := []*techdef.TechDef{
        {
            Name: "Go",
            CI: &techdef.CIDef{
                JobName: "go",
                SetupSteps: []techdef.CISetupStep{
                    {Name: "Set up Go", Uses: "actions/setup-go@v5", With: map[string]string{"go-version": "stable"}},
                },
                LintSteps: []techdef.CIStep{{Name: "Lint", Run: "golangci-lint run"}},
                TestSteps: []techdef.CIStep{{Name: "Test", Run: "go test ./..."}},
            },
        },
    }
    result, err := scaffold.RenderCIConfig(techs)
    require.NoError(t, err)
    ciStr := string(result)

    // Three stages present
    assert.Contains(t, ciStr, "lint-go")
    assert.Contains(t, ciStr, "test-go")
    assert.Contains(t, ciStr, "build")

    // Setup and steps present
    assert.Contains(t, ciStr, "actions/setup-go@v5")
    assert.Contains(t, ciStr, "golangci-lint run")
    assert.Contains(t, ciStr, "go test")

    // Build job has TODO
    assert.Contains(t, ciStr, "TODO")
}

func TestCIComposition_MultipleTechs(t *testing.T) {
    techs := []*techdef.TechDef{
        {
            Name: "Go",
            CI: &techdef.CIDef{
                JobName:   "go",
                LintSteps: []techdef.CIStep{{Name: "Lint", Run: "golangci-lint run"}},
                TestSteps: []techdef.CIStep{{Name: "Test", Run: "go test ./..."}},
            },
        },
        {
            Name: "Python",
            CI: &techdef.CIDef{
                JobName:   "python",
                LintSteps: []techdef.CIStep{{Name: "Lint", Run: "ruff check ."}},
                TestSteps: []techdef.CIStep{{Name: "Test", Run: "pytest"}},
            },
        },
    }
    result, err := scaffold.RenderCIConfig(techs)
    require.NoError(t, err)
    ciStr := string(result)

    // Lint jobs for both techs
    assert.Contains(t, ciStr, "lint-go")
    assert.Contains(t, ciStr, "lint-python")

    // Test jobs for both techs
    assert.Contains(t, ciStr, "test-go")
    assert.Contains(t, ciStr, "test-python")

    // Test jobs depend on ALL lint jobs
    assert.Contains(t, ciStr, "lint-go")
    assert.Contains(t, ciStr, "lint-python")

    // Single build job depends on ALL test jobs
    assert.Contains(t, ciStr, "build")
}

func TestCIComposition_SetupStepsInBothJobs(t *testing.T) {
    techs := []*techdef.TechDef{
        {
            Name: "PowerShell Module",
            CI: &techdef.CIDef{
                JobName: "powershell",
                SetupSteps: []techdef.CISetupStep{
                    {Name: "Install PowerShell", Run: "sudo apt-get install -y powershell"},
                },
                LintSteps: []techdef.CIStep{{Name: "Lint", Run: "Invoke-ScriptAnalyzer"}},
                TestSteps: []techdef.CIStep{{Name: "Test", Run: "Invoke-Pester"}},
            },
        },
    }
    result, err := scaffold.RenderCIConfig(techs)
    require.NoError(t, err)
    ciStr := string(result)

    // Setup steps appear in both lint and test jobs
    assert.Contains(t, ciStr, "lint-powershell")
    assert.Contains(t, ciStr, "test-powershell")
    // PowerShell install appears (should be in both jobs)
    assert.Contains(t, ciStr, "Install PowerShell")
}

func TestCIComposition_FallbackPlaceholder(t *testing.T) {
    techs := []*techdef.TechDef{
        {Name: "NoCi"},
    }
    result, err := scaffold.RenderCIConfig(techs)
    require.NoError(t, err)
    assert.Contains(t, string(result), "TODO")
}
```

#### CI validation

```go
func TestCIValidation_MissingJobName(t *testing.T) {
    def := &techdef.TechDef{
        Name:      "Test",
        Structure: []techdef.StructureEntry{{Path: "src/"}},
        Gitignore: "*.log",
        CI: &techdef.CIDef{
            JobName:   "",
            LintSteps: []techdef.CIStep{{Name: "Lint", Run: "echo"}},
            TestSteps: []techdef.CIStep{{Name: "Test", Run: "echo"}},
        },
    }
    err := def.Validate("test")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "job_name")
}

func TestCIValidation_EmptyLintSteps(t *testing.T) {
    def := &techdef.TechDef{
        Name:      "Test",
        Structure: []techdef.StructureEntry{{Path: "src/"}},
        Gitignore: "*.log",
        CI: &techdef.CIDef{
            JobName:   "test",
            LintSteps: []techdef.CIStep{},
            TestSteps: []techdef.CIStep{{Name: "Test", Run: "echo"}},
        },
    }
    err := def.Validate("test")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "lint_steps")
}

func TestCIValidation_EmptyTestSteps(t *testing.T) {
    def := &techdef.TechDef{
        Name:      "Test",
        Structure: []techdef.StructureEntry{{Path: "src/"}},
        Gitignore: "*.log",
        CI: &techdef.CIDef{
            JobName:   "test",
            LintSteps: []techdef.CIStep{{Name: "Lint", Run: "echo"}},
            TestSteps: []techdef.CIStep{},
        },
    }
    err := def.Validate("test")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "test_steps")
}

func TestCIValidation_SetupStepMustHaveUsesOrRun(t *testing.T) {
    def := &techdef.TechDef{
        Name:      "Test",
        Structure: []techdef.StructureEntry{{Path: "src/"}},
        Gitignore: "*.log",
        CI: &techdef.CIDef{
            JobName:    "test",
            SetupSteps: []techdef.CISetupStep{{Name: "Bad"}},
            LintSteps:  []techdef.CIStep{{Name: "Lint", Run: "echo"}},
            TestSteps:  []techdef.CIStep{{Name: "Test", Run: "echo"}},
        },
    }
    err := def.Validate("test")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "uses")
}
```

#### New technology definition loading

```go
func TestLoadAllTechDefs(t *testing.T) {
    defs, err := techdef.Load()
    require.NoError(t, err)

    expectedKeys := []string{"go", "powershell", "python", "terraform-infrastructure", "terraform-module"}
    actualKeys := make([]string, 0, len(defs))
    for key := range defs {
        actualKeys = append(actualKeys, key)
    }
    sort.Strings(actualKeys)
    assert.Equal(t, expectedKeys, actualKeys)
}
```

#### Makefile in universal outputs

```go
func TestMakefileAlwaysGenerated(t *testing.T) {
    baseDir := t.TempDir()
    defs, _ := techdef.Load()

    cfg := &config.Config{
        ProjectName:  "test-proj",
        Provider:     "github",
        Technologies: []string{"go"},
        Licence:      "none",
        Docs:         []string{},
        Tooling:      []string{},
        RepoConfig:   []string{},
        Confirmed:    true,
    }

    techs := []*techdef.TechDef{defs["go"]}
    err := scaffold.Generate(cfg, techs, baseDir)
    require.NoError(t, err)

    makefile := filepath.Join(baseDir, "test-proj", "Makefile")
    assert.FileExists(t, makefile)
    content, _ := os.ReadFile(makefile)
    assert.Empty(t, content, "Makefile should be empty")
}
```

### 8.4 Acceptance Tests

#### Acceptance test: Go (single technology, minimum selections)

```go
func TestGoMinimumSelections(t *testing.T) {
    baseDir := t.TempDir()
    defs, err := techdef.Load()
    require.NoError(t, err)

    cfg := &config.Config{
        ProjectName:  "go-project",
        Provider:     "github",
        Technologies: []string{"go"},
        Licence:      "none",
        Docs:         []string{},
        Tooling:      []string{},
        RepoConfig:   []string{},
        Confirmed:    true,
    }

    techs := []*techdef.TechDef{defs["go"]}
    err = scaffold.Generate(cfg, techs, baseDir)
    require.NoError(t, err)

    root := filepath.Join(baseDir, "go-project")

    expectedFiles := []string{
        ".devcontainer/devcontainer.json",
        ".devcontainer/Dockerfile",
        ".devcontainer/setup.sh",
        ".github/workflows/ci.yml",
        ".gitignore",
        "cmd/app/.gitkeep",
        "internal/.gitkeep",
        "Makefile",
        "README.md",
    }

    actualFiles := collectFiles(t, root)
    assert.ElementsMatch(t, expectedFiles, actualFiles)

    // Verify gitignore
    gitignore, _ := os.ReadFile(filepath.Join(root, ".gitignore"))
    assert.Contains(t, string(gitignore), "# Go")
    assert.Contains(t, string(gitignore), "*.exe")

    // Verify CI has three-stage pipeline with Go jobs
    ci, _ := os.ReadFile(filepath.Join(root, ".github/workflows/ci.yml"))
    assert.Contains(t, string(ci), "lint-go")
    assert.Contains(t, string(ci), "test-go")
    assert.Contains(t, string(ci), "build")
    assert.Contains(t, string(ci), "actions/setup-go")
    assert.Contains(t, string(ci), "go test")
    assert.Contains(t, string(ci), "golangci-lint")

    // Verify devcontainer
    dcJSON, _ := os.ReadFile(filepath.Join(root, ".devcontainer/devcontainer.json"))
    assert.Contains(t, string(dcJSON), "ghcr.io/devcontainers/features/go:1")

    setupSh, _ := os.ReadFile(filepath.Join(root, ".devcontainer/setup.sh"))
    assert.Contains(t, string(setupSh), "# === Go ===")
    assert.Contains(t, string(setupSh), "golangci-lint")
}
```

#### Acceptance test: Python (single technology, with package name prompt)

```go
func TestPythonMinimumSelections(t *testing.T) {
    baseDir := t.TempDir()
    defs, err := techdef.Load()
    require.NoError(t, err)

    cfg := &config.Config{
        ProjectName:         "my-app",
        Provider:            "github",
        Technologies:        []string{"python"},
        TechPromptResponses: map[string]string{"package_name": "my_app"},
        Licence:             "none",
        Docs:                []string{},
        Tooling:             []string{},
        RepoConfig:          []string{},
        Confirmed:           true,
    }

    techs := []*techdef.TechDef{defs["python"]}
    err = scaffold.Generate(cfg, techs, baseDir)
    require.NoError(t, err)

    root := filepath.Join(baseDir, "my-app")

    expectedFiles := []string{
        ".devcontainer/devcontainer.json",
        ".devcontainer/Dockerfile",
        ".devcontainer/setup.sh",
        ".github/workflows/ci.yml",
        ".gitignore",
        "Makefile",
        "pyproject.toml",
        "README.md",
        "src/my_app/.gitkeep",
        "src/my_app/__init__.py",
        "tests/.gitkeep",
        "tests/__init__.py",
    }

    actualFiles := collectFiles(t, root)
    assert.ElementsMatch(t, expectedFiles, actualFiles)

    // Verify template substitution in pyproject.toml
    pyproject, _ := os.ReadFile(filepath.Join(root, "pyproject.toml"))
    assert.Contains(t, string(pyproject), `name = "my_app"`)
    assert.Contains(t, string(pyproject), "pytest")
    assert.Contains(t, string(pyproject), "ruff")

    // Verify template substitution in __init__.py
    initPy, _ := os.ReadFile(filepath.Join(root, "src/my_app/__init__.py"))
    assert.Contains(t, string(initPy), "my_app")

    // Verify CI has three-stage pipeline with Python jobs
    ci, _ := os.ReadFile(filepath.Join(root, ".github/workflows/ci.yml"))
    assert.Contains(t, string(ci), "lint-python")
    assert.Contains(t, string(ci), "test-python")
    assert.Contains(t, string(ci), "build")
    assert.Contains(t, string(ci), "actions/setup-python")
    assert.Contains(t, string(ci), "ruff")
    assert.Contains(t, string(ci), "pytest")
}
```

#### Acceptance test: Terraform Module (standalone, maximum selections)

```go
func TestTerraformModuleMaximumSelections(t *testing.T) {
    baseDir := t.TempDir()
    defs, err := techdef.Load()
    require.NoError(t, err)

    cfg := &config.Config{
        ProjectName:  "my-tf-module",
        Provider:     "github",
        Technologies: []string{"terraform-module"},
        Licence:      "mit",
        Docs:         []string{"contributing"},
        Tooling:      []string{"editorconfig", "gitattributes"},
        RepoConfig:   []string{"issue_templates", "pr_template", "dependabot"},
        Confirmed:    true,
    }

    techs := []*techdef.TechDef{defs["terraform-module"]}
    err = scaffold.Generate(cfg, techs, baseDir)
    require.NoError(t, err)

    root := filepath.Join(baseDir, "my-tf-module")

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
        "examples/.gitkeep",
        "LICENSE",
        "main.tf",
        "Makefile",
        "modules/.gitkeep",
        "outputs.tf",
        "README.md",
        "variables.tf",
        "versions.tf",
    }

    actualFiles := collectFiles(t, root)
    assert.ElementsMatch(t, expectedFiles, actualFiles)

    mainTf, _ := os.ReadFile(filepath.Join(root, "main.tf"))
    assert.Contains(t, string(mainTf), "Main Terraform configuration")

    ci, _ := os.ReadFile(filepath.Join(root, ".github/workflows/ci.yml"))
    assert.Contains(t, string(ci), "lint-terraform")
    assert.Contains(t, string(ci), "test-terraform")
    assert.Contains(t, string(ci), "build")
    assert.Contains(t, string(ci), "terraform fmt")
    assert.Contains(t, string(ci), "terraform validate")
}
```

#### Acceptance test: Go + Terraform Infrastructure (multi-tech)

```go
func TestGoAndTerraformInfrastructure(t *testing.T) {
    baseDir := t.TempDir()
    defs, err := techdef.Load()
    require.NoError(t, err)

    cfg := &config.Config{
        ProjectName:  "my-app",
        Provider:     "github",
        Technologies: []string{"go", "terraform-infrastructure"},
        Licence:      "none",
        Docs:         []string{},
        Tooling:      []string{},
        RepoConfig:   []string{},
        Confirmed:    true,
    }

    techs := []*techdef.TechDef{defs["go"], defs["terraform-infrastructure"]}
    sort.Slice(techs, func(i, j int) bool { return techs[i].Name < techs[j].Name })

    err = scaffold.Generate(cfg, techs, baseDir)
    require.NoError(t, err)

    root := filepath.Join(baseDir, "my-app")

    expectedFiles := []string{
        ".devcontainer/devcontainer.json",
        ".devcontainer/Dockerfile",
        ".devcontainer/setup.sh",
        ".github/workflows/ci.yml",
        ".gitignore",
        "cmd/app/.gitkeep",
        "infrastructure/.gitkeep",
        "infrastructure/main.tf",
        "infrastructure/outputs.tf",
        "infrastructure/variables.tf",
        "infrastructure/versions.tf",
        "internal/.gitkeep",
        "Makefile",
        "README.md",
    }

    actualFiles := collectFiles(t, root)
    assert.ElementsMatch(t, expectedFiles, actualFiles)

    // Verify composite gitignore
    gitignore, _ := os.ReadFile(filepath.Join(root, ".gitignore"))
    assert.Contains(t, string(gitignore), "# Go")
    assert.Contains(t, string(gitignore), "# Terraform (Infrastructure)")

    // Verify CI has three-stage pipeline with both tech jobs
    ci, _ := os.ReadFile(filepath.Join(root, ".github/workflows/ci.yml"))
    assert.Contains(t, string(ci), "lint-go")
    assert.Contains(t, string(ci), "lint-terraform")
    assert.Contains(t, string(ci), "test-go")
    assert.Contains(t, string(ci), "test-terraform")
    assert.Contains(t, string(ci), "build")
    assert.Contains(t, string(ci), "go test")
    assert.Contains(t, string(ci), "terraform")

    // Verify composite devcontainer
    dcJSON, _ := os.ReadFile(filepath.Join(root, ".devcontainer/devcontainer.json"))
    assert.Contains(t, string(dcJSON), "ghcr.io/devcontainers/features/go:1")
    assert.Contains(t, string(dcJSON), "ghcr.io/devcontainers-contrib/features/terraform-asdf:1")

    setupSh, _ := os.ReadFile(filepath.Join(root, ".devcontainer/setup.sh"))
    assert.Contains(t, string(setupSh), "# === Go ===")
    assert.Contains(t, string(setupSh), "# === Terraform (Infrastructure) ===")
}
```

#### Acceptance test: Go + Python + Terraform Infrastructure (three composable techs)

```go
func TestThreeComposableTechs(t *testing.T) {
    baseDir := t.TempDir()
    defs, err := techdef.Load()
    require.NoError(t, err)

    cfg := &config.Config{
        ProjectName:         "polyglot",
        Provider:            "github",
        Technologies:        []string{"go", "python", "terraform-infrastructure"},
        TechPromptResponses: map[string]string{"package_name": "polyglot"},
        Licence:             "mit",
        Docs:                []string{"contributing"},
        Tooling:             []string{"editorconfig", "gitattributes"},
        RepoConfig:          []string{"issue_templates", "pr_template", "dependabot"},
        Confirmed:           true,
    }

    keys := []string{"go", "python", "terraform-infrastructure"}
    sort.Strings(keys)
    techs := make([]*techdef.TechDef, len(keys))
    for i, k := range keys {
        techs[i] = defs[k]
    }

    err = scaffold.Generate(cfg, techs, baseDir)
    require.NoError(t, err)

    root := filepath.Join(baseDir, "polyglot")

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
        "cmd/app/.gitkeep",
        "CONTRIBUTING.md",
        "infrastructure/.gitkeep",
        "infrastructure/main.tf",
        "infrastructure/outputs.tf",
        "infrastructure/variables.tf",
        "infrastructure/versions.tf",
        "internal/.gitkeep",
        "LICENSE",
        "Makefile",
        "pyproject.toml",
        "README.md",
        "src/polyglot/.gitkeep",
        "src/polyglot/__init__.py",
        "tests/.gitkeep",
        "tests/__init__.py",
    }

    actualFiles := collectFiles(t, root)
    assert.ElementsMatch(t, expectedFiles, actualFiles)

    // Verify gitignore ordering
    gitignore, _ := os.ReadFile(filepath.Join(root, ".gitignore"))
    goIdx := strings.Index(string(gitignore), "# Go")
    pyIdx := strings.Index(string(gitignore), "# Python")
    tfIdx := strings.Index(string(gitignore), "# Terraform (Infrastructure)")
    assert.Less(t, goIdx, pyIdx)
    assert.Less(t, pyIdx, tfIdx)

    // Verify setup.sh ordering
    setupSh, _ := os.ReadFile(filepath.Join(root, ".devcontainer/setup.sh"))
    goIdx = strings.Index(string(setupSh), "# === Go ===")
    pyIdx = strings.Index(string(setupSh), "# === Python ===")
    tfIdx = strings.Index(string(setupSh), "# === Terraform (Infrastructure) ===")
    assert.Less(t, goIdx, pyIdx)
    assert.Less(t, pyIdx, tfIdx)

    // Verify CI has three-stage pipeline with all three tech jobs
    ci, _ := os.ReadFile(filepath.Join(root, ".github/workflows/ci.yml"))
    assert.Contains(t, string(ci), "lint-go")
    assert.Contains(t, string(ci), "lint-python")
    assert.Contains(t, string(ci), "lint-terraform")
    assert.Contains(t, string(ci), "test-go")
    assert.Contains(t, string(ci), "test-python")
    assert.Contains(t, string(ci), "test-terraform")
    assert.Contains(t, string(ci), "build")
    assert.Contains(t, string(ci), "go test")
    assert.Contains(t, string(ci), "ruff")
    assert.Contains(t, string(ci), "terraform")
}
```

#### Acceptance test: standalone constraint rejection

```go
func TestStandaloneConstraintRejected(t *testing.T) {
    defs, err := techdef.Load()
    require.NoError(t, err)

    cfg := &config.Config{
        ProjectName:  "bad-combo",
        Provider:     "github",
        Technologies: []string{"terraform-module", "go"},
        Licence:      "none",
        Confirmed:    true,
    }

    err = cfg.Validate(defs)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "standalone")
}
```

### 8.5 Running Tests

Unchanged from Stage 1:

```bash
# Unit tests only
go test ./internal/...

# Acceptance tests only
go test ./tests/acceptance/

# All tests
go test ./...
```

---

## 9. Implementation Order

This section defines the order in which features should be implemented, following TDD. Each step begins with writing the test, then the production code.

### Step 1: Schema and config changes

1. Add `standalone`, `prompts`, and `ci` fields to `TechDef` struct.
2. Add `PromptDef`, `OptionDef`, `CIDef`, `CISetup`, `CIStep` types.
3. Write tests for new field parsing.
4. Write tests for new validation rules (prompt key format, select without options, CI field validation).
5. Change `Config.Technology` to `Config.Technologies`. Add `TechPromptResponses`.
6. Update `Config.Validate()` to accept `techDefs` parameter and enforce standalone constraint.
7. Write tests for all config validation scenarios.
8. Update all existing tests to use new struct shapes.

### Step 2: New technology definitions

1. Create `technologies/go.yaml`.
2. Create `technologies/terraform-module.yaml`.
3. Create `technologies/terraform-infrastructure.yaml`.
4. Create `technologies/python.yaml`.
5. Update `technologies/powershell.yaml` to add `standalone: true` and `ci` field.
6. Write tests verifying each definition loads correctly, passes validation, and has expected fields.

### Step 3: Template resolution in structure entries

1. Write tests for `SanitiseForIdentifier`.
2. Write tests for `ResolveTemplate` — paths and content with variables.
3. Implement template resolution.
4. Integrate into structure creation — resolve paths and content before file creation.

### Step 4: Path and prompt conflict detection

1. Write tests for `DetectPathConflicts` — file conflicts, directory overlaps allowed, with template resolution.
2. Write tests for `DetectPromptKeyConflicts`.
3. Implement both functions.
4. Integrate into `Generate` — call before any file creation.

### Step 5: Multi-technology composition

1. Write tests for `ComposeGitignore` with multiple technologies.
2. Write tests for devcontainer merging (features, extensions, setup.sh).
3. Write tests for `RenderCIConfig` — single tech, multi tech, no setup, fallback.
4. Update `scaffold.Generate` to accept `[]*techdef.TechDef` and iterate over all.
5. Add universal Makefile generation.
6. Implement all composition logic.

### Step 6: Prompt update

1. Implement two-phase form: phase 1 (name, provider, techs), phase 2 (tech prompts, licence, docs, tooling, repo config, confirm).
2. Implement `buildTechPromptGroups` — dynamic prompt generation from selected technologies.
3. Add standalone validation to the technology multi-select prompt.
4. Update `main.go` to resolve technologies and sort by key.
5. Manual testing of the interactive flow, including standalone enforcement and Python package name prompt.

### Step 7: Acceptance tests

1. Write acceptance test for Go (single tech, minimum selections).
2. Write acceptance test for Python (single tech, with package name prompt).
3. Write acceptance test for Terraform Module (standalone, maximum selections).
4. Write acceptance test for Go + Terraform Infrastructure (two composable techs).
5. Write acceptance test for Go + Python + Terraform Infrastructure (three composable techs).
6. Write acceptance test for standalone constraint rejection.
7. Update existing Stage 1 acceptance tests to use new function signatures and include Makefile + CI changes.
8. Verify all acceptance tests pass.

### Step 8: CI verification

1. Push changes and verify CI passes (lint, unit tests, acceptance tests).
2. Fix any linting issues introduced by new code.

---

## 10. Acceptance Criteria

Stage 2 is complete when all of the following are true:

1. The technology prompt is a multi-select, dynamically populated from YAML definitions.
2. All five technology definitions (PowerShell Module, Go, Terraform Module, Terraform Infrastructure, Python) load, parse, and validate correctly.
3. Standalone technologies (PowerShell Module, Terraform Module) cannot be selected alongside other technologies. The prompt prevents it and config validation enforces it.
4. Selecting multiple composable technologies produces a merged scaffold with no conflicts.
5. The `.gitignore` is composed from all selected technologies with sections in alphabetical order by technology name.
6. The `devcontainer.json` features, extensions, and `setup.sh` are correctly merged from all selected technologies.
7. The CI config is composed as a three-stage pipeline (lint → test → build) with per-technology jobs in the lint and test stages. All test jobs depend on all lint jobs passing. The build job depends on all test jobs passing. Setup steps are runner-agnostic.
8. Path conflict detection catches file collisions across technologies and reports a clear error.
9. Technology-driven prompts (e.g. Python package name) are presented after technology selection, with defaults derived from the project name and appropriate sanitisation.
10. Template variables in structure entry paths and content are correctly resolved from prompt responses.
11. An empty `Makefile` is generated for every scaffold.
12. Each new technology in isolation produces the correct directory structure and file contents as specified in Section 4.
13. Multi-technology combinations produce the correct merged output trees as specified in Section 7.
14. All Stage 1 acceptance tests continue to pass (updated for new function signatures, Makefile, and CI changes).
15. All new unit tests pass.
16. All new acceptance tests pass.
17. The project's own CI pipeline passes lint, unit tests, and acceptance tests.
18. Adding a hypothetical new technology still requires only a new YAML file in `technologies/` and corresponding acceptance tests — no changes to the scaffold engine or prompt code.