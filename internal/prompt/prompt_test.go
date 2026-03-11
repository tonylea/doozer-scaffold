package prompt_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tonylea/doozer-scaffold/internal/prompt"
	"github.com/tonylea/doozer-scaffold/internal/techdef"
)

func makeSyntheticDefs() map[string]*techdef.TechDef {
	return map[string]*techdef.TechDef{
		"helm-chart": {
			Name:         "Helm Chart",
			VariantGroup: "Helm",
			Standalone:   true,
			Structure:    []techdef.StructureEntry{{Path: "Chart.yaml"}},
			Gitignore:    "*.tgz",
		},
		"helm-deployment": {
			Name:         "Helm Deployment",
			VariantGroup: "Helm",
			Standalone:   false,
			Structure:    []techdef.StructureEntry{{Path: "deploy/helm/"}},
			Gitignore:    "*.tgz",
		},
		"go": {
			Name:       "Go",
			Standalone: false,
			Structure:  []techdef.StructureEntry{{Path: "cmd/"}},
			Gitignore:  "*.exe",
		},
		"powershell": {
			Name:       "PowerShell Module",
			Standalone: true,
			Structure:  []techdef.StructureEntry{{Path: "src/"}},
			Gitignore:  "*.nupkg",
		},
	}
}

func TestBuildTechOptionList_CollapsesVariantGroups(t *testing.T) {
	defs := makeSyntheticDefs()
	options := prompt.BuildTechOptionList(defs)

	keys := make([]string, len(options))
	names := make([]string, len(options))
	for i, o := range options {
		keys[i] = o.Key
		names[i] = o.Name
	}

	// "Helm" appears once (not "Helm Chart" and "Helm Deployment")
	assert.Contains(t, keys, "Helm")
	assert.NotContains(t, keys, "helm-chart")
	assert.NotContains(t, keys, "helm-deployment")

	// Regular techs appear by key
	assert.Contains(t, keys, "go")
	assert.Contains(t, keys, "powershell")

	// Total: 3 options (Helm, Go, PowerShell)
	require.Len(t, options, 3)
}

func TestBuildTechOptionList_VariantGroupDisplayName(t *testing.T) {
	defs := makeSyntheticDefs()
	options := prompt.BuildTechOptionList(defs)

	for _, o := range options {
		if o.Key == "Helm" {
			assert.Equal(t, "Helm", o.Name)
			return
		}
	}
	t.Fatal("Helm variant group option not found")
}

func TestBuildTechOptionList_SortedAlphabetically(t *testing.T) {
	defs := makeSyntheticDefs()
	options := prompt.BuildTechOptionList(defs)

	require.Len(t, options, 3)
	assert.Equal(t, "Go", options[0].Name)
	assert.Equal(t, "Helm", options[1].Name)
	assert.Equal(t, "PowerShell Module", options[2].Name)
}

func TestFilterPromptsByMode_NoMode_AlwaysShown(t *testing.T) {
	prompts := []techdef.PromptDef{
		{Key: "name", Title: "Name:", Type: "text"},
	}
	result := prompt.FilterPromptsByMode(prompts, "composable")
	assert.Len(t, result, 1)
	result2 := prompt.FilterPromptsByMode(prompts, "standalone")
	assert.Len(t, result2, 1)
}

func TestFilterPromptsByMode_ComposableMode_OnlyShownWhenComposable(t *testing.T) {
	prompts := []techdef.PromptDef{
		{Key: "chart_name", Title: "Chart name:", Type: "text", Mode: "composable"},
	}
	result := prompt.FilterPromptsByMode(prompts, "composable")
	assert.Len(t, result, 1)

	result2 := prompt.FilterPromptsByMode(prompts, "standalone")
	assert.Len(t, result2, 0)

	result3 := prompt.FilterPromptsByMode(prompts, "")
	assert.Len(t, result3, 0)
}

func TestFilterPromptsByMode_StandaloneMode_OnlyShownWhenStandalone(t *testing.T) {
	prompts := []techdef.PromptDef{
		{Key: "flavor", Title: "Flavor:", Type: "text", Mode: "standalone"},
	}
	result := prompt.FilterPromptsByMode(prompts, "standalone")
	assert.Len(t, result, 1)

	result2 := prompt.FilterPromptsByMode(prompts, "composable")
	assert.Len(t, result2, 0)
}

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
