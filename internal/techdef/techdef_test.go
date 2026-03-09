package techdef_test

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tonylea/doozer-scaffold/internal/techdef"
)

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

func TestStructureEntryIsDir(t *testing.T) {
	dirEntry := techdef.StructureEntry{Path: "src/public/"}
	fileEntry := techdef.StructureEntry{Path: "src/MyModule.psm1"}

	assert.True(t, dirEntry.IsDir())
	assert.False(t, fileEntry.IsDir())
}

func TestTechDefValidation_Valid(t *testing.T) {
	def := &techdef.TechDef{
		Name:      "Test",
		Structure: []techdef.StructureEntry{{Path: "src/"}},
		Gitignore: "*.log",
	}
	assert.NoError(t, def.Validate("test"))
}

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

func TestTechDefValidation_EmptyStructure(t *testing.T) {
	def := &techdef.TechDef{
		Name:      "Test",
		Structure: []techdef.StructureEntry{},
		Gitignore: "*.log",
	}
	err := def.Validate("test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one entry")
}

func TestTechDefValidation_EmptyPathInStructure(t *testing.T) {
	def := &techdef.TechDef{
		Name:      "Test",
		Structure: []techdef.StructureEntry{{Path: ""}},
		Gitignore: "*.log",
	}
	err := def.Validate("test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty path")
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

func TestTechDefValidation_EmptyGitignore(t *testing.T) {
	def := &techdef.TechDef{
		Name:      "Test",
		Structure: []techdef.StructureEntry{{Path: "src/"}},
		Gitignore: "   ",
	}
	err := def.Validate("test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gitignore is required")
}

func TestPowerShellDefIsValid(t *testing.T) {
	defs, err := techdef.Load()
	require.NoError(t, err)
	require.Contains(t, defs, "powershell")
	assert.NoError(t, defs["powershell"].Validate("powershell"))
}

// --- Stage 2: New field tests ---

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

func TestTechDefStandaloneField(t *testing.T) {
	defs, err := techdef.Load()
	require.NoError(t, err)

	assert.True(t, defs["powershell"].Standalone)
	assert.True(t, defs["terraform-module"].Standalone)
	assert.False(t, defs["go"].Standalone)
	assert.False(t, defs["terraform-infrastructure"].Standalone)
	assert.False(t, defs["python"].Standalone)
}

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

func TestAllNewDefsPassValidation(t *testing.T) {
	defs, err := techdef.Load()
	require.NoError(t, err)
	for key, def := range defs {
		assert.NoError(t, def.Validate(key), "validation failed for %s", key)
	}
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
