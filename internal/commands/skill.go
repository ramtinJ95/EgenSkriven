package commands

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/spf13/cobra"
)

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

// getGlobalSkillPath returns the global skills directory path
func getGlobalSkillPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(home, ".claude", "skills"), nil
}

// getProjectSkillPath returns the project skills directory path
func getProjectSkillPath() string {
	return filepath.Join(".claude", "skills")
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

Installation locations:
  Global:  ~/.claude/skills/
  Project: .claude/skills/

Both Claude Code and OpenCode read from these locations.`,
		Example: `  egenskriven skill install          # Interactive prompt
  egenskriven skill install --global  # Install globally
  egenskriven skill install --project # Install to current project
  egenskriven skill install --force   # Overwrite existing skills`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			// Determine installation path
			var installPath string
			var err error

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

			if global {
				installPath, err = getGlobalSkillPath()
				if err != nil {
					return err
				}
			} else {
				installPath = getProjectSkillPath()
			}

			if !out.JSON {
				fmt.Printf("\nInstalling skills to %s...\n\n", installPath)
			}

			// Install each skill
			installed := make([]string, 0, len(skillNames))
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

			if out.JSON {
				location := "global"
				if project {
					location = "project"
				}
				return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
					"installed": installed,
					"location":  location,
					"path":      installPath,
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

			if global {
				globalPath, err := getGlobalSkillPath()
				if err == nil {
					for _, skillName := range skillNames {
						skillDir := filepath.Join(globalPath, skillName)
						if _, err := os.Stat(skillDir); err == nil {
							if err := os.RemoveAll(skillDir); err == nil {
								removed = append(removed, skillDir)
								if !out.JSON {
									fmt.Printf("Removed: %s\n", skillDir)
								}
							}
						}
					}
				}
			}

			if project {
				projectPath := getProjectSkillPath()
				for _, skillName := range skillNames {
					skillDir := filepath.Join(projectPath, skillName)
					if _, err := os.Stat(skillDir); err == nil {
						if err := os.RemoveAll(skillDir); err == nil {
							removed = append(removed, skillDir)
							if !out.JSON {
								fmt.Printf("Removed: %s\n", skillDir)
							}
						}
					}
				}
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

Checks both global (~/.claude/skills/) and project (.claude/skills/) locations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			type LocationStatus struct {
				Path      string      `json:"path"`
				Exists    bool        `json:"exists"`
				Skills    []SkillInfo `json:"skills"`
				Available bool        `json:"available"`
			}

			globalPath, _ := getGlobalSkillPath()
			projectPath := getProjectSkillPath()

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

			globalStatus := checkLocation(globalPath)
			projectStatus := checkLocation(projectPath)

			if out.JSON {
				return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
					"global":  globalStatus,
					"project": projectStatus,
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

			printStatus("Global", globalStatus)
			printStatus("Project", projectStatus)

			if !globalStatus.Available && !projectStatus.Available {
				fmt.Println("No skills installed. Run: egenskriven skill install")
			}

			return nil
		},
	}
}

// extractDescription extracts the description from YAML frontmatter
func extractDescription(content string) string {
	lines := strings.Split(content, "\n")
	inFrontmatter := false

	for _, line := range lines {
		if strings.TrimSpace(line) == "---" {
			if inFrontmatter {
				break
			}
			inFrontmatter = true
			continue
		}

		if inFrontmatter && strings.HasPrefix(line, "description:") {
			desc := strings.TrimPrefix(line, "description:")
			return strings.TrimSpace(desc)
		}
	}

	return ""
}

// ListEmbeddedSkills returns the list of embedded skill names (for testing)
func ListEmbeddedSkills() ([]string, error) {
	var skills []string
	err := fs.WalkDir(embeddedSkills, "skills", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && path != "skills" {
			skills = append(skills, filepath.Base(path))
		}
		return nil
	})
	return skills, err
}
