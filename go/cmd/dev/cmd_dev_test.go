package dev

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
)

func testCommand(root *cli.Command, use string) *cli.Command {
	for _, cmd := range root.Commands() {
		if cmd.Use == use || core.HasPrefix(cmd.Use, use+" ") {
			return cmd
		}
	}
	return nil
}

func TestCmdDev_AddDevCommands_Good(t *core.T) {
	root := &cli.Command{Use: "root"}
	AddDevCommands(root)
	devCmd := testCommand(root, "dev")

	core.AssertNotNil(t, devCmd)
	core.AssertGreaterOrEqual(t, len(devCmd.Commands()), 10)
}

func TestCmdDev_AddDevCommands_Bad(t *core.T) {
	var root *cli.Command
	core.AssertPanics(t, func() {
		AddDevCommands(root)
	})
	core.AssertNil(t, root)
}

func TestCmdDev_AddDevCommands_Ugly(t *core.T) {
	root := &cli.Command{Use: "root"}
	root.AddCommand(&cli.Command{Use: "existing"})
	AddDevCommands(root)

	core.AssertLen(t, root.Commands(), 2)
	core.AssertNotNil(t, testCommand(root, "dev"))
}
