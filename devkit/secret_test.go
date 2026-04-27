package devkit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanDir_Good(t *testing.T) {
	root := t.TempDir()

	mustNoError(t, os.WriteFile(filepath.Join(root, "config.yml"), []byte(`
api_key: "ghp_abcdefghijklmnopqrstuvwxyz1234"
`), 0o600))

	mustNoError(t, os.Mkdir(filepath.Join(root, "nested"), 0o755))
	mustNoError(t, os.WriteFile(filepath.Join(root, "nested", "creds.txt"), []byte("access_key = AKIA1234567890ABCDEF\n"), 0o600))

	findings, err := ScanDir(root)
	mustNoError(t, err)
	mustLen(t, findings, 2)

	mustEqual(t, "github-token", findings[0].Rule)
	mustEqual(t, 2, findings[0].Line)
	mustEqual(t, "config.yml", filepath.Base(findings[0].Path))

	mustEqual(t, "aws-access-key-id", findings[1].Rule)
	mustEqual(t, 1, findings[1].Line)
	mustEqual(t, "creds.txt", filepath.Base(findings[1].Path))
}

func TestScanDir_SkipsBinaryAndIgnoredDirs_Good(t *testing.T) {
	root := t.TempDir()

	mustNoError(t, os.Mkdir(filepath.Join(root, ".git"), 0o755))
	mustNoError(t, os.WriteFile(filepath.Join(root, ".git", "config"), []byte("token=ghp_abcdefghijklmnopqrstuvwxyz1234"), 0o600))
	mustNoError(t, os.WriteFile(filepath.Join(root, "blob.bin"), []byte{0, 1, 2, 3, 4}, 0o600))

	findings, err := ScanDir(root)
	mustNoError(t, err)
	mustEmpty(t, findings)
}

func TestScanDir_ReportsGenericAssignments_Bad(t *testing.T) {
	root := t.TempDir()

	mustNoError(t, os.WriteFile(filepath.Join(root, "secrets.env"), []byte("client_secret: abcdefghijklmnop\n"), 0o600))

	findings, err := ScanDir(root)
	mustNoError(t, err)
	mustLen(t, findings, 1)
	mustEqual(t, "generic-secret-assignment", findings[0].Rule)
	mustEqual(t, 1, findings[0].Line)
	mustEqual(t, 1, findings[0].Column)
}
