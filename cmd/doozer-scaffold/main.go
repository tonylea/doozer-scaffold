package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/tonylea/doozer-scaffold/internal/config"
	"github.com/tonylea/doozer-scaffold/internal/prompt"
	"github.com/tonylea/doozer-scaffold/internal/scaffold"
	"github.com/tonylea/doozer-scaffold/internal/techdef"
)

func main() {
	cfg := &config.Config{}

	// If a positional argument is provided, use it as the project name
	if len(os.Args) > 1 {
		cfg.ProjectName = os.Args[1]
	}

	// Load technology definitions
	techDefs, err := techdef.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading technology definitions: %v\n", err)
		os.Exit(1)
	}

	// Run interactive prompts
	if err := prompt.Run(cfg, techDefs); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if !cfg.Confirmed {
		fmt.Println("Scaffold generation cancelled.")
		os.Exit(0)
	}

	// Validate config
	if err := cfg.Validate(techDefs); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Resolve selected technology definitions, sorted by key
	sort.Strings(cfg.Technologies)
	techs := make([]*techdef.TechDef, 0, len(cfg.Technologies))
	for _, key := range cfg.Technologies {
		techs = append(techs, techDefs[key])
	}

	if err := scaffold.Generate(cfg, techs, "."); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating scaffold: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Project '%s' scaffolded successfully.\n", cfg.ProjectName)
}
