package setup

import (
	core "dappco.re/go"
)

func TestCmdCi_DefaultCIConfig_Good(t *core.T) {
	cfg := DefaultCIConfig()
	core.AssertEqual(t, "host-uk/tap", cfg.Tap)

	core.AssertEqual(t, "core", cfg.Formula)
	core.AssertEqual(t, "dev", cfg.DefaultVersion)
}

func TestCmdCi_DefaultCIConfig_Bad(t *core.T) {
	cfg := DefaultCIConfig()
	cfg.DefaultVersion = ""

	core.AssertEqual(t, "", cfg.DefaultVersion)
	core.AssertEqual(t, "core-cli", cfg.ChocolateyPkg)
}

func TestCmdCi_DefaultCIConfig_Ugly(t *core.T) {
	first := DefaultCIConfig()
	second := DefaultCIConfig()
	first.Formula = "mutated"

	core.AssertEqual(t, "mutated", first.Formula)
	core.AssertEqual(t, "core", second.Formula)
}

func TestCmdCi_LoadCIConfig_Good(t *core.T) {
	dir := t.TempDir()
	core.RequireTrue(t, core.MkdirAll(core.Path(dir, ".core"), 0o755).OK)
	core.RequireTrue(t, core.WriteFile(core.Path(dir, ".core", "ci.yaml"), []byte("formula: devops\ndefault_version: v1\n"), 0o644).OK)
	t.Chdir(dir)

	cfg := LoadCIConfig()
	core.AssertEqual(t, "devops", cfg.Formula)
	core.AssertEqual(t, "v1", cfg.DefaultVersion)
}

func TestCmdCi_LoadCIConfig_Bad(t *core.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	cfg := LoadCIConfig()

	core.AssertEqual(t, "core", cfg.Formula)
	core.AssertEqual(t, "dev", cfg.DefaultVersion)
}

func TestCmdCi_LoadCIConfig_Ugly(t *core.T) {
	dir := t.TempDir()
	child := core.Path(dir, "a", "b")
	core.RequireTrue(t, core.MkdirAll(core.Path(dir, ".core"), 0o755).OK)
	core.RequireTrue(t, core.MkdirAll(child, 0o755).OK)
	core.RequireTrue(t, core.WriteFile(core.Path(dir, ".core", "ci.yaml"), []byte("tap: custom/tap\n"), 0o644).OK)
	t.Chdir(child)

	cfg := LoadCIConfig()
	core.AssertEqual(t, "custom/tap", cfg.Tap)
	core.AssertEqual(t, "core", cfg.Formula)
}
