package docs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCopyZensicalReadme_Good(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	src := filepath.Join(srcDir, "README.md")
	if err := os.WriteFile(src, []byte("# Hello\n\nBody text.\n"), 0o644); err != nil {
		t.Fatalf("write source README: %v", err)
	}

	if err := copyZensicalReadme(src, destDir); err != nil {
		t.Fatalf("copy README: %v", err)
	}

	output := filepath.Join(destDir, "index.md")
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read output index.md: %v", err)
	}

	content := string(data)
	if !strings.HasPrefix(content, "---\n") {
		t.Fatalf("expected Hugo front matter at start, got: %q", content)
	}
	if !strings.Contains(content, "title: \"README\"") {
		t.Fatalf("expected README title in front matter, got: %q", content)
	}
	if !strings.Contains(content, "Body text.") {
		t.Fatalf("expected README body to be preserved, got: %q", content)
	}
}

func TestResetOutputDir_ClearsExistingFiles(t *testing.T) {
	dir := t.TempDir()

	stale := filepath.Join(dir, "stale.md")
	if err := os.WriteFile(stale, []byte("old content"), 0o644); err != nil {
		t.Fatalf("write stale file: %v", err)
	}

	if err := resetOutputDir(dir); err != nil {
		t.Fatalf("reset output dir: %v", err)
	}

	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatalf("expected stale file to be removed, got err=%v", err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat output dir: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected output dir to exist as a directory")
	}
}
