package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_RelativeDirFindsParentConfig_Good(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".core"), 0o755); err != nil {
		t.Fatalf("create .core dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "packages", "app"), 0o755); err != nil {
		t.Fatalf("create app dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".core", "workspace.yaml"), []byte(`version: 1
active: app
packages_dir: ./packages
`), 0o600); err != nil {
		t.Fatalf("write workspace config: %v", err)
	}

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalWD); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})
	if err := os.Chdir(filepath.Join(root, "packages", "app")); err != nil {
		t.Fatalf("change working directory: %v", err)
	}

	cfg, err := LoadConfig(".")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected config")
	}
	if cfg.Active != "app" {
		t.Fatalf("active = %q, want app", cfg.Active)
	}
	if cfg.PackagesDir != "./packages" {
		t.Fatalf("packages dir = %q, want ./packages", cfg.PackagesDir)
	}
}
