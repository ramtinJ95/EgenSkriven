package commands

import (
	"fmt"

	"github.com/pocketbase/pocketbase"
	"github.com/spf13/cobra"

	"github.com/ramtinJ95/EgenSkriven/internal/config"
)

func newInitCmd(app *pocketbase.PocketBase) *cobra.Command {
	var (
		workflow   string
		mode       string
		opencode   bool
		claudeCode bool
		codex      bool
		all        bool
		force      bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize EgenSkriven configuration",
		Long: `Initialize EgenSkriven configuration for the current project.

Creates .egenskriven/config.json with agent workflow settings.

Workflow modes:
  strict   - Full enforcement: create before work, update during, summary after
  light    - Basic tracking: create/complete tasks, no structured sections
  minimal  - No enforcement: agent decides when to use

Agent modes:
  autonomous    - Agent executes actions directly, human reviews async
  collaborative - Agent proposes major changes, executes minor ones
  supervised    - Agent is read-only, outputs commands for human

Tool integrations:
  --opencode     - Generate OpenCode custom tool for session discovery
  --claude-code  - Generate Claude Code hooks for session discovery
  --codex        - Generate Codex helper script for session discovery
  --all          - Generate all tool integrations

Examples:
  egenskriven init
  egenskriven init --workflow strict
  egenskriven init --workflow light --mode collaborative
  egenskriven init --opencode
  egenskriven init --claude-code --codex
  egenskriven init --all
  egenskriven init --all --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			// Load existing config or create new
			cfg := config.DefaultConfig()

			// Override with flags if provided, validating values
			if workflow != "" {
				if err := config.ValidateWorkflow(workflow); err != nil {
					return out.Error(ExitValidation, err.Error(), nil)
				}
				cfg.Agent.Workflow = workflow
			}
			if mode != "" {
				if err := config.ValidateMode(mode); err != nil {
					return out.Error(ExitValidation, err.Error(), nil)
				}
				cfg.Agent.Mode = mode
			}

			// Save configuration
			if err := config.SaveConfig(".", cfg); err != nil {
				return out.Error(ExitGeneralError,
					fmt.Sprintf("failed to save config: %v", err), nil)
			}

			if !jsonOutput {
				fmt.Println("Created .egenskriven/config.json")
				fmt.Printf("  Workflow: %s\n", cfg.Agent.Workflow)
				fmt.Printf("  Mode: %s\n", cfg.Agent.Mode)
			}

			// Handle tool integrations
			if all {
				opencode, claudeCode, codex = true, true, true
			}

			generated := []string{}

			if opencode {
				files, err := generateOpenCodeIntegration(force)
				if err != nil {
					return out.Error(ExitGeneralError,
						fmt.Sprintf("failed to generate OpenCode integration: %v", err), nil)
				}
				generated = append(generated, files...)
			}

			if claudeCode {
				files, err := generateClaudeCodeIntegration(force)
				if err != nil {
					return out.Error(ExitGeneralError,
						fmt.Sprintf("failed to generate Claude Code integration: %v", err), nil)
				}
				generated = append(generated, files...)
			}

			if codex {
				files, err := generateCodexIntegration(force)
				if err != nil {
					return out.Error(ExitGeneralError,
						fmt.Sprintf("failed to generate Codex integration: %v", err), nil)
				}
				generated = append(generated, files...)
			}

			// Output results
			if !jsonOutput {
				if len(generated) > 0 {
					fmt.Println()
					fmt.Println("Generated tool integrations:")
					for _, f := range generated {
						fmt.Printf("  - %s\n", f)
					}
					fmt.Println()
					fmt.Println("See each file for usage instructions.")
				}

				fmt.Println()
				fmt.Println("Next steps:")
				fmt.Println("  1. Edit .egenskriven/config.json to customize settings")
				fmt.Println("  2. Run 'egenskriven prime' to see agent instructions")
				fmt.Println("  3. Configure your AI agent to run 'egenskriven prime' on session start")
			} else {
				out.WriteJSON(map[string]any{
					"config_created":  true,
					"generated_files": generated,
					"workflow":        cfg.Agent.Workflow,
					"mode":            cfg.Agent.Mode,
				})
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&workflow, "workflow", "",
		"Workflow mode (strict, light, minimal)")
	cmd.Flags().StringVar(&mode, "mode", "",
		"Agent mode (autonomous, collaborative, supervised)")
	cmd.Flags().BoolVar(&opencode, "opencode", false,
		"Generate OpenCode custom tool for session discovery")
	cmd.Flags().BoolVar(&claudeCode, "claude-code", false,
		"Generate Claude Code hooks for session discovery")
	cmd.Flags().BoolVar(&codex, "codex", false,
		"Generate Codex helper script for session discovery")
	cmd.Flags().BoolVar(&all, "all", false,
		"Generate all tool integrations")
	cmd.Flags().BoolVarP(&force, "force", "f", false,
		"Overwrite existing integration files")

	return cmd
}
