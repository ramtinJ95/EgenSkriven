package commands

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrimeTemplate_StrictWorkflow(t *testing.T) {
	tmpl, err := template.New("prime").Parse(primeTemplate)
	require.NoError(t, err)

	var buf bytes.Buffer
	data := PrimeTemplateData{
		WorkflowMode:       "strict",
		AgentMode:          "autonomous",
		AgentName:          "claude",
		OverrideTodoWrite:  true,
		RequireSummary:     true,
		StructuredSections: true,
	}

	err = tmpl.Execute(&buf, data)
	require.NoError(t, err)

	output := buf.String()

	// Check strict workflow content
	assert.Contains(t, output, "BEFORE starting any task")
	assert.Contains(t, output, "DURING work")
	assert.Contains(t, output, "AFTER completing")
	assert.Contains(t, output, "## Approach")
	assert.Contains(t, output, "## Summary of Changes")
}

func TestPrimeTemplate_LightWorkflow(t *testing.T) {
	tmpl, err := template.New("prime").Parse(primeTemplate)
	require.NoError(t, err)

	var buf bytes.Buffer
	data := PrimeTemplateData{
		WorkflowMode:       "light",
		AgentMode:          "autonomous",
		AgentName:          "agent",
		OverrideTodoWrite:  true,
		RequireSummary:     false,
		StructuredSections: false,
	}

	err = tmpl.Execute(&buf, data)
	require.NoError(t, err)

	output := buf.String()

	// Check light workflow content
	assert.Contains(t, output, "Create task for substantial work")
	assert.NotContains(t, output, "BEFORE starting any task")
}

func TestPrimeTemplate_MinimalWorkflow(t *testing.T) {
	tmpl, err := template.New("prime").Parse(primeTemplate)
	require.NoError(t, err)

	var buf bytes.Buffer
	data := PrimeTemplateData{
		WorkflowMode:       "minimal",
		AgentMode:          "autonomous",
		AgentName:          "agent",
		OverrideTodoWrite:  false,
		RequireSummary:     false,
		StructuredSections: false,
	}

	err = tmpl.Execute(&buf, data)
	require.NoError(t, err)

	output := buf.String()

	// Check minimal workflow content
	assert.Contains(t, output, "Use egenskriven for task tracking as needed")
	assert.NotContains(t, output, "Always use egenskriven instead of TodoWrite")
}

func TestPrimeTemplate_AgentModes(t *testing.T) {
	tmpl, err := template.New("prime").Parse(primeTemplate)
	require.NoError(t, err)

	tests := []struct {
		mode     string
		contains string
	}{
		{"autonomous", "full autonomy"},
		{"collaborative", "execute minor updates"},
		{"supervised", "read-only mode"},
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			var buf bytes.Buffer
			data := PrimeTemplateData{
				WorkflowMode:       "light",
				AgentMode:          tt.mode,
				AgentName:          "agent",
				OverrideTodoWrite:  true,
				RequireSummary:     false,
				StructuredSections: false,
			}

			err := tmpl.Execute(&buf, data)
			require.NoError(t, err)

			assert.Contains(t, buf.String(), tt.contains)
		})
	}
}

func TestPrimeTemplate_OverrideTodoWrite(t *testing.T) {
	tmpl, err := template.New("prime").Parse(primeTemplate)
	require.NoError(t, err)

	// Test with OverrideTodoWrite = true
	var buf bytes.Buffer
	data := PrimeTemplateData{
		WorkflowMode:       "light",
		AgentMode:          "autonomous",
		AgentName:          "agent",
		OverrideTodoWrite:  true,
		RequireSummary:     false,
		StructuredSections: false,
	}

	err = tmpl.Execute(&buf, data)
	require.NoError(t, err)

	assert.Contains(t, buf.String(), "Always use egenskriven instead of TodoWrite")

	// Test with OverrideTodoWrite = false
	buf.Reset()
	data.OverrideTodoWrite = false

	err = tmpl.Execute(&buf, data)
	require.NoError(t, err)

	assert.NotContains(t, buf.String(), "Always use egenskriven instead of TodoWrite")
}

