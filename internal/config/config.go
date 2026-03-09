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
	// Validate all technology keys exist
	for _, key := range c.Technologies {
		if _, ok := techDefs[key]; !ok {
			return fmt.Errorf("unknown technology '%s'", key)
		}
	}
	// Standalone constraint
	if len(c.Technologies) > 1 {
		for _, key := range c.Technologies {
			def := techDefs[key]
			if def.Standalone {
				return fmt.Errorf("technology '%s' is standalone and cannot be combined with others", def.Name)
			}
		}
	}
	return nil
}
