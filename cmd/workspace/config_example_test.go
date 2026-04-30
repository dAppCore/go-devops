package workspace

import . "dappco.re/go"

func ExampleDefaultConfig() {
	cfg := DefaultConfig()
	Println(cfg.Version, cfg.PackagesDir)
	// Output: 1 ./packages
}

func ExampleLoadConfig() {
	dir := MustCast[string](MkdirTemp("", "workspace-example-*"))
	defer RemoveAll(dir)
	MkdirAll(PathJoin(dir, ".core"), 0o755)
	WriteFile(PathJoin(dir, ".core", "workspace.yaml"), []byte("version: 1\nactive: app\npackages_dir: repos\n"), 0o600)

	cfg, r := LoadConfig(dir)
	Println(r.OK, cfg.Active)
	// Output: true app
}

func ExampleFindRoot() {
	dir := MustCast[string](MkdirTemp("", "workspace-root-*"))
	defer RemoveAll(dir)
	old := MustCast[string](Getwd())
	defer Chdir(old)
	MkdirAll(PathJoin(dir, ".core"), 0o755)
	WriteFile(PathJoin(dir, ".core", "workspace.yaml"), []byte("version: 1\n"), 0o600)
	Chdir(dir)

	root, r := FindRoot()
	Println(r.OK, root != "")
	// Output: true true
}
