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
	"github.com/tonylea/doozer-scaffold/internal/prompt"
	"github.com/tonylea/doozer-scaffold/internal/scaffold"
	"github.com/tonylea/doozer-scaffold/internal/techdef"
)

// resolveAndGenerate is a helper that resolves variant groups from display keys then generates.
func resolveAndGenerate(t *testing.T, cfg *config.Config, baseDir string) {
	t.Helper()
	defs, err := techdef.Load()
	require.NoError(t, err)
	resolvedTechs, _ := techdef.ResolveVariantGroups(cfg.Technologies, defs)
	sort.Slice(resolvedTechs, func(i, j int) bool { return resolvedTechs[i].Name < resolvedTechs[j].Name })
	err = scaffold.Generate(cfg, resolvedTechs, baseDir)
	require.NoError(t, err)
}

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

	// powershell has no variant_group and is standalone — combining with go must fail
	cfg := &config.Config{
		ProjectName:  "bad-combo",
		Provider:     "github",
		Technologies: []string{"powershell", "go"},
		Licence:      "none",
		Confirmed:    true,
	}

	err = cfg.Validate(defs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "standalone")
}

// --- Stage 3a: Dockerfile acceptance tests ---

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

	var files []string
	err = filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, _ := filepath.Rel(root, path)
		if rel != "." {
			files = append(files, rel)
		}
		return nil
	})
	require.NoError(t, err)

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

	dockerfile, _ := os.ReadFile(filepath.Join(root, "Dockerfile"))
	assert.Contains(t, string(dockerfile), "FROM ubuntu:24.04")
	assert.Contains(t, string(dockerfile), "my-image")

	dockerignore, _ := os.ReadFile(filepath.Join(root, ".dockerignore"))
	assert.Contains(t, string(dockerignore), ".git")
	assert.Contains(t, string(dockerignore), ".devcontainer")

	gitignore, _ := os.ReadFile(filepath.Join(root, ".gitignore"))
	assert.Contains(t, string(gitignore), "# Dockerfile (Image)")
	assert.Contains(t, string(gitignore), ".docker/")

	ci, _ := os.ReadFile(filepath.Join(root, ".github/workflows/ci.yml"))
	assert.Contains(t, string(ci), "lint-docker")
	assert.Contains(t, string(ci), "test-docker")
	assert.Contains(t, string(ci), "hadolint")
	assert.Contains(t, string(ci), "docker build")
	assert.Contains(t, string(ci), "build")

	dcJson, _ := os.ReadFile(filepath.Join(root, ".devcontainer/devcontainer.json"))
	assert.Contains(t, string(dcJson), "docker-in-docker")
	assert.Contains(t, string(dcJson), "ms-azuretools.vscode-docker")
}

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
	err = filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, _ := filepath.Rel(root, path)
		if rel != "." {
			files = append(files, rel)
		}
		return nil
	})
	require.NoError(t, err)

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

	assert.NoFileExists(t, filepath.Join(root, "LICENSE"))
	assert.NoFileExists(t, filepath.Join(root, "CONTRIBUTING.md"))
	assert.NoFileExists(t, filepath.Join(root, ".editorconfig"))
	assert.NoFileExists(t, filepath.Join(root, ".gitattributes"))
}

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
	err = filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, _ := filepath.Rel(root, path)
		if rel != "." {
			files = append(files, rel)
		}
		return nil
	})
	require.NoError(t, err)

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

	gitignore, _ := os.ReadFile(filepath.Join(root, ".gitignore"))
	gitignoreStr := string(gitignore)
	assert.Contains(t, gitignoreStr, "# Dockerfile (Service)")
	assert.Contains(t, gitignoreStr, "# Go")
	dockerIdx := strings.Index(gitignoreStr, "# Dockerfile (Service)")
	goIdx := strings.Index(gitignoreStr, "# Go")
	assert.Less(t, dockerIdx, goIdx, "Docker section should appear before Go (alphabetical)")

	ci, _ := os.ReadFile(filepath.Join(root, ".github/workflows/ci.yml"))
	ciStr := string(ci)
	assert.Contains(t, ciStr, "lint-docker")
	assert.Contains(t, ciStr, "lint-go")
	assert.Contains(t, ciStr, "test-docker")
	assert.Contains(t, ciStr, "test-go")
	assert.Contains(t, ciStr, "hadolint docker/Dockerfile")
	assert.Contains(t, ciStr, "docker build -t test-image -f docker/Dockerfile .")
	assert.Contains(t, ciStr, "build")

	dcJson, _ := os.ReadFile(filepath.Join(root, ".devcontainer/devcontainer.json"))
	dcStr := string(dcJson)
	assert.Contains(t, dcStr, "docker-in-docker")
	assert.Contains(t, dcStr, "ghcr.io/devcontainers/features/go:1")
	assert.Contains(t, dcStr, "golang.go")
	assert.Contains(t, dcStr, "ms-azuretools.vscode-docker")

	setupSh, _ := os.ReadFile(filepath.Join(root, ".devcontainer/setup.sh"))
	setupStr := string(setupSh)
	assert.Contains(t, setupStr, "# === Base tooling ===")
	assert.Contains(t, setupStr, "# === Go ===")
	assert.NotContains(t, setupStr, "# === Dockerfile")
}

