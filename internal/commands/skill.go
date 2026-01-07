package commands

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/spf13/cobra"
)

// embeddedSkills contains the skill files that are embedded into the binary.
// The source of truth for skill content is internal/commands/skills/*.
// The .claude/skills/ directory in the repo is for local development/testing
// but the embedded files here are what gets distributed to users.
//
//go:embed skills/*
var embeddedSkills embed.FS

// SkillInfo represents metadata about an installed skill
type SkillInfo struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Installed   bool   `json:"installed"`
	Description string `json:"description,omitempty"`
}

// newSkillCmd creates the skill command and its subcommands
func newSkillCmd(app *pocketbase.PocketBase) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skill",
		Short: "Manage AI agent skills",
		Long: `Manage EgenSkriven skills for AI coding agents.

Skills are instruction files that teach AI agents (Claude Code, OpenCode, Cursor, etc.)
how to use EgenSkriven. They are installed to agent-specific locations.

Available skills:
  - egenskriven           Core commands and task management
  - egenskriven-workflows Workflow modes and agent behaviors
  - egenskriven-advanced  Epics, dependencies, sub-tasks, batch operations`,
		Example: `  egenskriven skill install          # Interactive installation
  egenskriven skill install --global  # Install to ~/.claude/skills/
  egenskriven skill install --project # Install to .claude/skills/
  egenskriven skill uninstall
  egenskriven skill status`,
	}

	cmd.AddCommand(newSkillInstallCmd())
	cmd.AddCommand(newSkillUninstallCmd())
	cmd.AddCommand(newSkillStatusCmd())

	return cmd
}

// getGlobalSkillPath returns the global skills directory path for Claude Code
func getGlobalSkillPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(home, ".claude", "skills"), nil
}

// getGlobalOpenCodeSkillPath returns the global skills directory path for OpenCode
func getGlobalOpenCodeSkillPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(home, ".config", "opencode", "skill"), nil
}

// getProjectSkillPath returns the project skills directory path for Claude Code
func getProjectSkillPath() string {
	return filepath.Join(".claude", "skills")
}

// getProjectOpenCodeSkillPath returns the project skills directory path for OpenCode
func getProjectOpenCodeSkillPath() string {
	return filepath.Join(".opencode", "skill")
}

// skillNames is the list of skills to install
var skillNames = []string{
	"egenskriven",
	"egenskriven-workflows",
	"egenskriven-advanced",
}

