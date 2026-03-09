package scaffold

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/tonylea/doozer-scaffold/internal/config"
	"github.com/tonylea/doozer-scaffold/internal/techdef"
	"github.com/tonylea/doozer-scaffold/internal/templates"
)

// Generate creates the scaffold in a subdirectory of baseDir named after cfg.ProjectName.
// techs is the ordered list of selected technology definitions (sorted by key).
func Generate(cfg *config.Config, techs []*techdef.TechDef, baseDir string) error {
	targetDir := filepath.Join(baseDir, cfg.ProjectName)

	// Check target doesn't already exist
	if _, err := os.Stat(targetDir); err == nil {
		return fmt.Errorf("directory '%s' already exists", cfg.ProjectName)
	}

	// Build template data
	templateData := buildTemplateData(cfg)

	// Detect conflicts before creating anything
	if err := DetectPathConflicts(techs, templateData); err != nil {
		return err
	}
	if err := DetectPromptKeyConflicts(techs); err != nil {
		return err
	}

	// Create directory structure and files; on any error, clean up targetDir
	if err := createScaffold(cfg, techs, targetDir, templateData); err != nil {
		_ = os.RemoveAll(targetDir)
		return fmt.Errorf("scaffold generation failed: %w", err)
	}

	return nil
}

func buildTemplateData(cfg *config.Config) map[string]string {
	data := map[string]string{
		"ProjectName": cfg.ProjectName,
		"Year":        strconv.Itoa(time.Now().Year()),
	}
	for k, v := range cfg.TechPromptResponses {
		data[k] = v
	}
	return data
}

// ResolveTemplate executes a Go template string with the given data map.
func ResolveTemplate(tmplStr string, data map[string]string) (string, error) {
	t, err := template.New("").Parse(tmplStr)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// DetectPathConflicts checks for file path conflicts across all selected technologies.
// Directory paths (ending with /) are allowed to overlap.
func DetectPathConflicts(techs []*techdef.TechDef, templateData map[string]string) error {
	seen := make(map[string]string) // resolved path → technology name
	for _, tech := range techs {
		for _, entry := range tech.Structure {
			resolvedPath, err := ResolveTemplate(entry.Path, templateData)
			if err != nil {
				return fmt.Errorf("resolving path in '%s': %w", tech.Name, err)
			}
			if strings.HasSuffix(resolvedPath, "/") {
				continue // Directory overlaps are fine
			}
			if existing, ok := seen[resolvedPath]; ok {
				return fmt.Errorf(
					"path conflict: '%s' is defined by both '%s' and '%s'",
					resolvedPath, existing, tech.Name,
				)
			}
			seen[resolvedPath] = tech.Name
		}
	}
	return nil
}

// DetectPromptKeyConflicts checks for duplicate prompt keys across all selected technologies.
func DetectPromptKeyConflicts(techs []*techdef.TechDef) error {
	seen := make(map[string]string) // key → technology name
	for _, tech := range techs {
		for _, p := range tech.Prompts {
			if existing, ok := seen[p.Key]; ok {
				return fmt.Errorf(
					"prompt key conflict: '%s' is defined by both '%s' and '%s'",
					p.Key, existing, tech.Name,
				)
			}
			seen[p.Key] = tech.Name
		}
	}
	return nil
}

func createScaffold(cfg *config.Config, techs []*techdef.TechDef, targetDir string, templateData map[string]string) error {
	// 1. Create target directory
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("creating target directory: %w", err)
	}

	data := templates.TemplateData{
		ProjectName: cfg.ProjectName,
		Year:        time.Now().Format("2006"),
	}

	// 2. Universal outputs
	if err := renderTemplateToFile(targetDir, "README.md", "README.md.tmpl", data, 0o644); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(targetDir, "Makefile"), []byte{}, 0o644); err != nil {
		return fmt.Errorf("creating Makefile: %w", err)
	}

	gitignore := ComposeGitignore(techs)
	if err := writeFile(filepath.Join(targetDir, ".gitignore"), []byte(gitignore), 0o644); err != nil {
		return err
	}

	ciContent, err := RenderCIConfig(techs)
	if err != nil {
		return fmt.Errorf("rendering CI config: %w", err)
	}
	if err := writeFile(filepath.Join(targetDir, ".github/workflows/ci.yml"), ciContent, 0o644); err != nil {
		return err
	}

	// 3. Technology structure (with template resolution)
	for _, tech := range techs {
		if err := createStructureWithTemplate(targetDir, tech.Structure, templateData); err != nil {
			return err
		}
	}

	// 4. Devcontainer
	if err := generateDevcontainer(targetDir, cfg.ProjectName, techs, data); err != nil {
		return err
	}

	// 5. Conditional outputs
	if cfg.Licence == "mit" {
		if err := renderTemplateToFile(targetDir, "LICENSE", "licences/MIT.tmpl", data, 0o644); err != nil {
			return err
		}
	}

	for _, doc := range cfg.Docs {
		if doc == "contributing" {
			if err := renderTemplateToFile(targetDir, "CONTRIBUTING.md", "docs/CONTRIBUTING.md.tmpl", data, 0o644); err != nil {
				return err
			}
		}
	}

	for _, tool := range cfg.Tooling {
		switch tool {
		case "editorconfig":
			if err := renderTemplateToFile(targetDir, ".editorconfig", "editorconfig.tmpl", data, 0o644); err != nil {
				return err
			}
		case "gitattributes":
			if err := renderTemplateToFile(targetDir, ".gitattributes", "gitattributes.tmpl", data, 0o644); err != nil {
				return err
			}
		}
	}

	for _, rc := range cfg.RepoConfig {
		switch rc {
		case "issue_templates":
			if err := renderTemplateToFile(targetDir, ".github/ISSUE_TEMPLATE/bug_report.yaml", "github/bug_report.yaml.tmpl", data, 0o644); err != nil {
				return err
			}
			if err := renderTemplateToFile(targetDir, ".github/ISSUE_TEMPLATE/feature_request.yaml", "github/feature_request.yaml.tmpl", data, 0o644); err != nil {
				return err
			}
		case "pr_template":
			if err := renderTemplateToFile(targetDir, ".github/pull_request_template.md", "github/pull_request_template.md.tmpl", data, 0o644); err != nil {
				return err
			}
		case "dependabot":
			if err := renderTemplateToFile(targetDir, ".github/dependabot.yml", "github/dependabot.yml.tmpl", data, 0o644); err != nil {
				return err
			}
		}
	}

	return nil
}

