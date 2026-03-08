package prompt

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/huh"
	"github.com/tonylea/doozer-scaffold/internal/config"
	"github.com/tonylea/doozer-scaffold/internal/techdef"
)

// Run presents the interactive prompt flow and populates cfg with user selections.
func Run(cfg *config.Config, techDefs map[string]*techdef.TechDef) error {
	groups := []*huh.Group{}

	if cfg.ProjectName == "" {
		groups = append(groups, huh.NewGroup(
			huh.NewInput().
				Title("Project name:").
				Value(&cfg.ProjectName).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("project name is required")
					}
					return nil
				}),
		))
	}

	techOptions := buildTechOptions(techDefs)

	groups = append(groups,
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Remote hosting provider:").
				Options(
					huh.NewOption("GitHub", "github"),
				).
				Value(&cfg.Provider),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Technology:").
				Options(techOptions...).
				Value(&cfg.Technology),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Licence:").
				Options(
					huh.NewOption("MIT", "mit"),
					huh.NewOption("None", "none"),
				).
				Value(&cfg.Licence),
		),
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Supplementary documentation:").
				Options(
					huh.NewOption("CONTRIBUTING.md", "contributing"),
				).
				Value(&cfg.Docs),
		),
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Project tooling:").
				Options(
					huh.NewOption(".editorconfig", "editorconfig"),
					huh.NewOption(".gitattributes", "gitattributes"),
				).
				Value(&cfg.Tooling),
		),
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("GitHub repository configuration:").
				Options(
					huh.NewOption("Issue templates", "issue_templates"),
					huh.NewOption("Pull request template", "pr_template"),
					huh.NewOption("Dependabot configuration", "dependabot"),
				).
				Value(&cfg.RepoConfig),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Generate scaffold with these selections?").
				Affirmative("Yes").
				Negative("No").
				Value(&cfg.Confirmed),
		),
	)

	form := huh.NewForm(groups...)
	return form.Run()
}

// buildTechOptions creates huh.Option entries from loaded technology definitions.
// Options are sorted alphabetically by display name for consistent presentation.
func buildTechOptions(techDefs map[string]*techdef.TechDef) []huh.Option[string] {
	type techEntry struct {
		key  string
		name string
	}

	entries := make([]techEntry, 0, len(techDefs))
	for key, def := range techDefs {
		entries = append(entries, techEntry{key: key, name: def.Name})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].name < entries[j].name
	})

	options := make([]huh.Option[string], 0, len(entries))
	for _, e := range entries {
		options = append(options, huh.NewOption(e.name, e.key))
	}
	return options
}
