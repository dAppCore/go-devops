package setup

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}
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

	if err := w.Close(); err != nil {
		t.Fatalf("close pipe writer: %v", err)
	}
	if err := <-errC; err != nil {
		t.Fatalf("copy stdout: %v", err)
	}
	out := <-outC

	return out, runErr
}

func TestDefaultCIConfig_Good(t *testing.T) {
	cfg := DefaultCIConfig()

	checks := map[string]struct {
		got  string
		want string
	}{
		"tap":             {got: cfg.Tap, want: "host-uk/tap"},
		"formula":         {got: cfg.Formula, want: "core"},
		"scoop bucket":    {got: cfg.ScoopBucket, want: "https://forge.lthn.ai/core/scoop-bucket.git"},
		"chocolatey pkg":  {got: cfg.ChocolateyPkg, want: "core-cli"},
		"repository":      {got: cfg.Repository, want: "host-uk/core"},
		"default version": {got: cfg.DefaultVersion, want: "dev"},
	}
	for name, check := range checks {
		if check.got != check.want {
			t.Fatalf("%s = %q, want %q", name, check.got, check.want)
		}
	}
}

func TestOutputPowershellInstall_Good(t *testing.T) {
	out, err := captureStdout(t, func() error {
		return outputPowershellInstall(DefaultCIConfig(), "dev")
	})
	if err != nil {
		t.Fatalf("output powershell install: %v", err)
	}
	if !strings.Contains(out, `scoop bucket add host-uk $ScoopBucket`) {
		t.Fatalf("output missing scoop bucket command: %q", out)
	}
	if strings.Contains(out, `https://https://forge.lthn.ai/core/scoop-bucket.git`) {
		t.Fatalf("output contains doubled URL scheme: %q", out)
	}
}