// newSkillInstallCmd creates the 'skill install' subcommand
func newSkillInstallCmd() *cobra.Command {
	var (
		global  bool
		project bool
		force   bool
	)

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install EgenSkriven skills for AI agents",
		Long: `Install EgenSkriven skills to make them available to AI coding agents.

Skills can be installed globally (available in all projects) or locally
(only in the current project). Global installation is recommended for
personal use; project installation is better for shared repositories.

Installation locations (installs to both Claude Code and OpenCode directories):
  Global:  ~/.claude/skills/ and ~/.config/opencode/skill/
  Project: .claude/skills/ and .opencode/skill/

Note: In JSON mode (--json) without --global or --project, defaults to global.`,
		Example: `  egenskriven skill install          # Interactive prompt
  egenskriven skill install --global  # Install globally
  egenskriven skill install --project # Install to current project
  egenskriven skill install --force   # Overwrite existing skills
  egenskriven skill install --json    # JSON output (defaults to global)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			if global && project {
				return fmt.Errorf("cannot use both --global and --project")
			}

			if !global && !project {
				// Interactive prompt
				if !out.JSON {
					fmt.Println("EgenSkriven Skill Installation")
					fmt.Println()
					fmt.Println("Where would you like to install the skills?")
					fmt.Println()
					fmt.Println("  1. Global (~/.claude/skills/) - Available in all projects")
					fmt.Println("  2. Project (.claude/skills/) - Only this project")
					fmt.Println()
					fmt.Print("Enter choice [1/2]: ")

					var choice string
					fmt.Scanln(&choice)

					switch strings.TrimSpace(choice) {
					case "1":
						global = true
					case "2":
						project = true
					default:
						return fmt.Errorf("invalid choice: %s", choice)
					}
				} else {
					// In JSON mode without flags, default to global
					global = true
				}
			}

			// Collect installation paths for both Claude Code and OpenCode
			var installPaths []string
			if global {
				claudePath, err := getGlobalSkillPath()
				if err != nil {
					return err
				}
				installPaths = append(installPaths, claudePath)

				opencodePath, err := getGlobalOpenCodeSkillPath()
				if err != nil {
					return err
				}
				installPaths = append(installPaths, opencodePath)
			} else {
				installPaths = append(installPaths, getProjectSkillPath())
				installPaths = append(installPaths, getProjectOpenCodeSkillPath())
			}

			// Install each skill to all target paths
			installed := make([]string, 0, len(skillNames)*len(installPaths))
			for _, installPath := range installPaths {
				if !out.JSON {
					fmt.Printf("Installing skills to %s...\n", installPath)
				}

				for _, skillName := range skillNames {
					skillDir := filepath.Join(installPath, skillName)
					skillFile := filepath.Join(skillDir, "SKILL.md")

					// Check if skill already exists
					if _, err := os.Stat(skillFile); err == nil && !force {
						if !out.JSON {
							fmt.Printf("  Skipped: %s (already exists, use --force to overwrite)\n", skillName)
						}
						continue
					}

					// Create skill directory
					if err := os.MkdirAll(skillDir, 0755); err != nil {
						return fmt.Errorf("failed to create directory %s: %w", skillDir, err)
					}

					// Read embedded skill content
					content, err := embeddedSkills.ReadFile(filepath.Join("skills", skillName, "SKILL.md"))
					if err != nil {
						return fmt.Errorf("failed to read embedded skill %s: %w", skillName, err)
					}

					// Write skill file
					if err := os.WriteFile(skillFile, content, 0644); err != nil {
						return fmt.Errorf("failed to write skill file %s: %w", skillFile, err)
					}

					installed = append(installed, skillFile)
					if !out.JSON {
						fmt.Printf("  Created: %s\n", skillFile)
					}
				}
				if !out.JSON {
					fmt.Println()
				}
			}

			if out.JSON {
				location := "global"
				if project {
					location = "project"
				}
				return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
					"installed": installed,
					"location":  location,
					"paths":     installPaths,
					"count":     len(installed),
				})
			}

			if len(installed) > 0 {
				fmt.Println("\nSkills installed successfully!")
				fmt.Println()
				fmt.Println("Next steps:")
				fmt.Println("  1. Restart your AI agent (Claude Code, OpenCode, etc.)")
				fmt.Println("  2. The agent will automatically discover the new skills")
				fmt.Println("  3. Skills load on-demand when relevant to your task")
				fmt.Println()
				fmt.Println("To uninstall: egenskriven skill uninstall")
				fmt.Println("To update:    egenskriven skill install --force")
			} else {
				fmt.Println("\nNo skills were installed. All skills already exist.")
				fmt.Println("Use --force to overwrite existing skills.")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&global, "global", false, "Install to ~/.claude/skills/ (all projects)")
	cmd.Flags().BoolVar(&project, "project", false, "Install to .claude/skills/ (this project only)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing skills")

	return cmd
}

// newSkillUninstallCmd creates the 'skill uninstall' subcommand
func newSkillUninstallCmd() *cobra.Command {
	var (
		global  bool
		project bool
	)

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove installed EgenSkriven skills",
		Long: `Remove EgenSkriven skills from the specified location.

By default, removes skills from both global and project locations if they exist.
Use --global or --project to target a specific location.`,
		Example: `  egenskriven skill uninstall          # Remove from all locations
  egenskriven skill uninstall --global  # Remove from ~/.claude/skills/ only
  egenskriven skill uninstall --project # Remove from .claude/skills/ only`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			removed := make([]string, 0)

			// If neither flag is set, try both locations
			if !global && !project {
				global = true
				project = true
			}

			// Helper to remove skills from a path
			removeFromPath := func(basePath string) {
				for _, skillName := range skillNames {
					skillDir := filepath.Join(basePath, skillName)
					if _, err := os.Stat(skillDir); err == nil {
						if err := os.RemoveAll(skillDir); err != nil {
							// Log warning but continue with other skills
							warnLog("failed to remove %s: %v", skillDir, err)
						} else {
							removed = append(removed, skillDir)
							if !out.JSON {
								fmt.Printf("Removed: %s\n", skillDir)
							}
						}
					}
				}
			}

			if global {
				// Claude Code global
				if globalPath, err := getGlobalSkillPath(); err == nil {
					removeFromPath(globalPath)
				}
				// OpenCode global
				if opencodePath, err := getGlobalOpenCodeSkillPath(); err == nil {
					removeFromPath(opencodePath)
				}
			}

			if project {
				// Claude Code project
				removeFromPath(getProjectSkillPath())
				// OpenCode project
				removeFromPath(getProjectOpenCodeSkillPath())
			}

			if out.JSON {
				return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
					"removed": removed,
					"count":   len(removed),
				})
			}

			if len(removed) == 0 {
				fmt.Println("No skills found to uninstall.")
			} else {
				fmt.Printf("\nUninstalled %d skill(s).\n", len(removed))
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&global, "global", false, "Remove from ~/.claude/skills/ only")
	cmd.Flags().BoolVar(&project, "project", false, "Remove from .claude/skills/ only")

	return cmd
}

