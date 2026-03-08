package scaffold

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tonylea/doozer-scaffold/internal/config"
	"github.com/tonylea/doozer-scaffold/internal/techdef"
	"github.com/tonylea/doozer-scaffold/internal/templates"
)

// Generate creates the scaffold in a subdirectory of baseDir named after cfg.ProjectName.
func Generate(cfg *config.Config, tech *techdef.TechDef, baseDir string) error {
	targetDir := filepath.Join(baseDir, cfg.ProjectName)

	// Check target doesn't already exist
	if _, err := os.Stat(targetDir); err == nil {
		return fmt.Errorf("directory '%s' already exists", cfg.ProjectName)
	}

	// Create directory structure and files; on any error, clean up targetDir
	if err := createScaffold(cfg, tech, targetDir); err != nil {
		_ = os.RemoveAll(targetDir)
		return fmt.Errorf("scaffold generation failed: %w", err)
	}

	return nil
}

func createScaffold(cfg *config.Config, tech *techdef.TechDef, targetDir string) error {
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

	gitignore := ComposeGitignore(tech)
	if err := writeFile(filepath.Join(targetDir, ".gitignore"), []byte(gitignore), 0o644); err != nil {
		return err
	}

	if err := renderTemplateToFile(targetDir, ".github/workflows/ci.yml", "github/ci.yml.tmpl", data, 0o644); err != nil {
		return err
	}

	// 3. Technology structure
	if err := CreateStructure(targetDir, tech.Structure); err != nil {
		return err
	}

	// 4. Devcontainer
	if err := generateDevcontainer(targetDir, cfg.ProjectName, tech, data); err != nil {
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
	for _, entry := range structure {
		fullPath := filepath.Join(targetDir, entry.Path)

		if entry.IsDir() {
			if err := os.MkdirAll(fullPath, 0o755); err != nil {
				return fmt.Errorf("creating directory '%s': %w", entry.Path, err)
			}
			gitkeepPath := filepath.Join(fullPath, ".gitkeep")
			if err := os.WriteFile(gitkeepPath, []byte{}, 0o644); err != nil {
				return fmt.Errorf("creating .gitkeep in '%s': %w", entry.Path, err)
			}
		} else {
			parentDir := filepath.Dir(fullPath)
			if err := os.MkdirAll(parentDir, 0o755); err != nil {
				return fmt.Errorf("creating parent directory for '%s': %w", entry.Path, err)
			}
			content := []byte{}
			if entry.Content != nil {
				content = []byte(*entry.Content)
			}
			if err := os.WriteFile(fullPath, content, 0o644); err != nil {
				return fmt.Errorf("creating file '%s': %w", entry.Path, err)
			}
		}
	}
	return nil
}

// ComposeGitignore builds the .gitignore content by prepending a header comment to the technology's gitignore lines.
func ComposeGitignore(tech *techdef.TechDef) string {
	var sb strings.Builder
	sb.WriteString("# ")
	sb.WriteString(tech.Name)
	sb.WriteString("\n")
	sb.WriteString(tech.Gitignore)
	return sb.String()
}

func generateDevcontainer(targetDir, projectName string, tech *techdef.TechDef, data templates.TemplateData) error {
	dcDir := filepath.Join(targetDir, ".devcontainer")
	if err := os.MkdirAll(dcDir, 0o755); err != nil {
		return fmt.Errorf("creating .devcontainer directory: %w", err)
	}

	// Dockerfile
	if err := renderTemplateToFile(targetDir, ".devcontainer/Dockerfile", "devcontainer/Dockerfile.tmpl", data, 0o644); err != nil {
		return err
	}

	// devcontainer.json — rendered programmatically
	dcJSON, err := renderDevcontainerJSON(projectName, tech)
	if err != nil {
		return fmt.Errorf("rendering devcontainer.json: %w", err)
	}
	// Append trailing newline
	dcJSON = append(dcJSON, '\n')
	if err := writeFile(filepath.Join(dcDir, "devcontainer.json"), dcJSON, 0o644); err != nil {
		return err
	}

	// setup.sh — base template + technology setup block
	baseContent, err := templates.Render("devcontainer/setup/base.sh.tmpl", data)
	if err != nil {
		return fmt.Errorf("rendering base.sh.tmpl: %w", err)
	}
	setupContent := renderSetupSh(baseContent, tech)
	if err := writeFile(filepath.Join(dcDir, "setup.sh"), []byte(setupContent), 0o755); err != nil {
		return err
	}

	return nil
}

type devcontainerJSON struct {
	Name               string                 `json:"name"`
	Build              map[string]string      `json:"build"`
	Features           map[string]interface{} `json:"features"`
	Customizations     map[string]interface{} `json:"customizations"`
	PostCreateCommand  string                 `json:"postCreateCommand"`
}

func renderDevcontainerJSON(projectName string, tech *techdef.TechDef) ([]byte, error) {
	features := map[string]interface{}{
		"ghcr.io/devcontainers/features/node:1": map[string]interface{}{},
	}
	for k, v := range tech.Devcontainer.Features {
		features[k] = v
	}

	dc := devcontainerJSON{
		Name:  projectName,
		Build: map[string]string{"dockerfile": "Dockerfile"},
		Features: features,
		Customizations: map[string]interface{}{
			"vscode": map[string]interface{}{
				"extensions": tech.Devcontainer.Extensions,
			},
		},
		PostCreateCommand: "bash .devcontainer/setup.sh",
	}

	return json.MarshalIndent(dc, "", "    ")
}

func renderSetupSh(baseTmpl string, tech *techdef.TechDef) string {
	var sb strings.Builder
	sb.WriteString(baseTmpl)
	if strings.TrimSpace(tech.Devcontainer.Setup) != "" {
		sb.WriteString("\n# === ")
		sb.WriteString(tech.Name)
		sb.WriteString(" ===\n")
		sb.WriteString(tech.Devcontainer.Setup)
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
