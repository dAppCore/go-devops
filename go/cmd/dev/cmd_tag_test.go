package dev

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
)

func TestCmdTag_AddTagCommand_Good(t *core.T) {
	root := &cli.Command{Use: "root"}
	AddTagCommand(root)
	cmd := testCommand(root, "tag")

	core.AssertNotNil(t, cmd)
	core.AssertNotNil(t, cmd.Flag("dry-run"))
}

func TestCmdTag_AddTagCommand_Bad(t *core.T) {
	var root *cli.Command
	core.AssertPanics(t, func() {
		AddTagCommand(root)
	})
	core.AssertNil(t, root)
}

func TestCmdTag_AddTagCommand_Ugly(t *core.T) {
	root := &cli.Command{Use: "root"}
	root.AddCommand(&cli.Command{Use: "existing"})
	AddTagCommand(root)

	core.AssertLen(t, root.Commands(), 2)
	core.AssertNotNil(t, testCommand(root, "tag"))
}
