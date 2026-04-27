package dev

import (
	"maps"
	"path/filepath"
	"slices"
	"testing"

	"dappco.re/go/io"
)

func TestFindWorkflows_Good(t *testing.T) {
	// Create a temp directory with workflow files
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	mustNoError(t, io.Local.EnsureDir(workflowsDir))

	// Create some workflow files
	for _, name := range []string{"qa.yml", "tests.yml", "codeql.yaml"} {
		mustNoError(t, io.Local.Write(filepath.Join(workflowsDir, name), "name: Test"))
	}

	// Create a non-workflow file (should be ignored)
	mustNoError(t, io.Local.Write(filepath.Join(workflowsDir, "readme.md"), "# Workflows"))

	workflows := findWorkflows(tmpDir)
	mustLen(t, workflows, 3)

	// Check that all expected workflows are found
	found := make(map[string]bool)
	for _, wf := range workflows {
		found[wf] = true
	}

	for _, expected := range []string{"qa.yml", "tests.yml", "codeql.yaml"} {
		mustTrue(t, found[expected])
	}
}

func TestFindWorkflows_NoWorkflowsDir_Bad(t *testing.T) {
	tmpDir := t.TempDir()
	workflows := findWorkflows(tmpDir)

	mustLen(t, workflows, 0)
}

func TestFindTemplateWorkflow_Good(t *testing.T) {
	tmpDir := t.TempDir()
	templatesDir := filepath.Join(tmpDir, ".github", "workflow-templates")
	mustNoError(t, io.Local.EnsureDir(templatesDir))

	templateContent := "name: QA\non: [push]"
	mustNoError(t, io.Local.Write(filepath.Join(templatesDir, "qa.yml"), templateContent))

	// Test finding with .yml extension
	result := findTemplateWorkflow(tmpDir, "qa.yml")
	mustTrue(t, result != "")

	// Test finding without extension (should auto-add .yml)
	result = findTemplateWorkflow(tmpDir, "qa")
	mustTrue(t, result != "")
}

func TestFindTemplateWorkflow_FallbackToWorkflows_Good(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	mustNoError(t, io.Local.EnsureDir(workflowsDir))

	templateContent := "name: Tests\non: [push]"
	mustNoError(t, io.Local.Write(filepath.Join(workflowsDir, "tests.yml"), templateContent))

	result := findTemplateWorkflow(tmpDir, "tests.yml")
	mustTrue(t, result != "")
}

func TestFindTemplateWorkflow_NotFound_Bad(t *testing.T) {
	tmpDir := t.TempDir()

	result := findTemplateWorkflow(tmpDir, "nonexistent.yml")
	mustEqual(t, "", result)
}

func TestTemplateNames_Good(t *testing.T) {
	templateSet := map[string]bool{
		"z.yml": true,
		"a.yml": true,
		"m.yml": true,
	}

	names := slices.Sorted(maps.Keys(templateSet))

	mustLen(t, names, 3)
	mustEqual(t, "a.yml", names[0])
	mustEqual(t, "m.yml", names[1])
	mustEqual(t, "z.yml", names[2])
}
