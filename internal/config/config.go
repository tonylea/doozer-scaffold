package config

import "fmt"

// Config holds the user's selections from the interactive prompt.
type Config struct {
	ProjectName string
	Provider    string
	Technology  string
	Licence     string
	Docs        []string
	Tooling     []string
	RepoConfig  []string
	Confirmed   bool
}

// Validate checks that the Config has all required fields set.
func (c *Config) Validate() error {
	if c.ProjectName == "" {
		return fmt.Errorf("project name is required")
	}
	if c.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if c.Technology == "" {
		return fmt.Errorf("technology is required")
	}
	if c.Licence == "" {
		return fmt.Errorf("licence is required")
	}
	return nil
}
