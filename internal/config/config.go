package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Global config cache - loaded once per process
var (
	globalConfigOnce  sync.Once
	globalConfigCache *GlobalConfig
	globalConfigError error
)

// ValidWorkflows is the list of valid workflow values.
var ValidWorkflows = []string{"strict", "light", "minimal"}

// ValidModes is the list of valid agent mode values.
var ValidModes = []string{"autonomous", "collaborative", "supervised"}

// ValidResumeModes is the list of valid resume mode values.
var ValidResumeModes = []string{"manual", "command", "auto"}

// ServerConfig defines server connection settings for CLI hybrid mode.
type ServerConfig struct {
	// URL is the PocketBase server URL (default: http://localhost:8090)
	URL string `json:"url,omitempty"`
}

// AgentConfig defines agent-specific behavior settings.
type AgentConfig struct {
	// Workflow mode: "strict", "light", "minimal"
	// - strict: Full enforcement (create before, update during, summary after)
	// - light: Basic tracking (create/complete, no structured sections)
	// - minimal: No enforcement (agent decides when to use)
	Workflow string `json:"workflow,omitempty"`

	// Mode defines agent autonomy: "autonomous", "collaborative", "supervised"
	// - autonomous: Agent executes actions directly
	// - collaborative: Agent proposes major changes, executes minor ones
	// - supervised: Agent is read-only, outputs commands for human
	Mode string `json:"mode,omitempty"`

	// ResumeMode defines default resume behavior for new boards: "manual", "command", "auto"
	// - manual: Never auto-resume, user must explicitly resume
	// - command: Resume when /resume command is used (default)
	// - auto: Auto-resume when new comments are added
	ResumeMode string `json:"resume_mode,omitempty"`

	// OverrideTodoWrite tells agents to ignore built-in todo tools
	OverrideTodoWrite bool `json:"overrideTodoWrite,omitempty"`

	// RequireSummary requires agents to add summary section on completion
	RequireSummary bool `json:"requireSummary,omitempty"`

	// StructuredSections encourages structured markdown sections in descriptions
	StructuredSections bool `json:"structuredSections,omitempty"`
}

// Config represents the project configuration.
// Location: .egenskriven/config.json
type Config struct {
	Agent        AgentConfig  `json:"agent"`
	Server       ServerConfig `json:"server,omitempty"`
	DefaultBoard string       `json:"default_board,omitempty"` // Default board prefix for CLI commands
}

// DefaultsConfig contains default values for commands.
type DefaultsConfig struct {
	// Author is the default author name for comments
	Author string `json:"author,omitempty"`
	// Agent is the default agent name for block command
	Agent string `json:"agent,omitempty"`
}

// GlobalConfig represents user-wide configuration.
// Location: ~/.config/egenskriven/config.json
type GlobalConfig struct {
	// DataDir is the path to the data directory (default: use PocketBase default)
	// Supports ~ expansion for home directory
	DataDir string `json:"data_dir,omitempty"`
	// Defaults contains default values for commands
	Defaults DefaultsConfig `json:"defaults,omitempty"`
	// Agent contains agent behavior settings
	Agent AgentConfig `json:"agent,omitempty"`
	// Server contains server connection settings
	Server ServerConfig `json:"server,omitempty"`
}

// MergedConfig represents the effective configuration after merging
// global and project configs. Project values override global values.
type MergedConfig struct {
	// From global only
	DataDir  string
	Defaults DefaultsConfig

	// Merged (project overrides global)
	Agent  AgentConfig
	Server ServerConfig

	// From project only
	DefaultBoard string
}

// DefaultConfig returns project configuration with default values.
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

// DefaultGlobalConfig returns global configuration with default values.
func DefaultGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		DataDir: "", // Empty means use PocketBase default (pb_data in cwd)
		Defaults: DefaultsConfig{
			Author: "",
			Agent:  "agent",
		},
		Agent: AgentConfig{
			Workflow:   "light",
			Mode:       "autonomous",
			ResumeMode: "command",
		},
		Server: ServerConfig{
			URL: "http://localhost:8090",
		},
	}
}

// GlobalConfigPath returns the path to the global config file.
// Returns ~/.config/egenskriven/config.json
func GlobalConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "egenskriven", "config.json"), nil
}

