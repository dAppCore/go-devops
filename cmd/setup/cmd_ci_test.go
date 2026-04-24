package setup

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	mustNoError(t, err)
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

	mustNoError(t, w.Close())
	mustNoError(t, <-errC)
	out := <-outC

	return out, runErr
}

func TestDefaultCIConfig_Good(t *testing.T) {
	cfg := DefaultCIConfig()

	mustEqual(t, "host-uk/tap", cfg.Tap)
	mustEqual(t, "core", cfg.Formula)
	mustEqual(t, "https://forge.lthn.ai/core/scoop-bucket.git", cfg.ScoopBucket)
	mustEqual(t, "core-cli", cfg.ChocolateyPkg)
	mustEqual(t, "host-uk/core", cfg.Repository)
	mustEqual(t, "dev", cfg.DefaultVersion)
}

func TestOutputPowershellInstall_Good(t *testing.T) {
	out, err := captureStdout(t, func() error {
		return outputPowershellInstall(DefaultCIConfig(), "dev")
	})
	mustNoError(t, err)
	mustContains(t, out, `scoop bucket add host-uk $ScoopBucket`)
	mustNotContains(t, out, `https://https://forge.lthn.ai/core/scoop-bucket.git`)
}
