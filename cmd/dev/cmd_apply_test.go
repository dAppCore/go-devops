package dev

import (
	"testing"

	"dappco.re/go/scm/repos"
)

func TestFilterTargetRepos_Good(t *testing.T) {
	registry := &repos.Registry{
		Repos: map[string]*repos.Repo{
			"core-api":  &repos.Repo{Name: "core-api", Path: "packages/core-api"},
			"core-web":  &repos.Repo{Name: "core-web", Path: "packages/core-web"},
			"docs-site": &repos.Repo{Name: "docs-site", Path: "sites/docs"},
		},
	}

	t.Run("exact names", func(t *testing.T) {
		matched := filterTargetRepos(registry, "core-api,docs-site")
		if len(matched) != 2 {
			t.Fatalf("matched length = %d, want 2", len(matched))
		}
		if matched[0].Name != "core-api" {
			t.Fatalf("matched[0].Name = %q, want %q", matched[0].Name, "core-api")
		}
		if matched[1].Name != "docs-site" {
			t.Fatalf("matched[1].Name = %q, want %q", matched[1].Name, "docs-site")
		}
	})

	t.Run("glob patterns", func(t *testing.T) {
		matched := filterTargetRepos(registry, "core-*,sites/*")
		if len(matched) != 3 {
			t.Fatalf("matched length = %d, want 3", len(matched))
		}
		wantNames := []string{"core-api", "core-web", "docs-site"}
		for i, want := range wantNames {
			if matched[i].Name != want {
				t.Fatalf("matched[%d].Name = %q, want %q", i, matched[i].Name, want)
			}
		}
	})

	t.Run("all repos when empty", func(t *testing.T) {
		matched := filterTargetRepos(registry, "")
		if len(matched) != 3 {
			t.Fatalf("matched length = %d, want 3", len(matched))
		}
	})
}
