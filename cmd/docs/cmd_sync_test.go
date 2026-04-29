package docs

import (
	. "dappco.re/go"
	"testing"
)

func TestCopyZensicalReadme_Good(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	src := PathJoin(srcDir, "README.md")
	if r := WriteFile(src, []byte("# Hello\n\nBody text.\n"), 0o644); !r.OK {
		t.Fatalf("write source README: %v", r.Error())
	}

	if err := copyZensicalReadme(src, destDir); err != nil {
		t.Fatalf("copy README: %v", err)
	}

	output := PathJoin(destDir, "index.md")
	read := ReadFile(output)
	if !read.OK {
		t.Fatalf("read output index.md: %v", read.Error())
	}

	content := string(read.Value.([]byte))
	if !HasPrefix(content, "---\n") {
		t.Fatalf("expected Hugo front matter at start, got: %q", content)
	}
	if !Contains(content, "title: \"README\"") {
		t.Fatalf("expected README title in front matter, got: %q", content)
	}
	if !Contains(content, "Body text.") {
		t.Fatalf("expected README body to be preserved, got: %q", content)
	}
}

func TestResetOutputDirClearsExistingFiles(t *testing.T) {
	dir := t.TempDir()

	stale := PathJoin(dir, "stale.md")
	if r := WriteFile(stale, []byte("old content"), 0o644); !r.OK {
		t.Fatalf("write stale file: %v", r.Error())
	}

	if err := resetOutputDir(dir); err != nil {
		t.Fatalf("reset output dir: %v", err)
	}

	if r := Stat(stale); r.OK || !IsNotExist(r.Value.(error)) {
		t.Fatalf("expected stale file to be removed, got result=%v", r)
	}

	stat := Stat(dir)
	if !stat.OK {
		t.Fatalf("stat output dir: %v", stat.Error())
	}
	info := stat.Value.(interface{ IsDir() bool })
	if !info.IsDir() {
		t.Fatalf("expected output dir to exist as a directory")
	}
}

func TestGoHelpOutputName_Good(t *testing.T) {
	cases := map[string]string{
		"core":        "go",
		"core-admin":  "admin",
		"core-api":    "api",
		"go-example":  "go-example",
		"custom-repo": "custom-repo",
	}

	for input, want := range cases {
		if got := goHelpOutputName(input); got != want {
			t.Fatalf("goHelpOutputName(%q) = %q, want %q", input, got, want)
		}
	}
}