// CreateStructure creates the directory/file structure defined in the technology's structure field.
func CreateStructure(targetDir string, structure []techdef.StructureEntry) error {
	return createStructureWithTemplate(targetDir, structure, map[string]string{})
}

func createStructureWithTemplate(targetDir string, structure []techdef.StructureEntry, templateData map[string]string) error {
	for _, entry := range structure {
		resolvedPath, err := ResolveTemplate(entry.Path, templateData)
		if err != nil {
			return fmt.Errorf("resolving path '%s': %w", entry.Path, err)
		}
		fullPath := filepath.Join(targetDir, resolvedPath)

		if strings.HasSuffix(resolvedPath, "/") {
			if err := os.MkdirAll(fullPath, 0o755); err != nil {
				return fmt.Errorf("creating directory '%s': %w", resolvedPath, err)
			}
			gitkeepPath := filepath.Join(fullPath, ".gitkeep")
			if err := os.WriteFile(gitkeepPath, []byte{}, 0o644); err != nil {
				return fmt.Errorf("creating .gitkeep in '%s': %w", resolvedPath, err)
			}
		} else {
			parentDir := filepath.Dir(fullPath)
			if err := os.MkdirAll(parentDir, 0o755); err != nil {
				return fmt.Errorf("creating parent directory for '%s': %w", resolvedPath, err)
			}
			content := []byte{}
			if entry.Content != nil {
				resolvedContent, err := ResolveTemplate(*entry.Content, templateData)
				if err != nil {
					return fmt.Errorf("resolving content for '%s': %w", resolvedPath, err)
				}
				content = []byte(resolvedContent)
			}
			if err := os.WriteFile(fullPath, content, 0o644); err != nil {
				return fmt.Errorf("creating file '%s': %w", resolvedPath, err)
			}
		}
	}
	return nil
}