// newSkillStatusCmd creates the 'skill status' subcommand
func newSkillStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show installation status of EgenSkriven skills",
		Long: `Show which EgenSkriven skills are installed and where.

Checks both Claude Code and OpenCode locations (global and project).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			type LocationStatus struct {
				Path      string      `json:"path"`
				Exists    bool        `json:"exists"`
				Skills    []SkillInfo `json:"skills"`
				Available bool        `json:"available"`
			}

			checkLocation := func(basePath string) LocationStatus {
				status := LocationStatus{
					Path:   basePath,
					Skills: make([]SkillInfo, 0, len(skillNames)),
				}

				if _, err := os.Stat(basePath); err == nil {
					status.Exists = true
				}

				for _, skillName := range skillNames {
					skillFile := filepath.Join(basePath, skillName, "SKILL.md")
					info := SkillInfo{
						Name: skillName,
						Path: skillFile,
					}

					if _, err := os.Stat(skillFile); err == nil {
						info.Installed = true
						status.Available = true

						// Try to read description from frontmatter
						if content, err := os.ReadFile(skillFile); err == nil {
							info.Description = extractDescription(string(content))
						}
					}

					status.Skills = append(status.Skills, info)
				}

				return status
			}

			// Check all locations
			claudeGlobalPath, _ := getGlobalSkillPath()
			opencodeGlobalPath, _ := getGlobalOpenCodeSkillPath()

			claudeGlobalStatus := checkLocation(claudeGlobalPath)
			opencodeGlobalStatus := checkLocation(opencodeGlobalPath)
			claudeProjectStatus := checkLocation(getProjectSkillPath())
			opencodeProjectStatus := checkLocation(getProjectOpenCodeSkillPath())

			if out.JSON {
				return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
					"claude_global":    claudeGlobalStatus,
					"opencode_global":  opencodeGlobalStatus,
					"claude_project":   claudeProjectStatus,
					"opencode_project": opencodeProjectStatus,
				})
			}

			fmt.Println("EgenSkriven Skill Status")
			fmt.Println("========================")
			fmt.Println()

			printStatus := func(name string, status LocationStatus) {
				fmt.Printf("%s (%s)\n", name, status.Path)
				if !status.Exists {
					fmt.Println("  Directory does not exist")
					fmt.Println()
					return
				}

				for _, skill := range status.Skills {
					if skill.Installed {
						fmt.Printf("  [x] %s\n", skill.Name)
					} else {
						fmt.Printf("  [ ] %s\n", skill.Name)
					}
				}
				fmt.Println()
			}

			printStatus("Claude Code Global", claudeGlobalStatus)
			printStatus("OpenCode Global", opencodeGlobalStatus)
			printStatus("Claude Code Project", claudeProjectStatus)
			printStatus("OpenCode Project", opencodeProjectStatus)

			anyAvailable := claudeGlobalStatus.Available || opencodeGlobalStatus.Available ||
				claudeProjectStatus.Available || opencodeProjectStatus.Available
			if !anyAvailable {
				fmt.Println("No skills installed. Run: egenskriven skill install")
			}

			return nil
		},
	}
}

// extractDescription extracts the description from YAML frontmatter.
// Handles both single-line and multi-line YAML descriptions (using | or >).
func extractDescription(content string) string {
	lines := strings.Split(content, "\n")
	inFrontmatter := false
	inDescription := false
	var descLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for frontmatter boundaries
		if trimmed == "---" {
			if inFrontmatter {
				break
			}
			inFrontmatter = true
			continue
		}

		if !inFrontmatter {
			continue
		}

		// Check if we're starting the description field
		if strings.HasPrefix(line, "description:") {
			value := strings.TrimPrefix(line, "description:")
			value = strings.TrimSpace(value)

			// Check for multi-line indicators (| or >)
			if value == "|" || value == ">" || value == "|+" || value == ">+" ||
				value == "|-" || value == ">-" {
				inDescription = true
				continue
			}

			// Single-line description
			return value
		}

		// If we're in a multi-line description, collect indented lines
		if inDescription {
			// Multi-line content is indented; non-indented line ends it
			if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
				descLines = append(descLines, strings.TrimSpace(line))
			} else if trimmed != "" {
				// Hit a new field, stop collecting
				break
			}
		}
	}

	if len(descLines) > 0 {
		return strings.Join(descLines, " ")
	}

	return ""
}
