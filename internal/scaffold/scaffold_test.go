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

// --- Stage 2: updated signature tests ---

func TestComposeGitignore_SingleTech(t *testing.T) {
	techs := []*techdef.TechDef{
		{
			Name:      "PowerShell Module",
			Gitignore: "*.ps1xml\n*.nupkg\n",
		},
	}

	result := scaffold.ComposeGitignore(techs)
	assert.Contains(t, result, "# PowerShell Module")
	assert.Contains(t, result, "*.ps1xml")
	assert.Contains(t, result, "*.nupkg")
}

func TestComposeGitignoreMultipleTechs(t *testing.T) {
	techs := []*techdef.TechDef{
		{Name: "Go", Gitignore: "*.exe\nbin/\n"},
		{Name: "Terraform (Infrastructure)", Gitignore: ".terraform/\n*.tfstate\n"},
	}
	result := scaffold.ComposeGitignore(techs)
	assert.Contains(t, result, "# Go")
	assert.Contains(t, result, "# Terraform (Infrastructure)")
	goIdx := indexOf(result, "# Go")
	tfIdx := indexOf(result, "# Terraform (Infrastructure)")
	assert.Less(t, goIdx, tfIdx)
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func TestRejectsExistingDirectory(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "my-module")
	require.NoError(t, os.Mkdir(target, 0o755))

	tech := loadPowerShellDef(t)

	cfg := &config.Config{
		ProjectName:  "my-module",
		Provider:     "github",
		Technologies: []string{"powershell"},
		Licence:      "none",
		Confirmed:    true,
	}

	err := scaffold.Generate(cfg, []*techdef.TechDef{tech}, dir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestGenerate_MinimumSelections(t *testing.T) {
	dir := t.TempDir()
	tech := loadPowerShellDef(t)

	cfg := &config.Config{
		ProjectName:  "my-project",
		Provider:     "github",
		Technologies: []string{"powershell"},
		Licence:      "none",
		Docs:         []string{},
		Tooling:      []string{},
		RepoConfig:   []string{},
		Confirmed:    true,
	}

	err := scaffold.Generate(cfg, []*techdef.TechDef{tech}, dir)
	require.NoError(t, err)

	root := filepath.Join(dir, "my-project")
	assert.DirExists(t, root)
	assert.FileExists(t, filepath.Join(root, "README.md"))
	assert.FileExists(t, filepath.Join(root, ".gitignore"))
	assert.FileExists(t, filepath.Join(root, ".github/workflows/ci.yml"))
	assert.FileExists(t, filepath.Join(root, "Makefile"))
	assert.NoFileExists(t, filepath.Join(root, "LICENSE"))
	assert.NoFileExists(t, filepath.Join(root, "CONTRIBUTING.md"))
}

func TestGenerate_WithMITLicence(t *testing.T) {
	dir := t.TempDir()
	tech := loadPowerShellDef(t)

	cfg := &config.Config{
		ProjectName:  "licensed-project",
		Provider:     "github",
		Technologies: []string{"powershell"},
		Licence:      "mit",
		Confirmed:    true,
	}

	err := scaffold.Generate(cfg, []*techdef.TechDef{tech}, dir)
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
		ProjectName:  "dc-project",
		Provider:     "github",
		Technologies: []string{"powershell"},
		Licence:      "none",
		Confirmed:    true,
	}

	err := scaffold.Generate(cfg, []*techdef.TechDef{tech}, dir)
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

// --- Stage 2: New tests ---

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

	assert.Contains(t, ciStr, "lint-go")
	assert.Contains(t, ciStr, "test-go")
	assert.Contains(t, ciStr, "build")
	assert.Contains(t, ciStr, "actions/setup-go@v5")
	assert.Contains(t, ciStr, "golangci-lint run")
	assert.Contains(t, ciStr, "go test")
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

	assert.Contains(t, ciStr, "lint-go")
	assert.Contains(t, ciStr, "lint-python")
	assert.Contains(t, ciStr, "test-go")
	assert.Contains(t, ciStr, "test-python")
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

	assert.Contains(t, ciStr, "lint-powershell")
	assert.Contains(t, ciStr, "test-powershell")
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
