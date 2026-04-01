package dev

import (
	"testing"

	"github.com/stretchr/testify/require"

	"dappco.re/go/core/scm/repos"
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
		require.Len(t, matched, 2)
		require.Equal(t, "core-api", matched[0].Name)
		require.Equal(t, "docs-site", matched[1].Name)
	})

	t.Run("glob patterns", func(t *testing.T) {
		matched := filterTargetRepos(registry, "core-*,sites/*")
		require.Len(t, matched, 3)
		require.Equal(t, "core-api", matched[0].Name)
		require.Equal(t, "core-web", matched[1].Name)
		require.Equal(t, "docs-site", matched[2].Name)
	})

	t.Run("all repos when empty", func(t *testing.T) {
		matched := filterTargetRepos(registry, "")
		require.Len(t, matched, 3)
	})
}