// LoadGlobalConfig loads configuration from ~/.config/egenskriven/config.json.
// Returns default config if file doesn't exist.
// The config is cached after first load - subsequent calls return the cached value.
func LoadGlobalConfig() (*GlobalConfig, error) {
	globalConfigOnce.Do(func() {
		globalConfigCache, globalConfigError = loadGlobalConfigFromDisk()
	})
	return globalConfigCache, globalConfigError
}

// loadGlobalConfigFromDisk reads the global config from disk.
// This is the internal implementation; use LoadGlobalConfig for cached access.
func loadGlobalConfigFromDisk() (*GlobalConfig, error) {
	configPath, err := GlobalConfigPath()
	if err != nil {
		return DefaultGlobalConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		return DefaultGlobalConfig(), nil
	}
	if err != nil {
		return nil, err
	}

	cfg := DefaultGlobalConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Expand ~ in data_dir
	if strings.HasPrefix(cfg.DataDir, "~/") {
		home, _ := os.UserHomeDir()
		cfg.DataDir = filepath.Join(home, cfg.DataDir[2:])
	}

	// Validate values
	if cfg.Agent.Workflow != "" {
		if err := ValidateWorkflow(cfg.Agent.Workflow); err != nil {
			cfg.Agent.Workflow = "light"
		}
	}
	if cfg.Agent.Mode != "" {
		if err := ValidateMode(cfg.Agent.Mode); err != nil {
			cfg.Agent.Mode = "autonomous"
		}
	}
	if cfg.Agent.ResumeMode != "" {
		if err := ValidateResumeMode(cfg.Agent.ResumeMode); err != nil {
			cfg.Agent.ResumeMode = "command"
		}
	}

	return cfg, nil
}

// ResetGlobalConfigCache clears the cached global config.
// This is primarily useful for testing.
func ResetGlobalConfigCache() {
	globalConfigOnce = sync.Once{}
	globalConfigCache = nil
	globalConfigError = nil
}

// SaveGlobalConfig saves configuration to ~/.config/egenskriven/config.json.
func SaveGlobalConfig(cfg *GlobalConfig) error {
	configPath, err := GlobalConfigPath()
	if err != nil {
		return err
	}

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
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

// ValidateResumeMode checks if a resume mode value is valid.
// Returns an error if invalid.
func ValidateResumeMode(mode string) error {
	for _, valid := range ValidResumeModes {
		if mode == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid resume_mode '%s', must be one of: %v", mode, ValidResumeModes)
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

// Load loads and merges global and project configuration.
// Project config values override global config where both are set.
func Load() (*MergedConfig, error) {
	global, err := LoadGlobalConfig()
	if err != nil {
		return nil, fmt.Errorf("loading global config: %w", err)
	}

	project, err := LoadProjectConfig()
	if err != nil {
		return nil, fmt.Errorf("loading project config: %w", err)
	}

	return merge(global, project), nil
}

// merge combines global and project configs with project taking precedence.
func merge(global *GlobalConfig, project *Config) *MergedConfig {
	merged := &MergedConfig{
		// Global-only values
		DataDir:  global.DataDir,
		Defaults: global.Defaults,

		// Start with global agent config
		Agent: global.Agent,

		// Start with global server config
		Server: global.Server,

		// Project-only values
		DefaultBoard: project.DefaultBoard,
	}

	// Override agent settings from project if set
	if project.Agent.Workflow != "" {
		merged.Agent.Workflow = project.Agent.Workflow
	}
	if project.Agent.Mode != "" {
		merged.Agent.Mode = project.Agent.Mode
	}
	if project.Agent.ResumeMode != "" {
		merged.Agent.ResumeMode = project.Agent.ResumeMode
	}
	// Boolean fields from project always override (no way to distinguish unset from false)
	merged.Agent.OverrideTodoWrite = project.Agent.OverrideTodoWrite
	merged.Agent.RequireSummary = project.Agent.RequireSummary
	merged.Agent.StructuredSections = project.Agent.StructuredSections

	// Override server URL from project if set
	if project.Server.URL != "" {
		merged.Server.URL = project.Server.URL
	}

	return merged
}
