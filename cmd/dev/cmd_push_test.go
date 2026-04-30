package dev

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
)

func TestCmdPush_AddPushCommand_Good(t *core.T) {
	root := &cli.Command{Use: "root"}
	AddPushCommand(root)
	cmd := testCommand(root, "push")

	core.AssertNotNil(t, cmd)
	core.AssertNotNil(t, cmd.Flag("force"))
}

func TestCmdPush_AddPushCommand_Bad(t *core.T) {
	var root *cli.Command
	core.AssertPanics(t, func() {
		AddPushCommand(root)
	})
	core.AssertNil(t, root)
}

func TestCmdPush_AddPushCommand_Ugly(t *core.T) {
	root := &cli.Command{Use: "root"}
	root.AddCommand(&cli.Command{Use: "existing"})
	AddPushCommand(root)

	core.AssertLen(t, root.Commands(), 2)
	core.AssertNotNil(t, testCommand(root, "push"))
}
