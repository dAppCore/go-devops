package dev

import (
	core "dappco.re/go"
	"maps"
	"slices"
	"testing"

	"dappco.re/go/io"
)

func TestFindWorkflows_Good(t *testing.T) {
	// Create a temp directory with workflow files
	tmpDir := t.TempDir()
	workflowsDir := core.PathJoin(tmpDir, ".github", "workflows")
	if err := io.Local.EnsureDir(workflowsDir); err != nil {
		t.Fatalf("create workflows dir: %v", err)
	}

	// Create some workflow files
	for _, name := range []string{"qa.yml", "tests.yml", "codeql.yaml"} {
		if err := io.Local.Write(core.PathJoin(workflowsDir, name), "name: Test"); err != nil {
			t.Fatalf("write workflow %s: %v", name, err)
		}
	}

	// Create a non-workflow file (should be ignored)
	if err := io.Local.Write(core.PathJoin(workflowsDir, "readme.md"), "# Workflows"); err != nil {
		t.Fatalf("write readme: %v", err)
	}

	workflows := findWorkflows(tmpDir)
	if len(workflows) != 3 {
		t.Fatalf("workflows length = %d, want 3", len(workflows))
	}

	// Check that all expected workflows are found
	found := make(map[string]bool)
	for _, wf := range workflows {
		found[wf] = true
	}

	for _, expected := range []string{"qa.yml", "tests.yml", "codeql.yaml"} {
		if !found[expected] {
			t.Fatalf("expected workflow %s in %v", expected, workflows)
		}
	}
}

func TestFindWorkflowsNoWorkflowsDir(t *testing.T) {
	tmpDir := t.TempDir()
	workflows := findWorkflows(tmpDir)

	if len(workflows) != 0 {
		t.Fatalf("workflows length = %d, want 0", len(workflows))
	}
}

func TestFindTemplateWorkflow_Good(t *testing.T) {
	tmpDir := t.TempDir()
	templatesDir := core.PathJoin(tmpDir, ".github", "workflow-templates")
	if err := io.Local.EnsureDir(templatesDir); err != nil {
		t.Fatalf("create templates dir: %v", err)
	}

	templateContent := "name: QA\non: [push]"
	if err := io.Local.Write(core.PathJoin(templatesDir, "qa.yml"), templateContent); err != nil {
		t.Fatalf("write template workflow: %v", err)
	}

	// Test finding with .yml extension
	result := findTemplateWorkflow(tmpDir, "qa.yml")
	if result == "" {
		t.Fatal("expected template workflow for qa.yml")
	}

	// Test finding without extension (should auto-add .yml)
	result = findTemplateWorkflow(tmpDir, "qa")
	if result == "" {
		t.Fatal("expected template workflow for qa")
	}
}

func TestFindTemplateWorkflowFallbackToWorkflows(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := core.PathJoin(tmpDir, ".github", "workflows")
	if err := io.Local.EnsureDir(workflowsDir); err != nil {
		t.Fatalf("create workflows dir: %v", err)
	}

	templateContent := "name: Tests\non: [push]"
	if err := io.Local.Write(core.PathJoin(workflowsDir, "tests.yml"), templateContent); err != nil {
		t.Fatalf("write workflow: %v", err)
	}

	result := findTemplateWorkflow(tmpDir, "tests.yml")
	if result == "" {
		t.Fatal("expected fallback workflow")
	}
}

func TestFindTemplateWorkflowNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	result := findTemplateWorkflow(tmpDir, "nonexistent.yml")
	if result != "" {
		t.Fatalf("result = %q, want empty", result)
	}
}

func TestTemplateNames_Good(t *testing.T) {
	templateSet := map[string]bool{
		"z.yml": true,
		"a.yml": true,
		"m.yml": true,
	}

	names := slices.Sorted(maps.Keys(templateSet))

	want := []string{"a.yml", "m.yml", "z.yml"}
	if !slices.Equal(names, want) {
		t.Fatalf("names = %v, want %v", names, want)
	}
}