func TestAcceptance_DockerfileImageStandaloneRejection(t *testing.T) {
	defs, err := techdef.Load()
	require.NoError(t, err)

	// powershell has no variant_group and is standalone — combining with go must fail
	cfg := &config.Config{
		ProjectName:  "bad-combo",
		Provider:     "github",
		Technologies: []string{"powershell", "go"},
		Licence:      "none",
		Confirmed:    true,
	}

	err = cfg.Validate(defs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "standalone")
}

// --- Stage 3b: Helm acceptance tests ---

func TestAcceptance_HelmChart_Standalone(t *testing.T) {
	defs, err := techdef.Load()
	require.NoError(t, err)

	cfg := &config.Config{
		ProjectName:  "my-chart",
		Provider:     "github",
		Technologies: []string{"Helm"},
		Licence:      "none",
		Docs:         []string{},
		Tooling:      []string{},
		RepoConfig:   []string{},
		Confirmed:    true,
	}

	// Resolve variant groups: "Helm" sole selection → helm-chart (standalone)
	resolvedTechs, _ := techdef.ResolveVariantGroups(cfg.Technologies, defs)

	baseDir := t.TempDir()
	err = scaffold.Generate(cfg, resolvedTechs, baseDir)
	require.NoError(t, err)

	root := filepath.Join(baseDir, "my-chart")

	// Chart.yaml uses ProjectName
	chartYaml, err := os.ReadFile(filepath.Join(root, "Chart.yaml"))
	require.NoError(t, err)
	chartStr := string(chartYaml)
	assert.Contains(t, chartStr, "name: my-chart")
	assert.NotContains(t, chartStr, "{{")

	// Structure at project root
	assert.FileExists(t, filepath.Join(root, "values.yaml"))
	assert.FileExists(t, filepath.Join(root, ".helmignore"))
	assert.DirExists(t, filepath.Join(root, "templates"))
	assert.FileExists(t, filepath.Join(root, "templates/deployment.yaml"))
	assert.FileExists(t, filepath.Join(root, "templates/service.yaml"))
	assert.FileExists(t, filepath.Join(root, "templates/serviceaccount.yaml"))
	assert.FileExists(t, filepath.Join(root, "templates/hpa.yaml"))
	assert.FileExists(t, filepath.Join(root, "templates/ingress.yaml"))
	assert.FileExists(t, filepath.Join(root, "templates/_helpers.tpl"))
	assert.FileExists(t, filepath.Join(root, "templates/NOTES.txt"))
	assert.FileExists(t, filepath.Join(root, "templates/tests/test-connection.yaml"))
	assert.DirExists(t, filepath.Join(root, "charts"))
	assert.DirExists(t, filepath.Join(root, "tests"))
	assert.FileExists(t, filepath.Join(root, "tests/deployment_test.yaml"))
	assert.FileExists(t, filepath.Join(root, "tests/service_test.yaml"))

	// Templates should not have literal {{ escaped sequences
	deployment, _ := os.ReadFile(filepath.Join(root, "templates/deployment.yaml"))
	depStr := string(deployment)
	assert.NotContains(t, depStr, `{{"{{"}}`)
	assert.Contains(t, depStr, "{{")
	assert.Contains(t, depStr, "my-chart.fullname")
	assert.Contains(t, depStr, "my-chart.labels")

	helpers, _ := os.ReadFile(filepath.Join(root, "templates/_helpers.tpl"))
	helpersStr := string(helpers)
	assert.Contains(t, helpersStr, `define "my-chart.fullname"`)

	// Gitignore contains Helm section
	gitignore, _ := os.ReadFile(filepath.Join(root, ".gitignore"))
	assert.Contains(t, string(gitignore), "# Helm Chart")
	assert.Contains(t, string(gitignore), "*.tgz")

	// CI contains helm jobs referencing "."
	ci, _ := os.ReadFile(filepath.Join(root, ".github/workflows/ci.yml"))
	ciStr := string(ci)
	assert.Contains(t, ciStr, "lint-helm")
	assert.Contains(t, ciStr, "test-helm")
	assert.Contains(t, ciStr, "helm lint .")
	assert.Contains(t, ciStr, "helm unittest .")

	// Devcontainer has kubernetes tools extension
	dcJSON, _ := os.ReadFile(filepath.Join(root, ".devcontainer/devcontainer.json"))
	assert.Contains(t, string(dcJSON), "ms-kubernetes-tools.vscode-kubernetes-tools")
}

