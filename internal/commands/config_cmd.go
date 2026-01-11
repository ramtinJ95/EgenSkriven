package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pocketbase/pocketbase"
	"github.com/spf13/cobra"

	"github.com/ramtinJ95/EgenSkriven/internal/config"
)

func newConfigCmd(app *pocketbase.PocketBase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long: `Manage EgenSkriven configuration.

Configuration is loaded from two locations:
- Global: ~/.config/egenskriven/config.json (user-wide settings)
- Project: .egenskriven/config.json (project-specific overrides)

Project settings override global settings for:
- agent.workflow, agent.mode, agent.resume_mode
- server.url

Global-only settings (cannot be overridden per-project):
- data_dir (database location)
- defaults.author, defaults.agent`,
	}

	cmd.AddCommand(newConfigShowCmd())
	cmd.AddCommand(newConfigPathCmd())

	return cmd
}

func newConfigShowCmd() *cobra.Command {
	var showGlobal, showProject bool

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show configuration",
		Long: `Show configuration.

By default, shows the merged configuration (global + project).
Use --global or --project to show only that config.`,
		Example: `  # Show effective (merged) configuration
  egenskriven config show

  # Show global configuration only
  egenskriven config show --global

  # Show project configuration only
  egenskriven config show --project`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")

			if showGlobal {
				cfg, err := config.LoadGlobalConfig()
				if err != nil {
					return out.Error(ExitGeneralError, fmt.Sprintf("failed to load global config: %v", err), nil)
				}
				return encoder.Encode(cfg)
			}

			if showProject {
				cfg, err := config.LoadProjectConfig()
				if err != nil {
					return out.Error(ExitGeneralError, fmt.Sprintf("failed to load project config: %v", err), nil)
				}
				return encoder.Encode(cfg)
			}

			// Default: show merged config
			cfg, err := config.Load()
			if err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to load config: %v", err), nil)
			}
			return encoder.Encode(cfg)
		},
	}

	cmd.Flags().BoolVar(&showGlobal, "global", false, "Show global config only")
	cmd.Flags().BoolVar(&showProject, "project", false, "Show project config only")

	return cmd
}

func newConfigPathCmd() *cobra.Command {
	var showGlobal bool

	cmd := &cobra.Command{
		Use:   "path",
		Short: "Show config file path",
		Long: `Show the path to the configuration file.

By default, shows the project config path.
Use --global to show the global config path.`,
		Example: `  # Show project config path
  egenskriven config path

  # Show global config path
  egenskriven config path --global`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if showGlobal {
				path, err := config.GlobalConfigPath()
				if err != nil {
					return fmt.Errorf("failed to get global config path: %w", err)
				}
				fmt.Println(path)
				return nil
			}
			fmt.Println(".egenskriven/config.json")
			return nil
		},
	}

	cmd.Flags().BoolVar(&showGlobal, "global", false, "Show global config path")

	return cmd
}
