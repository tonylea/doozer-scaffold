package prompt

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/charmbracelet/huh"
	"github.com/tonylea/doozer-scaffold/internal/config"
	"github.com/tonylea/doozer-scaffold/internal/techdef"
)

// SanitiseForIdentifier derives a valid identifier from a project name.
// Rules: replace hyphens with underscores, replace non-alphanumeric/underscore
// chars with underscore, strip leading digits/underscores, lowercase the result.
// Falls back to "app" if the result is empty.
func SanitiseForIdentifier(projectName string) string {
	result := strings.ReplaceAll(projectName, "-", "_")
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			return r
		}
		return '_'
	}, result)
	cleaned = strings.TrimLeftFunc(cleaned, func(r rune) bool {
		return !unicode.IsLetter(r)
	})
	if cleaned == "" {
		return "app"
	}
	return strings.ToLower(cleaned)
}

// Run presents the interactive prompt flow and populates cfg with user selections.
// Phase 1: project name, provider, technology multi-select.
// Phase 2: tech-driven prompts, licence, docs, tooling, repo config, confirm.
func Run(cfg *config.Config, techDefs map[string]*techdef.TechDef) error {
	phase1Groups := buildPhase1Groups(cfg, techDefs)
	if err := huh.NewForm(phase1Groups...).Run(); err != nil {
		return err
	}

	techPromptGroups := buildTechPromptGroups(cfg, techDefs)
	phase2Groups := append(techPromptGroups, buildPhase2Groups(cfg)...)
	return huh.NewForm(phase2Groups...).Run()
}

func buildPhase1Groups(cfg *config.Config, techDefs map[string]*techdef.TechDef) []*huh.Group {
	var groups []*huh.Group

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

	variantGroups := techdef.BuildVariantGroups(techDefs)

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
			huh.NewMultiSelect[string]().
				Title("Technologies:").
				Options(buildTechOptions(techDefs)...).
				Value(&cfg.Technologies).
				Validate(func(selected []string) error {
					if len(selected) == 0 {
						return fmt.Errorf("at least one technology must be selected")
					}
					if len(selected) > 1 {
						for _, key := range selected {
							if _, isGroup := variantGroups[key]; isGroup {
								continue
							}
							if def, ok := techDefs[key]; ok && def.Standalone && def.VariantGroup == "" {
								return fmt.Errorf("'%s' is a standalone technology and cannot be combined with others", def.Name)
							}
						}
					}
					return nil
				}),
		),
	)

	return groups
}

func buildTechPromptGroups(cfg *config.Config, techDefs map[string]*techdef.TechDef) []*huh.Group {
	if cfg.TechPromptResponses == nil {
		cfg.TechPromptResponses = make(map[string]string)
	}

	keys := make([]string, len(cfg.Technologies))
	copy(keys, cfg.Technologies)
	sort.Strings(keys)

	var groups []*huh.Group
	for _, key := range keys {
		def := techDefs[key]
		for _, p := range def.Prompts {
			promptKey := p.Key
			switch p.Type {
			case "text":
				defaultVal := ""
				if p.DefaultFrom == "project_name" {
					defaultVal = SanitiseForIdentifier(cfg.ProjectName)
				}
				cfg.TechPromptResponses[promptKey] = defaultVal
				// Use a pointer to a local that writes back to map
				localVal := defaultVal
				localKey := promptKey
				title := p.Title
				groups = append(groups, huh.NewGroup(
					huh.NewInput().
						Title(title).
						Value(&localVal).
						Validate(func(s string) error {
							if s == "" {
								return fmt.Errorf("%s is required", title)
							}
							cfg.TechPromptResponses[localKey] = s
							return nil
						}),
				))
			case "select":
				options := make([]huh.Option[string], len(p.Options))
				for i, o := range p.Options {
					options[i] = huh.NewOption(o.Label, o.Value)
				}
				localVal := cfg.TechPromptResponses[promptKey]
				localKey := promptKey
				groups = append(groups, huh.NewGroup(
					huh.NewSelect[string]().
						Title(p.Title).
						Options(options...).
						Value(&localVal),
				))
				cfg.TechPromptResponses[localKey] = localVal
			case "multi_select":
				var selected []string
				options := make([]huh.Option[string], len(p.Options))
				for i, o := range p.Options {
					options[i] = huh.NewOption(o.Label, o.Value)
				}
				groups = append(groups, huh.NewGroup(
					huh.NewMultiSelect[string]().
						Title(p.Title).
						Options(options...).
						Value(&selected),
				))
			}
		}
	}
	return groups
}

func buildPhase2Groups(cfg *config.Config) []*huh.Group {
	return []*huh.Group{
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
	}
}

// TechOption represents a single selectable technology option in the prompt.
// For variant groups, Key is the group name and Name is the group name.
// For regular technologies, Key is the definition key and Name is the definition name.
type TechOption struct {
	Key  string
	Name string
}

// BuildTechOptionList builds an ordered list of TechOption entries from loaded technology
// definitions, collapsing variant groups into single entries.
// Options are sorted alphabetically by display name.
func BuildTechOptionList(techDefs map[string]*techdef.TechDef) []TechOption {
	variantGroups := techdef.BuildVariantGroups(techDefs)

	// Track which keys have been represented by a variant group
	coveredKeys := make(map[string]bool)
	for _, def := range techDefs {
		if def.VariantGroup != "" {
			coveredKeys[def.VariantGroup] = true // mark group as seen
		}
	}

	var entries []TechOption

	// Add one entry per variant group
	for groupName := range variantGroups {
		entries = append(entries, TechOption{Key: groupName, Name: groupName})
	}

	// Add regular (non-variant-group) technologies
	for key, def := range techDefs {
		if def.VariantGroup != "" {
			continue // skip variant group members
		}
		entries = append(entries, TechOption{Key: key, Name: def.Name})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	return entries
}

// buildTechOptions creates huh.Option entries from loaded technology definitions.
// Options are sorted alphabetically by display name for consistent presentation.
// Variant groups are collapsed into a single entry using the group name.
func buildTechOptions(techDefs map[string]*techdef.TechDef) []huh.Option[string] {
	entries := BuildTechOptionList(techDefs)
	options := make([]huh.Option[string], 0, len(entries))
	for _, e := range entries {
		options = append(options, huh.NewOption(e.Name, e.Key))
	}
	return options
}
