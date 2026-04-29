package devkit

import (
	. "dappco.re/go"
)

func TestSecret_ScanDir_Good(t *T) {
	dir := t.TempDir()
	RequireTrue(t, WriteFile(Path(dir, "config.env"), []byte("API_KEY=abcdefghijk\n"), 0o600).OK)
	findings, err := ScanDir(dir)

	AssertNoError(t, err)
	AssertEqual(t, "generic-secret-assignment", findings[0].Rule)
}

func TestSecret_ScanDir_Bad(t *T) {
	findings, err := ScanDir(Path(t.TempDir(), "missing"))
	AssertError(t, err)

	AssertNil(t, findings)
	AssertContains(t, err.Error(), "no such file")
}

func TestSecret_ScanDir_Ugly(t *T) {
	dir := t.TempDir()
	RequireTrue(t, MkdirAll(Path(dir, ".git"), 0o755).OK)
	RequireTrue(t, WriteFile(Path(dir, ".git", "secret.env"), []byte("API_KEY=abcdefghijk\n"), 0o600).OK)

	findings, err := ScanDir(dir)
	AssertNoError(t, err)
	AssertEmpty(t, findings)
}
