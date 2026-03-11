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

// FilterPromptsByMode returns only the prompts that should be shown for the given
// resolved variant mode. prompts with no mode are always shown. prompts with a
// mode ("standalone" or "composable") are only shown when the mode matches.
// resolvedMode is "" for non-variant-group techs, "standalone", or "composable".
func FilterPromptsByMode(prompts []techdef.PromptDef, resolvedMode string) []techdef.PromptDef {
	var result []techdef.PromptDef
	for _, p := range prompts {
		if p.Mode == "" {
			result = append(result, p)
		} else if p.Mode == resolvedMode {
			result = append(result, p)
		}
	}
	return result
}

// Run presents the interactive prompt flow and populates cfg with user selections.
// Phase 1: project name, provider, technology multi-select.
// Phase 2: tech-driven prompts (with variant group resolution), licence, docs, tooling, repo config, confirm.
func Run(cfg *config.Config, techDefs map[string]*techdef.TechDef) error {
	phase1Groups := buildPhase1Groups(cfg, techDefs)
	if err := huh.NewForm(phase1Groups...).Run(); err != nil {
		return err
	}

	// Resolve variant group selections to actual defs and determine mode
	resolvedTechs, modeMap := techdef.ResolveVariantGroups(cfg.Technologies, techDefs)

	techPromptGroups := buildTechPromptGroupsResolved(cfg, resolvedTechs, modeMap)
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

// buildTechPromptGroupsResolved builds prompt groups for resolved technology definitions,
// filtering prompts by mode (only showing composable prompts when composable was resolved, etc.).
// modeMap maps variant group names to their resolved mode ("standalone" or "composable").
func buildTechPromptGroupsResolved(cfg *config.Config, resolvedTechs []*techdef.TechDef, modeMap map[string]string) []*huh.Group {
	if cfg.TechPromptResponses == nil {
		cfg.TechPromptResponses = make(map[string]string)
	}

	// Sort by name for deterministic order
	sorted := make([]*techdef.TechDef, len(resolvedTechs))
	copy(sorted, resolvedTechs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})

	var groups []*huh.Group
	for _, def := range sorted {
		// Determine the resolved mode for this tech
		resolvedMode := ""
		if def.VariantGroup != "" {
			resolvedMode = modeMap[def.VariantGroup]
		}

		filtered := FilterPromptsByMode(def.Prompts, resolvedMode)
		for _, p := range filtered {
			promptKey := p.Key
			switch p.Type {
			case "text":
				defaultVal := ""
				if p.DefaultFrom == "project_name" {
					defaultVal = SanitiseForIdentifier(cfg.ProjectName)
				}
				cfg.TechPromptResponses[promptKey] = defaultVal
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