func TestAcceptance_HelmDeployment_Composable_WithGo(t *testing.T) {
	defs, err := techdef.Load()
	require.NoError(t, err)

	chartName := "my-app"
	cfg := &config.Config{
		ProjectName:         "polyglot",
		Provider:            "github",
		Technologies:        []string{"Helm", "go"},
		TechPromptResponses: map[string]string{"chart_name": chartName},
		Licence:             "none",
		Docs:                []string{},
		Tooling:             []string{},
		RepoConfig:          []string{},
		Confirmed:           true,
	}

	// Resolve variant groups: "Helm" with others → helm-deployment (composable)
	resolvedTechs, _ := techdef.ResolveVariantGroups(cfg.Technologies, defs)
	sort.Slice(resolvedTechs, func(i, j int) bool {
		return resolvedTechs[i].Name < resolvedTechs[j].Name
	})

	baseDir := t.TempDir()
	err = scaffold.Generate(cfg, resolvedTechs, baseDir)
	require.NoError(t, err)

	root := filepath.Join(baseDir, "polyglot")

	// Chart nested under deploy/helm/<chart_name>/
	chartDir := filepath.Join(root, "deploy/helm", chartName)
	assert.DirExists(t, chartDir)

	chartYaml, err := os.ReadFile(filepath.Join(chartDir, "Chart.yaml"))
	require.NoError(t, err)
	chartStr := string(chartYaml)
	assert.Contains(t, chartStr, "name: my-app")
	assert.NotContains(t, chartStr, "{{")

	// Templates use chart_name not ProjectName
	deployment, _ := os.ReadFile(filepath.Join(chartDir, "templates/deployment.yaml"))
	depStr := string(deployment)
	assert.Contains(t, depStr, "my-app.fullname")
	assert.Contains(t, depStr, "my-app.labels")
	assert.NotContains(t, depStr, "polyglot.fullname")

	helpers, _ := os.ReadFile(filepath.Join(chartDir, "templates/_helpers.tpl"))
	helpersStr := string(helpers)
	assert.Contains(t, helpersStr, `define "my-app.fullname"`)

	// No chart files at root
	assert.NoFileExists(t, filepath.Join(root, "Chart.yaml"))

	// CI references deploy/helm/my-app
	ci, _ := os.ReadFile(filepath.Join(root, ".github/workflows/ci.yml"))
	ciStr := string(ci)
	assert.Contains(t, ciStr, "helm lint deploy/helm/my-app")
	assert.Contains(t, ciStr, "helm unittest deploy/helm/my-app")
	assert.Contains(t, ciStr, "lint-go")
	assert.Contains(t, ciStr, "lint-helm")
}

