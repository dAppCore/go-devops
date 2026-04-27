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
		mustLen(t, matched, 2)
		mustEqual(t, "core-api", matched[0].Name)
		mustEqual(t, "docs-site", matched[1].Name)
	})

	t.Run("glob patterns", func(t *testing.T) {
		matched := filterTargetRepos(registry, "core-*,sites/*")
		mustLen(t, matched, 3)
		mustEqual(t, "core-api", matched[0].Name)
		mustEqual(t, "core-web", matched[1].Name)
		mustEqual(t, "docs-site", matched[2].Name)
	})

	t.Run("all repos when empty", func(t *testing.T) {
		matched := filterTargetRepos(registry, "")
		mustLen(t, matched, 3)
	})
}
