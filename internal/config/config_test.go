package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tonylea/doozer-scaffold/internal/config"
)

func TestConfigRequiresProjectName(t *testing.T) {
	cfg := &config.Config{
		ProjectName: "",
		Provider:    "github",
		Technology:  "powershell",
		Licence:     "mit",
	}
	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project name")
}

func TestConfigRequiresProvider(t *testing.T) {
	cfg := &config.Config{
		ProjectName: "my-module",
		Provider:    "",
		Technology:  "powershell",
		Licence:     "mit",
	}
	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider")
}

func TestConfigRequiresTechnology(t *testing.T) {
	cfg := &config.Config{
		ProjectName: "my-module",
		Provider:    "github",
		Technology:  "",
		Licence:     "mit",
	}
	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "technology")
}

func TestConfigRequiresLicence(t *testing.T) {
	cfg := &config.Config{
		ProjectName: "my-module",
		Provider:    "github",
		Technology:  "powershell",
		Licence:     "",
	}
	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "licence")
}

func TestConfigValidConfig(t *testing.T) {
	cfg := &config.Config{
		ProjectName: "my-module",
		Provider:    "github",
		Technology:  "powershell",
		Licence:     "mit",
	}
	assert.NoError(t, cfg.Validate())
}

func TestConfigValidConfigNoneLicence(t *testing.T) {
	cfg := &config.Config{
		ProjectName: "my-module",
		Provider:    "github",
		Technology:  "powershell",
		Licence:     "none",
	}
	assert.NoError(t, cfg.Validate())
}