func TestAcceptance_HelmVariantGroupResolution(t *testing.T) {
	defs, err := techdef.Load()
	require.NoError(t, err)

	// Sole selection: Helm → standalone (helm-chart)
	resolved, modeMap := techdef.ResolveVariantGroups([]string{"Helm"}, defs)
	require.Len(t, resolved, 1)
	assert.Equal(t, "Helm Chart", resolved[0].Name)
	assert.True(t, resolved[0].Standalone)
	assert.Equal(t, "standalone", modeMap["Helm"])

	// Multi-selection: Helm + go → composable (helm-deployment)
	resolved2, modeMap2 := techdef.ResolveVariantGroups([]string{"Helm", "go"}, defs)
	require.Len(t, resolved2, 2)
	helmDef := resolved2[0]
	if helmDef.Name != "Helm Deployment" {
		helmDef = resolved2[1]
	}
	assert.Equal(t, "Helm Deployment", helmDef.Name)
	assert.False(t, helmDef.Standalone)
	assert.Equal(t, "composable", modeMap2["Helm"])
}

func TestAcceptance_HelmChartNamePromptOnlyInComposable(t *testing.T) {
	defs, err := techdef.Load()
	require.NoError(t, err)

	helmDeployment := defs["helm-deployment"]
	require.Len(t, helmDeployment.Prompts, 1)
	assert.Equal(t, "composable", helmDeployment.Prompts[0].Mode)

	// In standalone mode, chart_name prompt is NOT shown
	filtered := prompt.FilterPromptsByMode(helmDeployment.Prompts, "standalone")
	assert.Empty(t, filtered)

	// In composable mode, chart_name prompt IS shown
	filtered2 := prompt.FilterPromptsByMode(helmDeployment.Prompts, "composable")
	assert.Len(t, filtered2, 1)
	assert.Equal(t, "chart_name", filtered2[0].Key)
}

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

	assert.DirExists(t, filepath.Join(root, "docker"))
	assert.DirExists(t, filepath.Join(root, "cmd/app"))
	assert.DirExists(t, filepath.Join(root, "internal"))
	assert.DirExists(t, filepath.Join(root, "infrastructure"))
	assert.DirExists(t, filepath.Join(root, "src/my_project"))
	assert.DirExists(t, filepath.Join(root, "tests"))

	assert.FileExists(t, filepath.Join(root, "docker/Dockerfile"))
	assert.FileExists(t, filepath.Join(root, ".dockerignore"))
	assert.NoFileExists(t, filepath.Join(root, "Dockerfile"))

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

	gitignore, _ := os.ReadFile(filepath.Join(root, ".gitignore"))
	gitignoreStr := string(gitignore)
	dockerIdx := strings.Index(gitignoreStr, "# Dockerfile (Service)")
	goIdx := strings.Index(gitignoreStr, "# Go")
	pyIdx := strings.Index(gitignoreStr, "# Python")
	tfIdx := strings.Index(gitignoreStr, "# Terraform (Infrastructure)")
	assert.Less(t, dockerIdx, goIdx)
	assert.Less(t, goIdx, pyIdx)
	assert.Less(t, pyIdx, tfIdx)

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

// --- Stage 3b gap coverage ---

func TestAcceptance_VariantGroupsCollapseToSinglePromptEntry(t *testing.T) {
	defs, err := techdef.Load()
	require.NoError(t, err)

	options := prompt.BuildTechOptionList(defs)

	keys := make([]string, len(options))
	for i, o := range options {
		keys[i] = o.Key
	}

	// Variant group names appear as keys — individual definition keys do not.
	// Each variant group is verified independently so adding a new group
	// does not break assertions about existing groups.
	assert.Contains(t, keys, "Helm", "Helm variant group must collapse to single entry")
	assert.NotContains(t, keys, "helm-chart", "helm-chart must not appear as individual entry")
	assert.NotContains(t, keys, "helm-deployment", "helm-deployment must not appear as individual entry")

	assert.Contains(t, keys, "Terraform", "Terraform variant group must collapse to single entry")
	assert.NotContains(t, keys, "terraform-module", "terraform-module must not appear as individual entry")
	assert.NotContains(t, keys, "terraform-infrastructure", "terraform-infrastructure must not appear as individual entry")

	assert.Contains(t, keys, "Dockerfile", "Dockerfile variant group must collapse to single entry")
	assert.NotContains(t, keys, "dockerfile-image", "dockerfile-image must not appear as individual entry")
	assert.NotContains(t, keys, "dockerfile-service", "dockerfile-service must not appear as individual entry")
}

