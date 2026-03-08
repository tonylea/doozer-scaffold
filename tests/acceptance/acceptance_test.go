package acceptance_test

import (
	"os"
	"path/filepath"
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
}

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
		ProjectName: "declined-project",
		Provider:    "github",
		Technology:  "powershell",
		Licence:     "mit",
		Confirmed:   false,
	}

	assert.False(t, cfg.Confirmed)
	assert.NoDirExists(t, filepath.Join(baseDir, "declined-project"))
}
