package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ValidWorkflows is the list of valid workflow values.
var ValidWorkflows = []string{"strict", "light", "minimal"}

// ValidModes is the list of valid agent mode values.
var ValidModes = []string{"autonomous", "collaborative", "supervised"}

// AgentConfig defines agent-specific behavior settings.
type AgentConfig struct {
	// Workflow mode: "strict", "light", "minimal"
	// - strict: Full enforcement (create before, update during, summary after)
	// - light: Basic tracking (create/complete, no structured sections)
	// - minimal: No enforcement (agent decides when to use)
	Workflow string `json:"workflow"`

	// Mode defines agent autonomy: "autonomous", "collaborative", "supervised"
	// - autonomous: Agent executes actions directly
	// - collaborative: Agent proposes major changes, executes minor ones
	// - supervised: Agent is read-only, outputs commands for human
	Mode string `json:"mode"`

	// OverrideTodoWrite tells agents to ignore built-in todo tools
	OverrideTodoWrite bool `json:"overrideTodoWrite"`

	// RequireSummary requires agents to add summary section on completion
	RequireSummary bool `json:"requireSummary"`

	// StructuredSections encourages structured markdown sections in descriptions
	StructuredSections bool `json:"structuredSections"`
}

// Config represents the project configuration.
type Config struct {
	Agent AgentConfig `json:"agent"`
}

// DefaultConfig returns configuration with default values.
func DefaultConfig() *Config {
	return &Config{
		Agent: AgentConfig{
			Workflow:           "light",
			Mode:               "autonomous",
			OverrideTodoWrite:  true,
			RequireSummary:     false,
			StructuredSections: false,
		},
	}
}

// LoadProjectConfig loads configuration from .egenskriven/config.json.
// Returns default config if file doesn't exist.
func LoadProjectConfig() (*Config, error) {
	return LoadProjectConfigFrom(".")
}

// LoadProjectConfigFrom loads configuration from a specific directory.
func LoadProjectConfigFrom(dir string) (*Config, error) {
	configPath := filepath.Join(dir, ".egenskriven", "config.json")

	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		// Return defaults if no config file
		return DefaultConfig(), nil
	}
	if err != nil {
		return nil, err
	}

	// Start with defaults, then overlay file values
	cfg := DefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Validate values
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// ValidateWorkflow checks if a workflow value is valid.
// Returns an error if invalid.
func ValidateWorkflow(workflow string) error {
	for _, valid := range ValidWorkflows {
		if workflow == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid workflow '%s', must be one of: %v", workflow, ValidWorkflows)
}

// ValidatMode checks if an agent mode value is valid.
// Returns an error if invalid.
func ValidateMode(mode string) error {
	for _, valid := range ValidModes {
		if mode == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid mode '%s', must be one of: %v", mode, ValidModes)
}

// validateConfig ensures config values are valid, normalizing invalid values to defaults.
func validateConfig(cfg *Config) error {
	if err := ValidateWorkflow(cfg.Agent.Workflow); err != nil {
		cfg.Agent.Workflow = "light" // Default to light if invalid
	}

	if err := ValidateMode(cfg.Agent.Mode); err != nil {
		cfg.Agent.Mode = "autonomous" // Default to autonomous if invalid
	}

	return nil
}

// SaveConfig saves configuration to .egenskriven/config.json.
func SaveConfig(dir string, cfg *Config) error {
	configDir := filepath.Join(dir, ".egenskriven")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "config.json")

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