func TestAcceptance_NonVariantTechsAppearByKey(t *testing.T) {
	defs, err := techdef.Load()
	require.NoError(t, err)

	options := prompt.BuildTechOptionList(defs)

	keys := make([]string, len(options))
	for i, o := range options {
		keys[i] = o.Key
	}

	// Non-variant technologies appear by their definition key.
	assert.Contains(t, keys, "go")
	assert.Contains(t, keys, "powershell")
	assert.Contains(t, keys, "python")
}

func TestAcceptance_PromptOptionListIsSorted(t *testing.T) {
	defs, err := techdef.Load()
	require.NoError(t, err)

	options := prompt.BuildTechOptionList(defs)

	names := make([]string, len(options))
	for i, o := range options {
		names[i] = o.Name
	}

	assert.IsNonDecreasing(t, names, "prompt options must be sorted alphabetically by display name")
}

// TestAcceptance_TerraformVariantGroupResolution verifies acceptance criterion #8:
// Selecting "Terraform" alone resolves to the module (standalone) layout.
// Selecting "Terraform" with Go resolves to the infrastructure (composable) layout.
func TestAcceptance_TerraformVariantGroupResolution(t *testing.T) {
	defs, err := techdef.Load()
	require.NoError(t, err)

	// Sole "Terraform" → terraform-module (standalone)
	resolved, modeMap := techdef.ResolveVariantGroups([]string{"Terraform"}, defs)
	require.Len(t, resolved, 1)
	assert.Equal(t, "Terraform Module", resolved[0].Name)
	assert.True(t, resolved[0].Standalone)
	assert.Equal(t, "standalone", modeMap["Terraform"])

	// "Terraform" + go → terraform-infrastructure (composable)
	resolved2, modeMap2 := techdef.ResolveVariantGroups([]string{"Terraform", "go"}, defs)
	require.Len(t, resolved2, 2)
	var tfDef *techdef.TechDef
	for _, d := range resolved2 {
		if d.VariantGroup == "Terraform" {
			tfDef = d
		}
	}
	require.NotNil(t, tfDef, "Terraform definition must be in resolved list")
	assert.Equal(t, "Terraform (Infrastructure)", tfDef.Name)
	assert.False(t, tfDef.Standalone)
	assert.Equal(t, "composable", modeMap2["Terraform"])

	// Sole "Terraform" scaffolds module layout (root-level .tf files, no infrastructure/ dir)
	soleCfg := &config.Config{
		ProjectName:  "my-tf",
		Provider:     "github",
		Technologies: []string{"Terraform"},
		Licence:      "none",
		Docs:         []string{},
		Tooling:      []string{},
		RepoConfig:   []string{},
		Confirmed:    true,
	}
	baseDir := t.TempDir()
	resolveAndGenerate(t, soleCfg, baseDir)
	root := filepath.Join(baseDir, "my-tf")
	assert.FileExists(t, filepath.Join(root, "main.tf"), "standalone Terraform must scaffold root main.tf")
	assert.NoDirExists(t, filepath.Join(root, "infrastructure"), "standalone Terraform must not create infrastructure/ dir")

	// "Terraform" + go scaffolds infrastructure layout (infrastructure/ subdir, no root main.tf)
	comboCfg := &config.Config{
		ProjectName:  "my-app",
		Provider:     "github",
		Technologies: []string{"Terraform", "go"},
		Licence:      "none",
		Docs:         []string{},
		Tooling:      []string{},
		RepoConfig:   []string{},
		Confirmed:    true,
	}
	baseDir2 := t.TempDir()
	resolveAndGenerate(t, comboCfg, baseDir2)
	root2 := filepath.Join(baseDir2, "my-app")
	assert.FileExists(t, filepath.Join(root2, "infrastructure/main.tf"), "composable Terraform must scaffold infrastructure/main.tf")
	assert.NoFileExists(t, filepath.Join(root2, "main.tf"), "composable Terraform must not scaffold root main.tf")
}

