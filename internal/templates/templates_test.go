package templates_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tonylea/doozer-scaffold/internal/templates"
)

func TestRenderReadme(t *testing.T) {
	data := templates.TemplateData{
		ProjectName: "my-module",
	}
	result, err := templates.Render("README.md.tmpl", data)
	require.NoError(t, err)
	assert.Equal(t, "# my-module\n", result)
}

func TestRenderMITLicence(t *testing.T) {
	data := templates.TemplateData{
		ProjectName: "my-module",
		Year:        "2025",
	}
	result, err := templates.Render("licences/MIT.tmpl", data)
	require.NoError(t, err)
	assert.Contains(t, result, "MIT License")
	assert.Contains(t, result, "my-module")
	assert.Contains(t, result, "2025")
}

func TestRenderCIWorkflow(t *testing.T) {
	data := templates.TemplateData{}
	result, err := templates.Render("github/ci.yml.tmpl", data)
	require.NoError(t, err)
	assert.Contains(t, result, "name: CI")
	assert.Contains(t, result, "actions/checkout@v4")
}

func TestRenderContributing(t *testing.T) {
	data := templates.TemplateData{
		ProjectName: "my-module",
	}
	result, err := templates.Render("docs/CONTRIBUTING.md.tmpl", data)
	require.NoError(t, err)
	assert.Contains(t, result, "Contributing to my-module")
}

func TestRenderDockerfile(t *testing.T) {
	data := templates.TemplateData{}
	result, err := templates.Render("devcontainer/Dockerfile.tmpl", data)
	require.NoError(t, err)
	assert.Contains(t, result, "FROM mcr.microsoft.com/devcontainers/base:ubuntu")
}

func TestRenderBaseSh(t *testing.T) {
	data := templates.TemplateData{}
	result, err := templates.Render("devcontainer/setup/base.sh.tmpl", data)
	require.NoError(t, err)
	assert.Contains(t, result, "#!/bin/bash")
	assert.Contains(t, result, "markdownlint-cli2")
}

func TestRenderUnknownTemplate(t *testing.T) {
	data := templates.TemplateData{}
	_, err := templates.Render("nonexistent.tmpl", data)
	assert.Error(t, err)
}
