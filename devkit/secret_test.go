package devkit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanDir_Good(t *testing.T) {
	root := t.TempDir()

	if err := os.WriteFile(filepath.Join(root, "config.yml"), []byte(`
api_key: "ghp_abcdefghijklmnopqrstuvwxyz1234"
`), 0o600); err != nil {
		t.Fatalf("write config.yml: %v", err)
	}

	if err := os.Mkdir(filepath.Join(root, "nested"), 0o755); err != nil {
		t.Fatalf("create nested dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "nested", "creds.txt"), []byte("access_key = AKIA1234567890ABCDEF\n"), 0o600); err != nil {
		t.Fatalf("write creds.txt: %v", err)
	}

	findings, err := ScanDir(root)
	if err != nil {
		t.Fatalf("scan dir: %v", err)
	}
	if len(findings) != 2 {
		t.Fatalf("findings length = %d, want 2", len(findings))
	}

	if findings[0].Rule != "github-token" {
		t.Fatalf("findings[0].Rule = %q, want %q", findings[0].Rule, "github-token")
	}
	if findings[0].Line != 2 {
		t.Fatalf("findings[0].Line = %d, want 2", findings[0].Line)
	}
	if got := filepath.Base(findings[0].Path); got != "config.yml" {
		t.Fatalf("findings[0] path base = %q, want %q", got, "config.yml")
	}

	if findings[1].Rule != "aws-access-key-id" {
		t.Fatalf("findings[1].Rule = %q, want %q", findings[1].Rule, "aws-access-key-id")
	}
	if findings[1].Line != 1 {
		t.Fatalf("findings[1].Line = %d, want 1", findings[1].Line)
	}
	if got := filepath.Base(findings[1].Path); got != "creds.txt" {
		t.Fatalf("findings[1] path base = %q, want %q", got, "creds.txt")
	}
}

func TestScanDir_SkipsBinaryAndIgnoredDirs_Good(t *testing.T) {
	root := t.TempDir()

	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatalf("create .git dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".git", "config"), []byte("token=ghp_abcdefghijklmnopqrstuvwxyz1234"), 0o600); err != nil {
		t.Fatalf("write .git config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "blob.bin"), []byte{0, 1, 2, 3, 4}, 0o600); err != nil {
		t.Fatalf("write blob.bin: %v", err)
	}

	findings, err := ScanDir(root)
	if err != nil {
		t.Fatalf("scan dir: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("findings length = %d, want 0", len(findings))
	}
}

func TestScanDir_ReportsGenericAssignments_Bad(t *testing.T) {
	root := t.TempDir()

	if err := os.WriteFile(filepath.Join(root, "secrets.env"), []byte("client_secret: abcdefghijklmnop\n"), 0o600); err != nil {
		t.Fatalf("write secrets.env: %v", err)
	}

	findings, err := ScanDir(root)
	if err != nil {
		t.Fatalf("scan dir: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("findings length = %d, want 1", len(findings))
	}
	if findings[0].Rule != "generic-secret-assignment" {
		t.Fatalf("findings[0].Rule = %q, want %q", findings[0].Rule, "generic-secret-assignment")
	}
	if findings[0].Line != 1 {
		t.Fatalf("findings[0].Line = %d, want 1", findings[0].Line)
	}
	if findings[0].Column != 1 {
		t.Fatalf("findings[0].Column = %d, want 1", findings[0].Column)
	}
}