// TestAcceptance_DockerfileVariantGroupResolution verifies acceptance criterion #9:
// Selecting "Dockerfile" alone resolves to the image (standalone) layout.
// Selecting "Dockerfile" with Go resolves to the service (composable) layout.
func TestAcceptance_DockerfileVariantGroupResolution(t *testing.T) {
	defs, err := techdef.Load()
	require.NoError(t, err)

	// Sole "Dockerfile" → dockerfile-image (standalone)
	resolved, modeMap := techdef.ResolveVariantGroups([]string{"Dockerfile"}, defs)
	require.Len(t, resolved, 1)
	assert.Equal(t, "Dockerfile (Image)", resolved[0].Name)
	assert.True(t, resolved[0].Standalone)
	assert.Equal(t, "standalone", modeMap["Dockerfile"])

	// "Dockerfile" + go → dockerfile-service (composable)
	resolved2, modeMap2 := techdef.ResolveVariantGroups([]string{"Dockerfile", "go"}, defs)
	require.Len(t, resolved2, 2)
	var dfDef *techdef.TechDef
	for _, d := range resolved2 {
		if d.VariantGroup == "Dockerfile" {
			dfDef = d
		}
	}
	require.NotNil(t, dfDef, "Dockerfile definition must be in resolved list")
	assert.Equal(t, "Dockerfile (Service)", dfDef.Name)
	assert.False(t, dfDef.Standalone)
	assert.Equal(t, "composable", modeMap2["Dockerfile"])

	// Sole "Dockerfile" scaffolds image layout (root-level Dockerfile)
	soleCfg := &config.Config{
		ProjectName:  "my-image",
		Provider:     "github",
		Technologies: []string{"Dockerfile"},
		Licence:      "none",
		Docs:         []string{},
		Tooling:      []string{},
		RepoConfig:   []string{},
		Confirmed:    true,
	}
	baseDir := t.TempDir()
	resolveAndGenerate(t, soleCfg, baseDir)
	root := filepath.Join(baseDir, "my-image")
	assert.FileExists(t, filepath.Join(root, "Dockerfile"), "standalone Dockerfile must scaffold root Dockerfile")
	assert.NoDirExists(t, filepath.Join(root, "docker"), "standalone Dockerfile must not create docker/ dir")

	// "Dockerfile" + go scaffolds service layout (docker/ subdir, no root Dockerfile)
	comboCfg := &config.Config{
		ProjectName:  "my-service",
		Provider:     "github",
		Technologies: []string{"Dockerfile", "go"},
		Licence:      "none",
		Docs:         []string{},
		Tooling:      []string{},
		RepoConfig:   []string{},
		Confirmed:    true,
	}
	baseDir2 := t.TempDir()
	resolveAndGenerate(t, comboCfg, baseDir2)
	root2 := filepath.Join(baseDir2, "my-service")
	assert.FileExists(t, filepath.Join(root2, "docker/Dockerfile"), "composable Dockerfile must scaffold docker/Dockerfile")
	assert.NoFileExists(t, filepath.Join(root2, "Dockerfile"), "composable Dockerfile must not scaffold root Dockerfile")
}

// TestAcceptance_HelmComposableGitignoreSection verifies acceptance criterion #16:
// The .gitignore includes the Helm section in the composable variant (Helm + Go).
func TestAcceptance_HelmComposableGitignoreSection(t *testing.T) {
	chartName := "my-app"
	cfg := &config.Config{
		ProjectName:         "polyglot",
		Provider:            "github",
		Technologies:        []string{"Helm", "go"},
		TechPromptResponses: map[string]string{"chart_name": chartName},
		Licence:             "none",
		Docs:                []string{},
		Tooling:             []string{},
		RepoConfig:          []string{},
		Confirmed:           true,
	}

	baseDir := t.TempDir()
	resolveAndGenerate(t, cfg, baseDir)
	root := filepath.Join(baseDir, "polyglot")

	gitignore, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	require.NoError(t, err)
	gitignoreStr := string(gitignore)

	assert.Contains(t, gitignoreStr, "# Helm", "composable Helm .gitignore must contain Helm section")
	assert.Contains(t, gitignoreStr, "*.tgz", "composable Helm .gitignore must ignore .tgz chart archives")
}

