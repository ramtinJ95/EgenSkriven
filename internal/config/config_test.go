package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadProjectConfig_Defaults(t *testing.T) {
	// Create a temp directory without config file
	tmpDir, err := os.MkdirTemp("", "egenskriven-config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	cfg, err := LoadProjectConfigFrom(tmpDir)

	require.NoError(t, err)
	assert.Equal(t, "light", cfg.Agent.Workflow)
	assert.Equal(t, "autonomous", cfg.Agent.Mode)
	assert.True(t, cfg.Agent.OverrideTodoWrite)
	assert.False(t, cfg.Agent.RequireSummary)
	assert.False(t, cfg.Agent.StructuredSections)
}

func TestLoadProjectConfig_FromFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "egenskriven-config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create config file
	configDir := filepath.Join(tmpDir, ".egenskriven")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	configContent := `{
		"agent": {
			"workflow": "strict",
			"mode": "collaborative",
			"overrideTodoWrite": false,
			"requireSummary": true,
			"structuredSections": true
		}
	}`

	require.NoError(t, os.WriteFile(
		filepath.Join(configDir, "config.json"),
		[]byte(configContent),
		0644,
	))

	cfg, err := LoadProjectConfigFrom(tmpDir)

	require.NoError(t, err)
	assert.Equal(t, "strict", cfg.Agent.Workflow)
	assert.Equal(t, "collaborative", cfg.Agent.Mode)
	assert.False(t, cfg.Agent.OverrideTodoWrite)
	assert.True(t, cfg.Agent.RequireSummary)
	assert.True(t, cfg.Agent.StructuredSections)
}

func TestLoadProjectConfig_InvalidJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "egenskriven-config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create invalid config file
	configDir := filepath.Join(tmpDir, ".egenskriven")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(configDir, "config.json"),
		[]byte("invalid json"),
		0644,
	))

	_, err = LoadProjectConfigFrom(tmpDir)

	assert.Error(t, err)
}

func TestLoadProjectConfig_InvalidWorkflow(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "egenskriven-config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create config with invalid workflow
	configDir := filepath.Join(tmpDir, ".egenskriven")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	configContent := `{"agent": {"workflow": "invalid"}}`
	require.NoError(t, os.WriteFile(
		filepath.Join(configDir, "config.json"),
		[]byte(configContent),
		0644,
	))

	cfg, err := LoadProjectConfigFrom(tmpDir)

	require.NoError(t, err)
	assert.Equal(t, "light", cfg.Agent.Workflow) // Should default to light
}

func TestLoadProjectConfig_InvalidMode(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "egenskriven-config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create config with invalid mode
	configDir := filepath.Join(tmpDir, ".egenskriven")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	configContent := `{"agent": {"mode": "invalid"}}`
	require.NoError(t, os.WriteFile(
		filepath.Join(configDir, "config.json"),
		[]byte(configContent),
		0644,
	))

	cfg, err := LoadProjectConfigFrom(tmpDir)

	require.NoError(t, err)
	assert.Equal(t, "autonomous", cfg.Agent.Mode) // Should default to autonomous
}

func TestSaveConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "egenskriven-config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	cfg := &Config{
		Agent: AgentConfig{
			Workflow:           "strict",
			Mode:               "supervised",
			OverrideTodoWrite:  true,
			RequireSummary:     true,
			StructuredSections: true,
		},
	}

	err = SaveConfig(tmpDir, cfg)
	require.NoError(t, err)

	// Verify file was created
	configPath := filepath.Join(tmpDir, ".egenskriven", "config.json")
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	// Load and verify contents
	loaded, err := LoadProjectConfigFrom(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, cfg.Agent.Workflow, loaded.Agent.Workflow)
	assert.Equal(t, cfg.Agent.Mode, loaded.Agent.Mode)
	assert.Equal(t, cfg.Agent.OverrideTodoWrite, loaded.Agent.OverrideTodoWrite)
	assert.Equal(t, cfg.Agent.RequireSummary, loaded.Agent.RequireSummary)
	assert.Equal(t, cfg.Agent.StructuredSections, loaded.Agent.StructuredSections)
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, "light", cfg.Agent.Workflow)
	assert.Equal(t, "autonomous", cfg.Agent.Mode)
	assert.True(t, cfg.Agent.OverrideTodoWrite)
	assert.False(t, cfg.Agent.RequireSummary)
	assert.False(t, cfg.Agent.StructuredSections)
}