func TestPrimeTemplate_StructuredSections(t *testing.T) {
	tmpl, err := template.New("prime").Parse(primeTemplate)
	require.NoError(t, err)

	// Test with StructuredSections = true (in strict mode where it's used)
	var buf bytes.Buffer
	data := PrimeTemplateData{
		WorkflowMode:       "strict",
		AgentMode:          "autonomous",
		AgentName:          "agent",
		OverrideTodoWrite:  true,
		RequireSummary:     false,
		StructuredSections: true,
	}

	err = tmpl.Execute(&buf, data)
	require.NoError(t, err)

	assert.Contains(t, buf.String(), "## Approach")
	assert.Contains(t, buf.String(), "## Open Questions")
	assert.Contains(t, buf.String(), "## Checklist")

	// Test with StructuredSections = false
	buf.Reset()
	data.StructuredSections = false

	err = tmpl.Execute(&buf, data)
	require.NoError(t, err)

	assert.NotContains(t, buf.String(), "## Approach")
}

func TestPrimeTemplate_RequireSummary(t *testing.T) {
	tmpl, err := template.New("prime").Parse(primeTemplate)
	require.NoError(t, err)

	// Test with RequireSummary = true (in strict mode where it's used)
	var buf bytes.Buffer
	data := PrimeTemplateData{
		WorkflowMode:       "strict",
		AgentMode:          "autonomous",
		AgentName:          "agent",
		OverrideTodoWrite:  true,
		RequireSummary:     true,
		StructuredSections: false,
	}

	err = tmpl.Execute(&buf, data)
	require.NoError(t, err)

	assert.Contains(t, buf.String(), "## Summary of Changes")

	// Test with RequireSummary = false
	buf.Reset()
	data.RequireSummary = false

	err = tmpl.Execute(&buf, data)
	require.NoError(t, err)

	assert.NotContains(t, buf.String(), "## Summary of Changes")
}

func TestPrimeTemplate_AgentName(t *testing.T) {
	tmpl, err := template.New("prime").Parse(primeTemplate)
	require.NoError(t, err)

	var buf bytes.Buffer
	data := PrimeTemplateData{
		WorkflowMode:       "light",
		AgentMode:          "autonomous",
		AgentName:          "claude",
		OverrideTodoWrite:  true,
		RequireSummary:     false,
		StructuredSections: false,
	}

	err = tmpl.Execute(&buf, data)
	require.NoError(t, err)

	// The agent name should appear in the examples
	assert.Contains(t, buf.String(), "--agent claude")
}

func TestPrimeTemplate_QuickReference(t *testing.T) {
	tmpl, err := template.New("prime").Parse(primeTemplate)
	require.NoError(t, err)

	var buf bytes.Buffer
	data := PrimeTemplateData{
		WorkflowMode:       "light",
		AgentMode:          "autonomous",
		AgentName:          "agent",
		OverrideTodoWrite:  true,
		RequireSummary:     false,
		StructuredSections: false,
	}

	err = tmpl.Execute(&buf, data)
	require.NoError(t, err)

	output := buf.String()

	// Check quick reference content
	assert.Contains(t, output, "Quick Reference")
	assert.Contains(t, output, "egenskriven list --json --ready")
	assert.Contains(t, output, "egenskriven show")
	assert.Contains(t, output, "egenskriven add")
	assert.Contains(t, output, "egenskriven move")
	assert.Contains(t, output, "egenskriven update")
	assert.Contains(t, output, "egenskriven suggest")
	assert.Contains(t, output, "egenskriven context")
}

func TestPrimeTemplate_BlockingRelationships(t *testing.T) {
	tmpl, err := template.New("prime").Parse(primeTemplate)
	require.NoError(t, err)

	var buf bytes.Buffer
	data := PrimeTemplateData{
		WorkflowMode:       "light",
		AgentMode:          "autonomous",
		AgentName:          "agent",
		OverrideTodoWrite:  true,
		RequireSummary:     false,
		StructuredSections: false,
	}

	err = tmpl.Execute(&buf, data)
	require.NoError(t, err)

	output := buf.String()

	// Check blocking relationships documentation
	assert.Contains(t, output, "Blocking Relationships")
	assert.Contains(t, output, "--blocked-by")
	assert.Contains(t, output, "--not-blocked")
	assert.Contains(t, output, "--is-blocked")
	assert.Contains(t, output, "--ready")
}