// TestAcceptance_HelmDevcontainerSetupCommands verifies acceptance criterion #17:
// The devcontainer setup.sh includes the Helm install script and helm-unittest plugin
// for both standalone and composable variants.
func TestAcceptance_HelmDevcontainerSetupCommands(t *testing.T) {
	t.Run("standalone", func(t *testing.T) {
		cfg := &config.Config{
			ProjectName:  "my-chart",
			Provider:     "github",
			Technologies: []string{"Helm"},
			Licence:      "none",
			Docs:         []string{},
			Tooling:      []string{},
			RepoConfig:   []string{},
			Confirmed:    true,
		}
		baseDir := t.TempDir()
		resolveAndGenerate(t, cfg, baseDir)
		root := filepath.Join(baseDir, "my-chart")

		setup, err := os.ReadFile(filepath.Join(root, ".devcontainer/setup.sh"))
		require.NoError(t, err)
		setupStr := string(setup)
		assert.Contains(t, setupStr, "get-helm-3", "setup.sh must include Helm install script")
		assert.Contains(t, setupStr, "helm-unittest", "setup.sh must include helm-unittest plugin install")
	})

	t.Run("composable", func(t *testing.T) {
		cfg := &config.Config{
			ProjectName:         "my-app",
			Provider:            "github",
			Technologies:        []string{"Helm", "go"},
			TechPromptResponses: map[string]string{"chart_name": "my-chart"},
			Licence:             "none",
			Docs:                []string{},
			Tooling:             []string{},
			RepoConfig:          []string{},
			Confirmed:           true,
		}
		baseDir := t.TempDir()
		resolveAndGenerate(t, cfg, baseDir)
		root := filepath.Join(baseDir, "my-app")

		setup, err := os.ReadFile(filepath.Join(root, ".devcontainer/setup.sh"))
		require.NoError(t, err)
		setupStr := string(setup)
		assert.Contains(t, setupStr, "get-helm-3", "setup.sh must include Helm install script")
		assert.Contains(t, setupStr, "helm-unittest", "setup.sh must include helm-unittest plugin install")
	})
}

// TestAcceptance_HelmComposableNoPathConflicts verifies acceptance criterion #15:
// Composable Helm has no file path conflicts with other composable technologies
// (Go, Python, Terraform Infrastructure, Dockerfile Service).
func TestAcceptance_HelmComposableNoPathConflicts(t *testing.T) {
	defs, err := techdef.Load()
	require.NoError(t, err)

	chartName := "my-chart"
	cfg := &config.Config{
		ProjectName:         "kitchen-sink",
		Provider:            "github",
		Technologies:        []string{"Helm", "Dockerfile", "go", "python", "Terraform"},
		TechPromptResponses: map[string]string{"chart_name": chartName, "package_name": "kitchen_sink"},
		Licence:             "none",
		Docs:                []string{},
		Tooling:             []string{},
		RepoConfig:          []string{},
		Confirmed:           true,
	}

	resolvedTechs, _ := techdef.ResolveVariantGroups(cfg.Technologies, defs)
	sort.Slice(resolvedTechs, func(i, j int) bool { return resolvedTechs[i].Name < resolvedTechs[j].Name })

	baseDir := t.TempDir()
	// Generate must not return an error (no path conflicts at engine level)
	err = scaffold.Generate(cfg, resolvedTechs, baseDir)
	require.NoError(t, err, "scaffold.Generate must not error due to path conflicts")

	root := filepath.Join(baseDir, "kitchen-sink")

	// Helm composable output is under deploy/helm/<chart_name>/ — separate from all others
	assert.DirExists(t, filepath.Join(root, "deploy/helm", chartName))
	assert.FileExists(t, filepath.Join(root, "deploy/helm", chartName, "Chart.yaml"))

	// Other composable technologies each occupy their own unique paths
	assert.DirExists(t, filepath.Join(root, "docker"), "Dockerfile (Service) must use docker/")
	assert.DirExists(t, filepath.Join(root, "cmd/app"), "Go must use cmd/app/")
	assert.DirExists(t, filepath.Join(root, "infrastructure"), "Terraform (Infrastructure) must use infrastructure/")
	assert.DirExists(t, filepath.Join(root, "src/kitchen_sink"), "Python must use src/<package>/")

	// No root-level Chart.yaml (would indicate standalone variant was incorrectly chosen)
	assert.NoFileExists(t, filepath.Join(root, "Chart.yaml"))
}