// ComposeGitignore builds the composite .gitignore from all selected technologies.
// Technologies are expected to be sorted by key (alphabetical).
func ComposeGitignore(techs []*techdef.TechDef) string {
	var sb strings.Builder
	for i, tech := range techs {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString("# ")
		sb.WriteString(tech.Name)
		sb.WriteString("\n")
		sb.WriteString(tech.Gitignore)
	}
	return sb.String()
}

// RenderCIConfig composes a three-stage CI pipeline from all selected technologies.
func RenderCIConfig(techs []*techdef.TechDef) ([]byte, error) {
	// Collect technologies that contribute CI
	var ciTechs []*techdef.TechDef
	for _, tech := range techs {
		if tech.CI != nil {
			ciTechs = append(ciTechs, tech)
		}
	}

	if len(ciTechs) == 0 {
		return renderPlaceholderCI()
	}

	// Build job name lists for needs dependencies
	lintJobNames := make([]string, len(ciTechs))
	for i, tech := range ciTechs {
		lintJobNames[i] = "lint-" + tech.CI.JobName
	}
	testJobNames := make([]string, len(ciTechs))
	for i, tech := range ciTechs {
		testJobNames[i] = "test-" + tech.CI.JobName
	}

	var jobEntries []ciJobEntry

	for _, tech := range ciTechs {
		setupSteps := []interface{}{
			map[string]interface{}{"uses": "actions/checkout@v4"},
		}
		for _, s := range tech.CI.SetupSteps {
			step := map[string]interface{}{"name": s.Name}
			if s.Uses != "" {
				step["uses"] = s.Uses
				if len(s.With) > 0 {
					// Sort with keys for deterministic output
					with := make(map[string]string)
					for k, v := range s.With {
						with[k] = v
					}
					step["with"] = with
				}
			}
			if s.Run != "" {
				step["run"] = s.Run
			}
			setupSteps = append(setupSteps, step)
		}

		// Lint job
		lintSteps := make([]interface{}, len(setupSteps))
		copy(lintSteps, setupSteps)
		for _, s := range tech.CI.LintSteps {
			lintSteps = append(lintSteps, map[string]interface{}{
				"name": s.Name,
				"run":  s.Run,
			})
		}
		jobEntries = append(jobEntries, ciJobEntry{
			key: "lint-" + tech.CI.JobName,
			val: map[string]interface{}{
				"runs-on": "ubuntu-latest",
				"steps":   lintSteps,
			},
		})

		// Test job
		testSteps := make([]interface{}, len(setupSteps))
		copy(testSteps, setupSteps)
		for _, s := range tech.CI.TestSteps {
			testSteps = append(testSteps, map[string]interface{}{
				"name": s.Name,
				"run":  s.Run,
			})
		}
		jobEntries = append(jobEntries, ciJobEntry{
			key: "test-" + tech.CI.JobName,
			val: map[string]interface{}{
				"runs-on": "ubuntu-latest",
				"needs":   lintJobNames,
				"steps":   testSteps,
			},
		})
	}

	// Build job
	jobEntries = append(jobEntries, ciJobEntry{
		key: "build",
		val: map[string]interface{}{
			"runs-on": "ubuntu-latest",
			"needs":   testJobNames,
			"steps": []interface{}{
				map[string]interface{}{"uses": "actions/checkout@v4"},
				map[string]interface{}{"name": "Build", "run": "echo 'TODO: Add build steps'"},
			},
		},
	})

	// Build the YAML manually for deterministic output
	return renderCIYAML(jobEntries)
}

func renderPlaceholderCI() ([]byte, error) {
	content := `name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Build
        run: echo 'TODO: Add build steps'
`
	return []byte(content), nil
}

type ciJobEntry struct {
	key string
	val interface{}
}

