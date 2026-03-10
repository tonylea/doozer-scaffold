package acceptance_test

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tonylea/doozer-scaffold/internal/config"
	"github.com/tonylea/doozer-scaffold/internal/scaffold"
	"github.com/tonylea/doozer-scaffold/internal/techdef"
)

func loadPowerShellDef(t *testing.T) *techdef.TechDef {
	t.Helper()
	defs, err := techdef.Load()
	require.NoError(t, err)
	require.Contains(t, defs, "powershell")
	return defs["powershell"]
}

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

func TestMaximumSelections(t *testing.T) {
	baseDir := t.TempDir()
	tech := loadPowerShellDef(t)

	cfg := &config.Config{
		ProjectName:  "my-module",
		Provider:     "github",
		Technologies: []string{"powershell"},
		Licence:      "mit",
		Docs:         []string{"contributing"},
		Tooling:      []string{"editorconfig", "gitattributes"},
		RepoConfig:   []string{"issue_templates", "pr_template", "dependabot"},
		Confirmed:    true,
	}

	err := scaffold.Generate(cfg, []*techdef.TechDef{tech}, baseDir)
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
		"Makefile",
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

	info, err := os.Stat(filepath.Join(root, ".devcontainer/setup.sh"))
	require.NoError(t, err)
	assert.True(t, info.Mode().Perm()&0o111 != 0, "setup.sh must be executable")

	// CI should have PowerShell jobs
	ci, err := os.ReadFile(filepath.Join(root, ".github/workflows/ci.yml"))
	require.NoError(t, err)
	assert.Contains(t, string(ci), "lint-powershell")
	assert.Contains(t, string(ci), "test-powershell")
	assert.Contains(t, string(ci), "build")
}

func TestMinimumSelections(t *testing.T) {
	baseDir := t.TempDir()
	tech := loadPowerShellDef(t)

	cfg := &config.Config{
		ProjectName:  "bare-project",
		Provider:     "github",
		Technologies: []string{"powershell"},
		Licence:      "none",
		Docs:         []string{},
		Tooling:      []string{},
		RepoConfig:   []string{},
		Confirmed:    true,
	}

	err := scaffold.Generate(cfg, []*techdef.TechDef{tech}, baseDir)
	require.NoError(t, err)

	root := filepath.Join(baseDir, "bare-project")

	expectedFiles := []string{
		".devcontainer/devcontainer.json",
		".devcontainer/Dockerfile",
		".devcontainer/setup.sh",
		".github/workflows/ci.yml",
		".gitignore",
		"Makefile",
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

	assert.NoFileExists(t, filepath.Join(root, "LICENSE"))
	assert.NoFileExists(t, filepath.Join(root, "CONTRIBUTING.md"))
	assert.NoFileExists(t, filepath.Join(root, ".editorconfig"))
	assert.NoFileExists(t, filepath.Join(root, ".gitattributes"))

	readme, err := os.ReadFile(filepath.Join(root, "README.md"))
	require.NoError(t, err)
	assert.Equal(t, "# bare-project\n", string(readme))
}

func TestConfirmationDeclined(t *testing.T) {
	baseDir := t.TempDir()

	cfg := &config.Config{
		ProjectName:  "declined-project",
		Provider:     "github",
		Technologies: []string{"powershell"},
		Licence:      "mit",
		Confirmed:    false,
	}

	assert.False(t, cfg.Confirmed)
	assert.NoDirExists(t, filepath.Join(baseDir, "declined-project"))
}

// --- Stage 2 acceptance tests ---

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

	gitignore, _ := os.ReadFile(filepath.Join(root, ".gitignore"))
	assert.Contains(t, string(gitignore), "# Go")
	assert.Contains(t, string(gitignore), "*.exe")

	ci, _ := os.ReadFile(filepath.Join(root, ".github/workflows/ci.yml"))
	assert.Contains(t, string(ci), "lint-go")
	assert.Contains(t, string(ci), "test-go")
	assert.Contains(t, string(ci), "build")
	assert.Contains(t, string(ci), "actions/setup-go")
	assert.Contains(t, string(ci), "go test")
	assert.Contains(t, string(ci), "golangci-lint")

	dcJSON, _ := os.ReadFile(filepath.Join(root, ".devcontainer/devcontainer.json"))
	assert.Contains(t, string(dcJSON), "ghcr.io/devcontainers/features/go:1")

	setupSh, _ := os.ReadFile(filepath.Join(root, ".devcontainer/setup.sh"))
	assert.Contains(t, string(setupSh), "# === Go ===")
	assert.Contains(t, string(setupSh), "golangci-lint")
}

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

	pyproject, _ := os.ReadFile(filepath.Join(root, "pyproject.toml"))
	assert.Contains(t, string(pyproject), `name = "my_app"`)
	assert.Contains(t, string(pyproject), "pytest")
	assert.Contains(t, string(pyproject), "ruff")

	initPy, _ := os.ReadFile(filepath.Join(root, "src/my_app/__init__.py"))
	assert.Contains(t, string(initPy), "my_app")

	ci, _ := os.ReadFile(filepath.Join(root, ".github/workflows/ci.yml"))
	assert.Contains(t, string(ci), "lint-python")
	assert.Contains(t, string(ci), "test-python")
	assert.Contains(t, string(ci), "build")
	assert.Contains(t, string(ci), "actions/setup-python")
	assert.Contains(t, string(ci), "ruff")
	assert.Contains(t, string(ci), "pytest")
}

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

	gitignore, _ := os.ReadFile(filepath.Join(root, ".gitignore"))
	assert.Contains(t, string(gitignore), "# Go")
	assert.Contains(t, string(gitignore), "# Terraform (Infrastructure)")

	ci, _ := os.ReadFile(filepath.Join(root, ".github/workflows/ci.yml"))
	assert.Contains(t, string(ci), "lint-go")
	assert.Contains(t, string(ci), "lint-terraform")
	assert.Contains(t, string(ci), "test-go")
	assert.Contains(t, string(ci), "test-terraform")
	assert.Contains(t, string(ci), "build")
	assert.Contains(t, string(ci), "go test")
	assert.Contains(t, string(ci), "terraform")

	dcJSON, _ := os.ReadFile(filepath.Join(root, ".devcontainer/devcontainer.json"))
	assert.Contains(t, string(dcJSON), "ghcr.io/devcontainers/features/go:1")
	assert.Contains(t, string(dcJSON), "ghcr.io/devcontainers-contrib/features/terraform-asdf:1")

	setupSh, _ := os.ReadFile(filepath.Join(root, ".devcontainer/setup.sh"))
	assert.Contains(t, string(setupSh), "# === Go ===")
	assert.Contains(t, string(setupSh), "# === Terraform (Infrastructure) ===")
}

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

	gitignore, _ := os.ReadFile(filepath.Join(root, ".gitignore"))
	goIdx := strings.Index(string(gitignore), "# Go")
	pyIdx := strings.Index(string(gitignore), "# Python")
	tfIdx := strings.Index(string(gitignore), "# Terraform (Infrastructure)")
	assert.Less(t, goIdx, pyIdx)
	assert.Less(t, pyIdx, tfIdx)

	setupSh, _ := os.ReadFile(filepath.Join(root, ".devcontainer/setup.sh"))
	goIdx = strings.Index(string(setupSh), "# === Go ===")
	pyIdx = strings.Index(string(setupSh), "# === Python ===")
	tfIdx = strings.Index(string(setupSh), "# === Terraform (Infrastructure) ===")
	assert.Less(t, goIdx, pyIdx)
	assert.Less(t, pyIdx, tfIdx)

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
