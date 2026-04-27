package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_RelativeDirFindsParentConfig_Good(t *testing.T) {
	root := t.TempDir()
	mustNoError(t, os.MkdirAll(filepath.Join(root, ".core"), 0o755))
	mustNoError(t, os.MkdirAll(filepath.Join(root, "packages", "app"), 0o755))
	mustNoError(t, os.WriteFile(filepath.Join(root, ".core", "workspace.yaml"), []byte(`version: 1
active: app
packages_dir: ./packages
`), 0o600))

	originalWD, err := os.Getwd()
	mustNoError(t, err)
	t.Cleanup(func() {
		mustNoError(t, os.Chdir(originalWD))
	})
	mustNoError(t, os.Chdir(filepath.Join(root, "packages", "app")))

	cfg, err := LoadConfig(".")
	mustNoError(t, err)
	mustNotNil(t, cfg)
	mustEqual(t, "app", cfg.Active)
	mustEqual(t, "./packages", cfg.PackagesDir)
}

func mustNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func mustEqual[T comparable](t *testing.T, want, got T) {
	t.Helper()
	if want != got {
		t.Fatalf("want %v, got %v", want, got)
	}
}

func mustNotNil(t *testing.T, got any) {
	t.Helper()
	if got == nil {
		t.Fatal("expected non-nil")
	}
}
