package setup

import (
	core "dappco.re/go"
)

func TestGithubConfig_LoadGitHubConfig_Good(t *core.T) {
	path := core.Path(t.TempDir(), "github.yaml")
	t.Setenv("HOOK_URL", "https://hooks.example")
	core.RequireTrue(t, core.WriteFile(path, []byte("version: 1\nlabels:\n- name: bug\n  color: ff0000\nwebhooks:\n  ci:\n    url: ${HOOK_URL}\n    events: [push]\n"), 0o644).OK)

	cfg, err := LoadGitHubConfig(path)
	core.AssertNoError(t, err)
	core.AssertEqual(t, "https://hooks.example", cfg.Webhooks["ci"].URL)
}

func TestGithubConfig_LoadGitHubConfig_Bad(t *core.T) {
	cfg, err := LoadGitHubConfig(core.Path(t.TempDir(), "missing.yaml"))
	core.AssertError(t, err)

	core.AssertNil(t, cfg)
	core.AssertContains(t, err.Error(), "failed to read")
}

func TestGithubConfig_LoadGitHubConfig_Ugly(t *core.T) {
	path := core.Path(t.TempDir(), "github.yaml")
	core.RequireTrue(t, core.WriteFile(path, []byte("version: 1\nwebhooks:\n  ci:\n    url: ${MISSING_URL:-https://fallback.example}\n    events: [push]\n"), 0o644).OK)

	cfg, err := LoadGitHubConfig(path)
	core.AssertNoError(t, err)
	core.AssertEqual(t, "https://fallback.example", cfg.Webhooks["ci"].URL)
}

func TestGithubConfig_FindGitHubConfig_Good(t *core.T) {
	dir := t.TempDir()
	path := core.Path(dir, ".core", "github.yaml")
	core.RequireTrue(t, core.MkdirAll(core.Path(dir, ".core"), 0o755).OK)
	core.RequireTrue(t, core.WriteFile(path, []byte("version: 1\n"), 0o644).OK)

	got, err := FindGitHubConfig(dir, "")
	core.AssertNoError(t, err)
	core.AssertEqual(t, path, got)
}

func TestGithubConfig_FindGitHubConfig_Bad(t *core.T) {
	got, err := FindGitHubConfig(t.TempDir(), "missing.yaml")
	core.AssertError(t, err)

	core.AssertEqual(t, "", got)
	core.AssertContains(t, err.Error(), "config file not found")
}

func TestGithubConfig_FindGitHubConfig_Ugly(t *core.T) {
	dir := t.TempDir()
	path := core.Path(dir, "github.yaml")
	core.RequireTrue(t, core.WriteFile(path, []byte("version: 1\n"), 0o644).OK)

	got, err := FindGitHubConfig(dir, "")
	core.AssertNoError(t, err)
	core.AssertEqual(t, path, got)
}

func TestGithubConfig_GitHubConfig_Validate_Good(t *core.T) {
	cfg := &GitHubConfig{Version: 1, Labels: []LabelConfig{{Name: "bug", Color: "ff0000"}}}
	err := cfg.Validate()

	core.AssertNoError(t, err)
	core.AssertEqual(t, 1, cfg.Version)
}

func TestGithubConfig_GitHubConfig_Validate_Bad(t *core.T) {
	cfg := &GitHubConfig{Version: 2}
	err := cfg.Validate()

	core.AssertError(t, err)
	core.AssertContains(t, err.Error(), "unsupported")
}

func TestGithubConfig_GitHubConfig_Validate_Ugly(t *core.T) {
	cfg := &GitHubConfig{Version: 1, Labels: []LabelConfig{{Name: "bug", Color: "nope"}}}
	err := cfg.Validate()

	core.AssertError(t, err)
	core.AssertContains(t, err.Error(), "invalid color")
}
