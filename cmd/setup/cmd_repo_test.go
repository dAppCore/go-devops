package setup

import (
	. "dappco.re/go"
	"testing"
)

func TestRunRepoSetupCreatesCoreConfigs(t *testing.T) {
	dir := t.TempDir()
	if r := WriteFile(PathJoin(dir, "go.mod"), []byte("module example.com/test\n"), 0o644); !r.OK {
		t.Fatalf("write go.mod: %v", r.Error())
	}

	if err := runRepoSetup(dir, false); !err.OK {
		t.Fatalf("run repo setup: %v", err)
	}

	for _, name := range []string{"build.yaml", "release.yaml", "test.yaml"} {
		path := PathJoin(dir, ".core", name)
		if r := Stat(path); !r.OK {
			t.Fatalf("expected %s to exist: %v", path, r.Error())
		}
	}
}

func TestDetectProjectTypePrefersPackageOverComposer(t *testing.T) {
	dir := t.TempDir()
	if r := WriteFile(PathJoin(dir, "package.json"), []byte("{}\n"), 0o644); !r.OK {
		t.Fatalf("write package.json: %v", r.Error())
	}
	if r := WriteFile(PathJoin(dir, "composer.json"), []byte("{}\n"), 0o644); !r.OK {
		t.Fatalf("write composer.json: %v", r.Error())
	}

	if got := detectProjectType(dir); got != "node" {
		t.Fatalf("detectProjectType = %q, want %q", got, "node")
	}
}

func TestParseGitHubRepoURL_Good(t *testing.T) {
	cases := map[string]string{
		"git@github.com:owner/repo.git":            "owner/repo",
		"ssh://git@github.com/owner/repo.git":      "owner/repo",
		"https://github.com/owner/repo.git":        "owner/repo",
		"git://github.com/owner/repo.git":          "owner/repo",
		"https://www.github.com/owner/repo":        "owner/repo",
		"git@github.com:owner/nested/repo.git":     "owner/nested/repo",
		"ssh://git@github.com/owner/nested/repo/":  "owner/nested/repo",
		"ssh://git@github.com:443/owner/repo.git":  "owner/repo",
		"https://example.com/owner/repo.git":       "",
		"git@bitbucket.org:owner/repo.git":         "",
		"   ssh://git@github.com/owner/repo.git  ": "owner/repo",
	}

	for remote, expected := range cases {
		t.Run(remote, func(t *testing.T) {
			if got := parseGitHubRepoURL(remote); got != expected {
				t.Fatalf("parseGitHubRepoURL(%q) = %q, want %q", remote, got, expected)
			}
		})
	}
}
