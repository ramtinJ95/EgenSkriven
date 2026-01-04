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
