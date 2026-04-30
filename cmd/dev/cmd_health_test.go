package dev

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
)

func TestCmdHealth_AddHealthCommand_Good(t *core.T) {
	root := &cli.Command{Use: "root"}
	AddHealthCommand(root)
	cmd := testCommand(root, "health")

	core.AssertNotNil(t, cmd)
	core.AssertNotNil(t, cmd.Flag("verbose"))
}

func TestCmdHealth_AddHealthCommand_Bad(t *core.T) {
	var root *cli.Command
	core.AssertPanics(t, func() {
		AddHealthCommand(root)
	})
	core.AssertNil(t, root)
}

func TestCmdHealth_AddHealthCommand_Ugly(t *core.T) {
	root := &cli.Command{Use: "root"}
	AddHealthCommand(root)
	AddHealthCommand(root)

	core.AssertLen(t, root.Commands(), 2)
	core.AssertNotNil(t, testCommand(root, "health"))
}
