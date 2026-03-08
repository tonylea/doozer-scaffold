package scaffold_test

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

// loadPowerShellDef loads the powershell technology definition for use in tests.
func loadPowerShellDef(t *testing.T) *techdef.TechDef {
	t.Helper()
	defs, err := techdef.Load()
	require.NoError(t, err)
	require.Contains(t, defs, "powershell")
	return defs["powershell"]
}

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

func TestRejectsExistingDirectory(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "my-module")
	require.NoError(t, os.Mkdir(target, 0o755))

	tech := loadPowerShellDef(t)

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

func TestGenerate_MinimumSelections(t *testing.T) {
	dir := t.TempDir()
	tech := loadPowerShellDef(t)

	cfg := &config.Config{
		ProjectName: "my-project",
		Provider:    "github",
		Technology:  "powershell",
		Licence:     "none",
		Docs:        []string{},
		Tooling:     []string{},
		RepoConfig:  []string{},
		Confirmed:   true,
	}

	err := scaffold.Generate(cfg, tech, dir)
	require.NoError(t, err)

	root := filepath.Join(dir, "my-project")
	assert.DirExists(t, root)
	assert.FileExists(t, filepath.Join(root, "README.md"))
	assert.FileExists(t, filepath.Join(root, ".gitignore"))
	assert.FileExists(t, filepath.Join(root, ".github/workflows/ci.yml"))
	assert.NoFileExists(t, filepath.Join(root, "LICENSE"))
	assert.NoFileExists(t, filepath.Join(root, "CONTRIBUTING.md"))
}

func TestGenerate_WithMITLicence(t *testing.T) {
	dir := t.TempDir()
	tech := loadPowerShellDef(t)

	cfg := &config.Config{
		ProjectName: "licensed-project",
		Provider:    "github",
		Technology:  "powershell",
		Licence:     "mit",
		Confirmed:   true,
	}

	err := scaffold.Generate(cfg, tech, dir)
	require.NoError(t, err)

	licenceContent, err := os.ReadFile(filepath.Join(dir, "licensed-project", "LICENSE"))
	require.NoError(t, err)
	assert.Contains(t, string(licenceContent), "MIT License")
	assert.Contains(t, string(licenceContent), "licensed-project")
}

func TestGenerate_DevcontainerAlwaysGenerated(t *testing.T) {
	dir := t.TempDir()
	tech := loadPowerShellDef(t)

	cfg := &config.Config{
		ProjectName: "dc-project",
		Provider:    "github",
		Technology:  "powershell",
		Licence:     "none",
		Confirmed:   true,
	}

	err := scaffold.Generate(cfg, tech, dir)
	require.NoError(t, err)

	root := filepath.Join(dir, "dc-project")
	assert.FileExists(t, filepath.Join(root, ".devcontainer/devcontainer.json"))
	assert.FileExists(t, filepath.Join(root, ".devcontainer/Dockerfile"))
	assert.FileExists(t, filepath.Join(root, ".devcontainer/setup.sh"))

	// setup.sh must be executable
	info, err := os.Stat(filepath.Join(root, ".devcontainer/setup.sh"))
	require.NoError(t, err)
	assert.True(t, info.Mode().Perm()&0o111 != 0, "setup.sh must be executable")

	// devcontainer.json must include project name and base + tech features
	dcContent, err := os.ReadFile(filepath.Join(root, ".devcontainer/devcontainer.json"))
	require.NoError(t, err)
	assert.Contains(t, string(dcContent), "dc-project")
	assert.Contains(t, string(dcContent), "ghcr.io/devcontainers/features/node:1")
	assert.Contains(t, string(dcContent), "ghcr.io/devcontainers/features/powershell:1")
	assert.Contains(t, string(dcContent), "ms-vscode.powershell")
}