// renderCIYAML produces GitHub Actions YAML from ordered job entries.
func renderCIYAML(jobEntries []ciJobEntry) ([]byte, error) {
	jobs := make(map[string]interface{})
	for _, je := range jobEntries {
		jobs[je.key] = je.val
	}

	workflow := map[string]interface{}{
		"name": "CI",
		"on": map[string]interface{}{
			"push":         map[string]interface{}{"branches": []string{"main"}},
			"pull_request": map[string]interface{}{"branches": []string{"main"}},
		},
		"jobs": jobs,
	}

	return yaml.Marshal(workflow)
}

type devcontainerJSON struct {
	Name              string                 `json:"name"`
	Build             map[string]string      `json:"build"`
	Features          map[string]interface{} `json:"features"`
	Customizations    map[string]interface{} `json:"customizations"`
	PostCreateCommand string                 `json:"postCreateCommand"`
}

func generateDevcontainer(targetDir, projectName string, techs []*techdef.TechDef, data templates.TemplateData) error {
	dcDir := filepath.Join(targetDir, ".devcontainer")
	if err := os.MkdirAll(dcDir, 0o755); err != nil {
		return fmt.Errorf("creating .devcontainer directory: %w", err)
	}

	// Dockerfile
	if err := renderTemplateToFile(targetDir, ".devcontainer/Dockerfile", "devcontainer/Dockerfile.tmpl", data, 0o644); err != nil {
		return err
	}

	// devcontainer.json
	dcJSON, err := renderDevcontainerJSON(projectName, techs)
	if err != nil {
		return fmt.Errorf("rendering devcontainer.json: %w", err)
	}
	dcJSON = append(dcJSON, '\n')
	if err := writeFile(filepath.Join(dcDir, "devcontainer.json"), dcJSON, 0o644); err != nil {
		return err
	}

	// setup.sh
	baseContent, err := templates.Render("devcontainer/setup/base.sh.tmpl", data)
	if err != nil {
		return fmt.Errorf("rendering base.sh.tmpl: %w", err)
	}
	setupContent := renderSetupSh(baseContent, techs)
	if err := writeFile(filepath.Join(dcDir, "setup.sh"), []byte(setupContent), 0o755); err != nil {
		return err
	}

	return nil
}

func renderDevcontainerJSON(projectName string, techs []*techdef.TechDef) ([]byte, error) {
	features := map[string]interface{}{
		"ghcr.io/devcontainers/features/node:1": map[string]interface{}{},
	}
	for _, tech := range techs {
		for k, v := range tech.Devcontainer.Features {
			features[k] = v
		}
	}

	// Merge and deduplicate extensions, then sort
	extSet := make(map[string]bool)
	for _, tech := range techs {
		for _, ext := range tech.Devcontainer.Extensions {
			extSet[ext] = true
		}
	}
	extensions := make([]string, 0, len(extSet))
	for ext := range extSet {
		extensions = append(extensions, ext)
	}
	sort.Strings(extensions)

	dc := devcontainerJSON{
		Name:  projectName,
		Build: map[string]string{"dockerfile": "Dockerfile"},
		Features: features,
		Customizations: map[string]interface{}{
			"vscode": map[string]interface{}{
				"extensions": extensions,
			},
		},
		PostCreateCommand: "bash .devcontainer/setup.sh",
	}

	return json.MarshalIndent(dc, "", "    ")
}

func renderSetupSh(baseTmpl string, techs []*techdef.TechDef) string {
	var sb strings.Builder
	sb.WriteString(baseTmpl)
	for _, tech := range techs {
		if strings.TrimSpace(tech.Devcontainer.Setup) != "" {
			sb.WriteString("\n# === ")
			sb.WriteString(tech.Name)
			sb.WriteString(" ===\n")
			sb.WriteString(tech.Devcontainer.Setup)
		}
	}
	return sb.String()
}

func renderTemplateToFile(targetDir, relPath, tmplName string, data templates.TemplateData, perm os.FileMode) error {
	content, err := templates.Render(tmplName, data)
	if err != nil {
		return fmt.Errorf("rendering %s: %w", tmplName, err)
	}
	fullPath := filepath.Join(targetDir, relPath)
	return writeFile(fullPath, []byte(content), perm)
}

func writeFile(path string, content []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating directory for %s: %w", path, err)
	}
	if err := os.WriteFile(path, content, perm); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}

