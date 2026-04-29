package workspace

import (
	. "dappco.re/go"
)

func TestConfig_DefaultConfig_Good(t *T) {
	cfg := DefaultConfig()

	AssertEqual(t, 1, cfg.Version)
	AssertEqual(t, "./packages", cfg.PackagesDir)
}

func TestConfig_DefaultConfig_Bad(t *T) {
	cfg := DefaultConfig()

	AssertEqual(t, "", cfg.Active)
	AssertEmpty(t, cfg.DefaultOnly)
}

func TestConfig_DefaultConfig_Ugly(t *T) {
	first := DefaultConfig()
	second := DefaultConfig()
	first.PackagesDir = "changed"

	AssertEqual(t, "./packages", second.PackagesDir)
	AssertNotEqual(t, first.PackagesDir, second.PackagesDir)
}

func TestConfig_LoadConfig_Good(t *T) {
	dir := t.TempDir()
	RequireTrue(t, MkdirAll(Path(dir, ".core"), 0o755).OK)
	RequireTrue(t, WriteFile(Path(dir, ".core", "workspace.yaml"), []byte("version: 1\nactive: devops\npackages_dir: repos\n"), 0o644).OK)

	cfg, err := LoadConfig(dir)
	AssertNoError(t, err)
	AssertEqual(t, "devops", cfg.Active)
	AssertEqual(t, "repos", cfg.PackagesDir)
}

func TestConfig_LoadConfig_Bad(t *T) {
	cfg, err := LoadConfig(t.TempDir())
	AssertNoError(t, err)

	AssertNil(t, cfg)
	AssertNoError(t, err)
}

func TestConfig_LoadConfig_Ugly(t *T) {
	dir := t.TempDir()
	child := Path(dir, "a", "b")
	RequireTrue(t, MkdirAll(Path(dir, ".core"), 0o755).OK)
	RequireTrue(t, MkdirAll(child, 0o755).OK)
	RequireTrue(t, WriteFile(Path(dir, ".core", "workspace.yaml"), []byte("version: 1\npackages_dir: nested\n"), 0o644).OK)

	cfg, err := LoadConfig(child)
	AssertNoError(t, err)
	AssertEqual(t, "nested", cfg.PackagesDir)
}

func TestConfig_FindRoot_Good(t *T) {
	dir := t.TempDir()
	RequireTrue(t, MkdirAll(Path(dir, ".core"), 0o755).OK)
	RequireTrue(t, WriteFile(Path(dir, ".core", "workspace.yaml"), []byte("version: 1\n"), 0o644).OK)
	t.Chdir(dir)

	root, err := FindRoot()
	AssertNoError(t, err)
	AssertEqual(t, dir, root)
}

func TestConfig_FindRoot_Bad(t *T) {
	dir := t.TempDir()
	t.Chdir(dir)

	root, err := FindRoot()
	AssertError(t, err)
	AssertEqual(t, "", root)
}

func TestConfig_FindRoot_Ugly(t *T) {
	dir := t.TempDir()
	child := Path(dir, "nested", "deep")
	RequireTrue(t, MkdirAll(Path(dir, ".core"), 0o755).OK)
	RequireTrue(t, MkdirAll(child, 0o755).OK)
	RequireTrue(t, WriteFile(Path(dir, ".core", "workspace.yaml"), []byte("version: 1\n"), 0o644).OK)
	t.Chdir(child)

	root, err := FindRoot()
	AssertNoError(t, err)
	AssertEqual(t, dir, root)
}
