package config

import (
	"fmt"

	"github.com/tonylea/doozer-scaffold/internal/techdef"
)

// Config holds the user's selections from the interactive prompt.
type Config struct {
	ProjectName         string
	Provider            string
	Technologies        []string          // Multi-select; replaces single Technology field
	TechPromptResponses map[string]string // Responses to technology-driven prompts
	Licence             string
	Docs                []string
	Tooling             []string
	RepoConfig          []string
	Confirmed           bool
}

// Validate checks that the Config has all required fields set and that
// technology constraints are satisfied.
func (c *Config) Validate(techDefs map[string]*techdef.TechDef) error {
	if c.ProjectName == "" {
		return fmt.Errorf("project name is required")
	}
	if c.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if len(c.Technologies) == 0 {
		return fmt.Errorf("at least one technology must be selected")
	}
	if c.Licence == "" {
		return fmt.Errorf("licence is required")
	}
	variantGroups := techdef.BuildVariantGroups(techDefs)
	// Validate all technology keys exist (in defs or as a variant group name)
	for _, key := range c.Technologies {
		if _, ok := techDefs[key]; !ok {
			if _, ok := variantGroups[key]; !ok {
				return fmt.Errorf("unknown technology '%s'", key)
			}
		}
	}
	// Standalone constraint: only applies to non-variant-group techs
	if len(c.Technologies) > 1 {
		for _, key := range c.Technologies {
			if _, isGroup := variantGroups[key]; isGroup {
				continue // Variant group techs are always composable-capable
			}
			def, ok := techDefs[key]
			if !ok {
				continue
			}
			if def.Standalone && def.VariantGroup == "" {
				return fmt.Errorf("technology '%s' is standalone and cannot be combined with others", def.Name)
			}
		}
	}
	return nil
}
