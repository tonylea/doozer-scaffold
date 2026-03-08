package techdef_test

import (
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
