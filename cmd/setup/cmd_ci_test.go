package setup

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	defer func() {
		_ = r.Close()
	}()

	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()

	outC := make(chan string, 1)
	errC := make(chan error, 1)

	go func() {
		var buf bytes.Buffer
		_, copyErr := io.Copy(&buf, r)
		errC <- copyErr
		outC <- buf.String()
	}()

	runErr := fn()

	require.NoError(t, w.Close())
	require.NoError(t, <-errC)
	out := <-outC

	return out, runErr
}

func TestDefaultCIConfig_Good(t *testing.T) {
	cfg := DefaultCIConfig()

	require.Equal(t, "host-uk/tap", cfg.Tap)
	require.Equal(t, "core", cfg.Formula)
	require.Equal(t, "https://forge.lthn.ai/core/scoop-bucket.git", cfg.ScoopBucket)
	require.Equal(t, "core-cli", cfg.ChocolateyPkg)
	require.Equal(t, "host-uk/core", cfg.Repository)
	require.Equal(t, "dev", cfg.DefaultVersion)
}

func TestOutputPowershellInstall_Good(t *testing.T) {
	out, err := captureStdout(t, func() error {
		return outputPowershellInstall(DefaultCIConfig(), "dev")
	})
	require.NoError(t, err)
	require.Contains(t, out, `scoop bucket add host-uk https://forge.lthn.ai/core/scoop-bucket.git`)
	require.NotContains(t, out, `https://https://forge.lthn.ai/core/scoop-bucket.git`)
}
