package setup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunRepoSetup_CreatesCoreConfigs(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/test\n"), 0o644))

	require.NoError(t, runRepoSetup(dir, false))

	for _, name := range []string{"build.yaml", "release.yaml", "test.yaml"} {
		path := filepath.Join(dir, ".core", name)
		_, err := os.Stat(path)
		require.NoErrorf(t, err, "expected %s to exist", path)
	}
}

func TestDetectProjectType_PrefersPackageOverComposer(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "composer.json"), []byte("{}\n"), 0o644))

	require.Equal(t, "node", detectProjectType(dir))
}
