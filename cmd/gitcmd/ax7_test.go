package gitcmd

import (
	. "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
)

func TestAX7_AddGitCommands_Good(t *T) {
	root := &cli.Command{Use: "root"}
	AddGitCommands(root)
	gitCmd := root.Commands()[0]

	AssertEqual(t, "git", gitCmd.Use)
	AssertGreaterOrEqual(t, len(gitCmd.Commands()), 7)
}

func TestAX7_AddGitCommands_Bad(t *T) {
	var root *cli.Command
	AssertPanics(t, func() {
		AddGitCommands(root)
	})
	AssertNil(t, root)
}

func TestAX7_AddGitCommands_Ugly(t *T) {
	root := &cli.Command{Use: "root"}
	root.AddCommand(&cli.Command{Use: "existing"})
	AddGitCommands(root)

	foundExisting := false
	foundGit := false
	for _, cmd := range root.Commands() {
		foundExisting = foundExisting || cmd.Use == "existing"
		foundGit = foundGit || cmd.Use == "git"
	}
	AssertLen(t, root.Commands(), 2)
	AssertTrue(t, foundExisting)
	AssertTrue(t, foundGit)
}
