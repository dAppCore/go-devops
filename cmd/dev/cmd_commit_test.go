package dev

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
)

func TestCmdCommit_AddCommitCommand_Good(t *core.T) {
	root := &cli.Command{Use: "root"}
	AddCommitCommand(root)
	cmd := testCommand(root, "commit")

	core.AssertNotNil(t, cmd)
	core.AssertNotNil(t, cmd.Flag("all"))
}

func TestCmdCommit_AddCommitCommand_Bad(t *core.T) {
	var root *cli.Command
	core.AssertPanics(t, func() {
		AddCommitCommand(root)
	})
	core.AssertNil(t, root)
}

func TestCmdCommit_AddCommitCommand_Ugly(t *core.T) {
	root := &cli.Command{Use: "root"}
	root.AddCommand(&cli.Command{Use: "existing"})
	AddCommitCommand(root)

	core.AssertLen(t, root.Commands(), 2)
	core.AssertNotNil(t, testCommand(root, "commit"))
}
