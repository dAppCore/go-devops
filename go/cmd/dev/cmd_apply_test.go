package dev

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
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
	root := &cli.Command{Use: "root"}
	AddApplyCommand(root)
	cmd := testCommand(root, "apply")

	core.AssertNotNil(t, cmd)
	core.AssertNotNil(t, cmd.Flag("command"))
}

func TestCmdApply_AddApplyCommand_Bad(t *core.T) {
	var root *cli.Command
	core.AssertPanics(t, func() {
		AddApplyCommand(root)
	})
	core.AssertNil(t, root)
}

func TestCmdApply_AddApplyCommand_Ugly(t *core.T) {
	root := &cli.Command{Use: "root"}
	AddApplyCommand(root)
	AddApplyCommand(root)

	core.AssertLen(t, root.Commands(), 2)
	core.AssertNotNil(t, testCommand(root, "apply"))
}
