package dev

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
)

func TestCmdWork_AddWorkCommand_Good(t *core.T) {
	root := &cli.Command{Use: "root"}
	AddWorkCommand(root)
	cmd := testCommand(root, "work")

	core.AssertNotNil(t, cmd)
	core.AssertNotNil(t, cmd.Flag("status"))
}

func TestCmdWork_AddWorkCommand_Bad(t *core.T) {
	var root *cli.Command
	core.AssertPanics(t, func() {
		AddWorkCommand(root)
	})
	core.AssertNil(t, root)
}

func TestCmdWork_AddWorkCommand_Ugly(t *core.T) {
	root := &cli.Command{Use: "root"}
	AddWorkCommand(root)
	AddWorkCommand(root)

	core.AssertLen(t, root.Commands(), 2)
	core.AssertNotNil(t, testCommand(root, "work"))
}