func TestLoadProjectConfig_PartialConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "egenskriven-config-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create config with only some fields
	configDir := filepath.Join(tmpDir, ".egenskriven")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	configContent := `{"agent": {"workflow": "strict"}}`
	require.NoError(t, os.WriteFile(
		filepath.Join(configDir, "config.json"),
		[]byte(configContent),
		0644,
	))

	cfg, err := LoadProjectConfigFrom(tmpDir)

	require.NoError(t, err)
	// Specified field should be overridden
	assert.Equal(t, "strict", cfg.Agent.Workflow)
	// Other fields should keep defaults
	assert.Equal(t, "autonomous", cfg.Agent.Mode)
	assert.True(t, cfg.Agent.OverrideTodoWrite)
}

// Tests for GlobalConfig and merged config

func TestDefaultGlobalConfig(t *testing.T) {
	cfg := DefaultGlobalConfig()

	assert.Empty(t, cfg.DataDir)
	assert.Empty(t, cfg.Defaults.Author)
	assert.Equal(t, "agent", cfg.Defaults.Agent)
	assert.Equal(t, "light", cfg.Agent.Workflow)
	assert.Equal(t, "autonomous", cfg.Agent.Mode)
	assert.Equal(t, "command", cfg.Agent.ResumeMode)
	assert.Equal(t, "http://localhost:8090", cfg.Server.URL)
}

func TestValidateResumeMode(t *testing.T) {
	tests := []struct {
		mode    string
		wantErr bool
	}{
		{"manual", false},
		{"command", false},
		{"auto", false},
		{"invalid", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			err := ValidateResumeMode(tt.mode)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMerge_GlobalOnly(t *testing.T) {
	global := &GlobalConfig{
		DataDir: "/data",
		Defaults: DefaultsConfig{
			Author: "testuser",
			Agent:  "claude",
		},
		Agent: AgentConfig{
			Workflow:   "strict",
			Mode:       "collaborative",
			ResumeMode: "auto",
		},
		Server: ServerConfig{
			URL: "http://localhost:9090",
		},
	}

	// Empty project config (no overrides)
	project := &Config{}

	merged := merge(global, project)

	// Global-only values
	assert.Equal(t, "/data", merged.DataDir)
	assert.Equal(t, "testuser", merged.Defaults.Author)
	assert.Equal(t, "claude", merged.Defaults.Agent)

	// With empty project config, global agent settings should be used
	assert.Equal(t, "strict", merged.Agent.Workflow)
	assert.Equal(t, "collaborative", merged.Agent.Mode)
	assert.Equal(t, "auto", merged.Agent.ResumeMode)
	assert.Equal(t, "http://localhost:9090", merged.Server.URL)
}

func TestMerge_ProjectOverrides(t *testing.T) {
	global := &GlobalConfig{
		DataDir: "/data",
		Defaults: DefaultsConfig{
			Author: "testuser",
			Agent:  "claude",
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

	project := &Config{
		Agent: AgentConfig{
			Workflow:          "strict",
			Mode:              "supervised",
			ResumeMode:        "auto",
			OverrideTodoWrite: true,
		},
		Server: ServerConfig{
			URL: "http://localhost:9999",
		},
		DefaultBoard: "WRK",
	}

	merged := merge(global, project)

	// Global-only values should remain unchanged
	assert.Equal(t, "/data", merged.DataDir)
	assert.Equal(t, "testuser", merged.Defaults.Author)
	assert.Equal(t, "claude", merged.Defaults.Agent)

	// Project overrides should take precedence
	assert.Equal(t, "strict", merged.Agent.Workflow)
	assert.Equal(t, "supervised", merged.Agent.Mode)
	assert.Equal(t, "auto", merged.Agent.ResumeMode)
	assert.True(t, merged.Agent.OverrideTodoWrite)
	assert.Equal(t, "http://localhost:9999", merged.Server.URL)
	assert.Equal(t, "WRK", merged.DefaultBoard)
}

func TestMerge_PartialProjectOverride(t *testing.T) {
	global := &GlobalConfig{
		DataDir: "/data",
		Agent: AgentConfig{
			Workflow:   "light",
			Mode:       "autonomous",
			ResumeMode: "command",
		},
		Server: ServerConfig{
			URL: "http://localhost:8090",
		},
	}

	// Project only overrides workflow
	project := &Config{
		Agent: AgentConfig{
			Workflow: "strict",
		},
	}

	merged := merge(global, project)

	// Workflow should be overridden
	assert.Equal(t, "strict", merged.Agent.Workflow)
	// Other values should come from global
	assert.Equal(t, "autonomous", merged.Agent.Mode)
	assert.Equal(t, "command", merged.Agent.ResumeMode)
	assert.Equal(t, "http://localhost:8090", merged.Server.URL)
}
