package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tonylea/doozer-scaffold/internal/config"
	"github.com/tonylea/doozer-scaffold/internal/techdef"
)

func makeTechDefs() map[string]*techdef.TechDef {
	defs, err := techdef.Load()
	if err != nil {
		panic(err)
	}
	return defs
}

func TestConfigRequiresProjectName(t *testing.T) {
	techDefs := makeTechDefs()
	cfg := &config.Config{
		ProjectName:  "",
		Provider:     "github",
		Technologies: []string{"powershell"},
		Licence:      "mit",
	}
	err := cfg.Validate(techDefs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project name")
}

func TestConfigRequiresProvider(t *testing.T) {
	techDefs := makeTechDefs()
	cfg := &config.Config{
		ProjectName:  "my-module",
		Provider:     "",
		Technologies: []string{"powershell"},
		Licence:      "mit",
	}
	err := cfg.Validate(techDefs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider")
}

func TestConfigRequiresLicence(t *testing.T) {
	techDefs := makeTechDefs()
	cfg := &config.Config{
		ProjectName:  "my-module",
		Provider:     "github",
		Technologies: []string{"powershell"},
		Licence:      "",
	}
	err := cfg.Validate(techDefs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "licence")
}

func TestConfigValidConfig(t *testing.T) {
	techDefs := makeTechDefs()
	cfg := &config.Config{
		ProjectName:  "my-module",
		Provider:     "github",
		Technologies: []string{"powershell"},
		Licence:      "mit",
	}
	assert.NoError(t, cfg.Validate(techDefs))
}

func TestConfigValidConfigNoneLicence(t *testing.T) {
	techDefs := makeTechDefs()
	cfg := &config.Config{
		ProjectName:  "my-module",
		Provider:     "github",
		Technologies: []string{"powershell"},
		Licence:      "none",
	}
	assert.NoError(t, cfg.Validate(techDefs))
}

func TestConfigRequiresAtLeastOneTech(t *testing.T) {
	techDefs := makeTechDefs()
	cfg := &config.Config{
		ProjectName:  "test",
		Provider:     "github",
		Technologies: []string{},
		Licence:      "none",
	}
	err := cfg.Validate(techDefs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one")
}

func TestConfigRejectsStandaloneWithOtherTechs(t *testing.T) {
	techDefs := makeTechDefs()
	cfg := &config.Config{
		ProjectName:  "test",
		Provider:     "github",
		Technologies: []string{"powershell", "go"},
		Licence:      "none",
	}
	err := cfg.Validate(techDefs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "standalone")
}

func TestConfigAllowsMultipleComposableTechs(t *testing.T) {
	techDefs := makeTechDefs()
	cfg := &config.Config{
		ProjectName:  "test",
		Provider:     "github",
		Technologies: []string{"go", "terraform-infrastructure"},
		Licence:      "none",
	}
	err := cfg.Validate(techDefs)
	assert.NoError(t, err)
}

func TestConfigAllowsSingleStandaloneTech(t *testing.T) {
	techDefs := makeTechDefs()
	cfg := &config.Config{
		ProjectName:  "test",
		Provider:     "github",
		Technologies: []string{"terraform-module"},
		Licence:      "none",
	}
	err := cfg.Validate(techDefs)
	assert.NoError(t, err)
}
