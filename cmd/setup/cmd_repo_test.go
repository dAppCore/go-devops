package setup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunRepoSetup_CreatesCoreConfigs(t *testing.T) {
	dir := t.TempDir()
	mustNoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/test\n"), 0o644))

	mustNoError(t, runRepoSetup(dir, false))

	for _, name := range []string{"build.yaml", "release.yaml", "test.yaml"} {
		path := filepath.Join(dir, ".core", name)
		_, err := os.Stat(path)
		mustNoErrorf(t, err, "expected %s to exist", path)
	}
}

func TestDetectProjectType_PrefersPackageOverComposer(t *testing.T) {
	dir := t.TempDir()
	mustNoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}\n"), 0o644))
	mustNoError(t, os.WriteFile(filepath.Join(dir, "composer.json"), []byte("{}\n"), 0o644))

	mustEqual(t, "node", detectProjectType(dir))
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
			mustEqual(t, expected, parseGitHubRepoURL(remote))
		})
	}
}
