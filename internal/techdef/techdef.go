package techdef

import (
	"embed"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed technologies/*
var techFS embed.FS

// TechDef represents a parsed technology definition.
type TechDef struct {
	Name         string           `yaml:"name"`
	VariantGroup string           `yaml:"variant_group,omitempty"`
	Standalone   bool             `yaml:"standalone"`
	Prompts      []PromptDef      `yaml:"prompts,omitempty"`
	Structure    []StructureEntry `yaml:"structure"`
	Gitignore    string           `yaml:"gitignore"`
	Devcontainer DevcontainerDef  `yaml:"devcontainer"`
	CI           *CIDef           `yaml:"ci,omitempty"`
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

// DevcontainerDef holds the devcontainer configuration from a technology definition.
type DevcontainerDef struct {
	Features   map[string]interface{} `yaml:"features"`
	Extensions []string               `yaml:"extensions"`
	Setup      string                 `yaml:"setup"`
}

// PromptDef defines a user prompt driven by a technology definition.
type PromptDef struct {
	Key         string      `yaml:"key"`
	Title       string      `yaml:"title"`
	Type        string      `yaml:"type"`              // "text", "select", "multi_select"
	DefaultFrom string      `yaml:"default_from,omitempty"`
	Options     []OptionDef `yaml:"options,omitempty"`
	Mode        string      `yaml:"mode,omitempty"` // "", "standalone", "composable"
}

// OptionDef defines a single option for select/multi_select prompts.
type OptionDef struct {
	Label string `yaml:"label"`
	Value string `yaml:"value"`
}

// CIDef defines a technology's CI job contribution.
type CIDef struct {
	JobName    string        `yaml:"job_name"`
	SetupSteps []CISetupStep `yaml:"setup_steps,omitempty"`
	LintSteps  []CIStep      `yaml:"lint_steps"`
	TestSteps  []CIStep      `yaml:"test_steps"`
}

// CISetupStep is a setup step in a CI job (supports uses or run).
type CISetupStep struct {
	Name string            `yaml:"name"`
	Uses string            `yaml:"uses,omitempty"`
	With map[string]string `yaml:"with,omitempty"`
	Run  string            `yaml:"run,omitempty"`
}

// CIStep is a lint or test step in a CI job.
type CIStep struct {
	Name string `yaml:"name"`
	Run  string `yaml:"run"`
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

var promptKeyPattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)

// Validate checks that a TechDef is well-formed. The key is the technology's
// filename key, used for error messages.
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

	// Prompt validation
	for i, p := range t.Prompts {
		if p.Key == "" {
			return fmt.Errorf("technology '%s': prompts[%d] key is required", key, i)
		}
		if !promptKeyPattern.MatchString(p.Key) {
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
		if p.Mode != "" && p.Mode != "standalone" && p.Mode != "composable" {
			return fmt.Errorf("technology '%s': prompts[%d] mode must be empty, 'standalone', or 'composable'", key, i)
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
