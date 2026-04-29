package setup

import . "dappco.re/go"

func ExampleLoadGitHubConfig() {
	dir := MustCast[string](MkdirTemp("", "github-config-*"))
	defer RemoveAll(dir)
	path := PathJoin(dir, "github.yaml")
	WriteFile(path, []byte("version: 1\nlabels:\n  - name: bug\n    color: ff0000\n"), 0o600)

	cfg, err := LoadGitHubConfig(path)
	Println(err == nil, cfg.Labels[0].Name)
	// Output: true bug
}

func ExampleFindGitHubConfig() {
	dir := MustCast[string](MkdirTemp("", "github-find-*"))
	defer RemoveAll(dir)
	MkdirAll(PathJoin(dir, ".core"), 0o755)
	path := PathJoin(dir, ".core", "github.yaml")
	WriteFile(path, []byte("version: 1\n"), 0o600)

	found, err := FindGitHubConfig(dir, "")
	Println(err == nil, found == path)
	// Output: true true
}

func ExampleGitHubConfig_Validate() {
	cfg := &GitHubConfig{Version: 1, Labels: []LabelConfig{{Name: "bug", Color: "ff0000"}}}
	err := cfg.Validate()
	Println(err == nil)
	// Output: true
}
