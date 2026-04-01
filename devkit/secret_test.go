package devkit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScanDir_Good(t *testing.T) {
	root := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(root, "config.yml"), []byte(`
api_key: "ghp_abcdefghijklmnopqrstuvwxyz1234"
`), 0o600))

	require.NoError(t, os.Mkdir(filepath.Join(root, "nested"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "nested", "creds.txt"), []byte("access_key = AKIA1234567890ABCDEF\n"), 0o600))

	findings, err := ScanDir(root)
	require.NoError(t, err)
	require.Len(t, findings, 2)

	require.Equal(t, "github-token", findings[0].Rule)
	require.Equal(t, 2, findings[0].Line)
	require.Equal(t, "config.yml", filepath.Base(findings[0].Path))

	require.Equal(t, "aws-access-key-id", findings[1].Rule)
	require.Equal(t, 1, findings[1].Line)
	require.Equal(t, "creds.txt", filepath.Base(findings[1].Path))
}

func TestScanDir_SkipsBinaryAndIgnoredDirs(t *testing.T) {
	root := t.TempDir()

	require.NoError(t, os.Mkdir(filepath.Join(root, ".git"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, ".git", "config"), []byte("token=ghp_abcdefghijklmnopqrstuvwxyz1234"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "blob.bin"), []byte{0, 1, 2, 3, 4}, 0o600))

	findings, err := ScanDir(root)
	require.NoError(t, err)
	require.Empty(t, findings)
}

func TestScanDir_ReportsGenericAssignments(t *testing.T) {
	root := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(root, "secrets.env"), []byte("client_secret: abcdefghijklmnop\n"), 0o600))

	findings, err := ScanDir(root)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	require.Equal(t, "generic-secret-assignment", findings[0].Rule)
	require.Equal(t, 1, findings[0].Line)
	require.Equal(t, 1, findings[0].Column)
}
