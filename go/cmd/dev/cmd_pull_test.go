package dev

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
)

func TestCmdPull_AddPullCommand_Good(t *core.T) {
	root := &cli.Command{Use: "root"}
	AddPullCommand(root)
	cmd := testCommand(root, "pull")

	core.AssertNotNil(t, cmd)
	core.AssertNotNil(t, cmd.Flag("all"))
}

func TestCmdPull_AddPullCommand_Bad(t *core.T) {
	var root *cli.Command
	core.AssertPanics(t, func() {
		AddPullCommand(root)
	})
	core.AssertNil(t, root)
}

func TestCmdPull_AddPullCommand_Ugly(t *core.T) {
	root := &cli.Command{Use: "root"}
	AddPullCommand(root)
	AddPullCommand(root)

	core.AssertLen(t, root.Commands(), 2)
	core.AssertNotNil(t, testCommand(root, "pull"))
}
