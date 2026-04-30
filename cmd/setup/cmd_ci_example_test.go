package setup

import . "dappco.re/go"

func ExampleDefaultCIConfig() {
	cfg := DefaultCIConfig()
	Println(cfg.Formula, cfg.DefaultVersion)
	// Output: core dev
}

func ExampleLoadCIConfig() {
	dir := MustCast[string](MkdirTemp("", "ci-config-*"))
	defer RemoveAll(dir)
	old := MustCast[string](Getwd())
	defer Chdir(old)
	MkdirAll(PathJoin(dir, ".core"), 0o755)
	WriteFile(PathJoin(dir, ".core", "ci.yaml"), []byte("formula: devops\ndefault_version: v1\n"), 0o600)
	Chdir(dir)

	cfg := LoadCIConfig()
	Println(cfg.Formula, cfg.DefaultVersion)
	// Output: devops v1
}
