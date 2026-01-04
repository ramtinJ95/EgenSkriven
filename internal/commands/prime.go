package commands

import (
	_ "embed"
	"fmt"
	"os"
	"text/template"

	"github.com/pocketbase/pocketbase"
	"github.com/spf13/cobra"

	"github.com/ramtinJ95/EgenSkriven/internal/config"
)

//go:embed prime.tmpl
var primeTemplate string

// PrimeTemplateData holds data for the prime template.
type PrimeTemplateData struct {
	WorkflowMode       string
	AgentMode          string
	AgentName          string
	OverrideTodoWrite  bool
	RequireSummary     bool
	StructuredSections bool
}

func newPrimeCmd(app *pocketbase.PocketBase) *cobra.Command {
	var (
		workflowOverride string
		agentName        string
	)

	cmd := &cobra.Command{
		Use:   "prime",
		Short: "Output instructions for AI coding agents",
		Long: `Output instructions that teach AI agents how to use EgenSkriven.

Reads configuration from .egenskriven/config.json in the current project.
Use --workflow to override the workflow mode.

This command is typically called automatically via agent hooks (Claude, OpenCode)
rather than manually. The output is designed to be injected into agent context.

Examples:
  egenskriven prime
  egenskriven prime --workflow strict
  egenskriven prime --agent claude`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load project configuration
			cfg, err := config.LoadProjectConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
				cfg = config.DefaultConfig()
			}

			// Override workflow if specified
			workflowMode := cfg.Agent.Workflow
			if workflowOverride != "" {
				workflowMode = workflowOverride
			}

			// Set default agent name if not provided
			if agentName == "" {
				agentName = "agent"
			}

			// Parse and execute template
			tmpl, err := template.New("prime").Parse(primeTemplate)
			if err != nil {
				return fmt.Errorf("failed to parse template: %w", err)
			}

			data := PrimeTemplateData{
				WorkflowMode:       workflowMode,
				AgentMode:          cfg.Agent.Mode,
				AgentName:          agentName,
				OverrideTodoWrite:  cfg.Agent.OverrideTodoWrite,
				RequireSummary:     cfg.Agent.RequireSummary,
				StructuredSections: cfg.Agent.StructuredSections,
			}

			return tmpl.Execute(os.Stdout, data)
		},
	}

	cmd.Flags().StringVar(&workflowOverride, "workflow", "",
		"Override workflow mode (strict, light, minimal)")
	cmd.Flags().StringVar(&agentName, "agent", "",
		"Agent identifier for --agent flag in examples")

	return cmd
}
