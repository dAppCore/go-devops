package dev

import (
	core "dappco.re/go"
	"dappco.re/go/scm/repos"
	"testing"
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

func TestCmdApply_AddApplyCommand_Good(t *core.T) {
	c := core.New()
	r := AddApplyCommand(c, "dev")
	core.AssertTrue(t, r.OK)

	cmd := c.Command("dev/apply")
	core.AssertTrue(t, cmd.OK)
	core.AssertNotNil(t, cmd.Value.(*core.Command).Action)
}

func TestCmdApply_AddApplyCommand_Bad(t *core.T) {
	var c *core.Core
	core.AssertPanics(t, func() {
		AddApplyCommand(c, "dev")
	})
	core.AssertNil(t, c)
}

func TestCmdApply_AddApplyCommand_Ugly(t *core.T) {
	c := core.New()
	core.AssertTrue(t, AddApplyCommand(c, "dev").OK)

	r := AddApplyCommand(c, "git")
	core.AssertTrue(t, r.OK)
	core.AssertTrue(t, c.Command("git/apply").OK)
}
