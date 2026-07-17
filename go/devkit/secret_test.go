package devkit

import (
	core "dappco.re/go"
)

func TestSecret_ScanDir_Good(t *core.T) {
	dir := t.TempDir()
	core.RequireTrue(t, core.WriteFile(core.Path(dir, "config.env"), []byte("API_KEY=abcdefghijk\n"), 0o600).OK)
	findings, r := ScanDir(dir)

	core.AssertTrue(t, r.OK)
	core.AssertEqual(t, "generic-secret-assignment", findings[0].Rule)
}

func TestSecret_ScanDir_Bad(t *core.T) {
	findings, r := ScanDir(core.Path(t.TempDir(), "missing"))
	core.AssertFalse(t, r.OK)

	core.AssertNil(t, findings)
	core.AssertContains(t, r.Error(), "no such file")
}

func TestSecret_ScanDir_Ugly(t *core.T) {
	dir := t.TempDir()
	core.RequireTrue(t, core.MkdirAll(core.Path(dir, ".git"), 0o755).OK)
	core.RequireTrue(t, core.WriteFile(core.Path(dir, ".git", "secret.env"), []byte("API_KEY=abcdefghijk\n"), 0o600).OK)

	findings, r := ScanDir(dir)
	core.AssertTrue(t, r.OK)
	core.AssertEmpty(t, findings)
}
