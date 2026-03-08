package templates

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
)

//go:embed templates/*
var templateFS embed.FS

// TemplateData holds variables available to all templates.
type TemplateData struct {
	ProjectName string
	Year        string // e.g. "2025", used in licence files
}

// Render renders a named template with the given data and returns the result as a string.
// The name is a path relative to the templates directory (e.g. "README.md.tmpl").
func Render(name string, data TemplateData) (string, error) {
	path := "templates/" + name
	content, err := templateFS.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("loading template %q: %w", name, err)
	}

	tmpl, err := template.New(name).Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("parsing template %q: %w", name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("rendering template %q: %w", name, err)
	}

	return buf.String(), nil
}
