package dev

import (
	"maps"
	"path/filepath"
	"slices"
	"testing"

	"dappco.re/go/core/io"
)

func TestFindWorkflows_Good(t *testing.T) {
	// Create a temp directory with workflow files
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := io.Local.EnsureDir(workflowsDir); err != nil {
		t.Fatalf("Failed to create workflows dir: %v", err)
	}

	// Create some workflow files
	for _, name := range []string{"qa.yml", "tests.yml", "codeql.yaml"} {
		if err := io.Local.Write(filepath.Join(workflowsDir, name), "name: Test"); err != nil {
			t.Fatalf("Failed to create workflow file: %v", err)
		}
	}

	// Create a non-workflow file (should be ignored)
	if err := io.Local.Write(filepath.Join(workflowsDir, "readme.md"), "# Workflows"); err != nil {
		t.Fatalf("Failed to create readme file: %v", err)
	}

	workflows := findWorkflows(tmpDir)

	if len(workflows) != 3 {
		t.Errorf("Expected 3 workflows, got %d", len(workflows))
	}

	// Check that all expected workflows are found
	found := make(map[string]bool)
	for _, wf := range workflows {
		found[wf] = true
	}

	for _, expected := range []string{"qa.yml", "tests.yml", "codeql.yaml"} {
		if !found[expected] {
			t.Errorf("Expected to find workflow %s", expected)
		}
	}
}

func TestFindWorkflows_NoWorkflowsDir(t *testing.T) {
	tmpDir := t.TempDir()
	workflows := findWorkflows(tmpDir)

	if len(workflows) != 0 {
		t.Errorf("Expected 0 workflows for non-existent dir, got %d", len(workflows))
	}
}

func TestFindTemplateWorkflow_Good(t *testing.T) {
	tmpDir := t.TempDir()
	templatesDir := filepath.Join(tmpDir, ".github", "workflow-templates")
	if err := io.Local.EnsureDir(templatesDir); err != nil {
		t.Fatalf("Failed to create templates dir: %v", err)
	}

	templateContent := "name: QA\non: [push]"
	if err := io.Local.Write(filepath.Join(templatesDir, "qa.yml"), templateContent); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	// Test finding with .yml extension
	result := findTemplateWorkflow(tmpDir, "qa.yml")
	if result == "" {
		t.Error("Expected to find qa.yml template")
	}

	// Test finding without extension (should auto-add .yml)
	result = findTemplateWorkflow(tmpDir, "qa")
	if result == "" {
		t.Error("Expected to find qa template without extension")
	}
}

func TestFindTemplateWorkflow_FallbackToWorkflows(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := io.Local.EnsureDir(workflowsDir); err != nil {
		t.Fatalf("Failed to create workflows dir: %v", err)
	}

	templateContent := "name: Tests\non: [push]"
	if err := io.Local.Write(filepath.Join(workflowsDir, "tests.yml"), templateContent); err != nil {
		t.Fatalf("Failed to create workflow file: %v", err)
	}

	result := findTemplateWorkflow(tmpDir, "tests.yml")
	if result == "" {
		t.Error("Expected to find tests.yml in workflows dir")
	}
}

func TestFindTemplateWorkflow_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	result := findTemplateWorkflow(tmpDir, "nonexistent.yml")
	if result != "" {
		t.Errorf("Expected empty string for non-existent template, got %s", result)
	}
}

func TestTemplateNames_Good(t *testing.T) {
	templateSet := map[string]bool{
		"z.yml": true,
		"a.yml": true,
		"m.yml": true,
	}

	names := slices.Sorted(maps.Keys(templateSet))

	if len(names) != 3 {
		t.Fatalf("Expected 3 template names, got %d", len(names))
	}

	if names[0] != "a.yml" || names[1] != "m.yml" || names[2] != "z.yml" {
		t.Fatalf("Expected sorted template names, got %v", names)
	}
}
