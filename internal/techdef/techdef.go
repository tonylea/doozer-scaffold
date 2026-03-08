package techdef

import (
	"embed"
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed technologies/*
var techFS embed.FS

// TechDef represents a parsed technology definition.
type TechDef struct {
	Name         string           `yaml:"name"`
	Structure    []StructureEntry `yaml:"structure"`
	Gitignore    string           `yaml:"gitignore"`
	Devcontainer DevcontainerDef  `yaml:"devcontainer"`
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
	return nil
}
