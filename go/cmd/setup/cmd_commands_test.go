package setup

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
)

func TestCmdCommands_AddSetupCommands_Good(t *core.T) {
	resetSetupCommand(t)
	root := &cli.Command{Use: "root"}
	AddSetupCommands(root)
	cmd := setupCommand(root, "setup")

	core.AssertNotNil(t, cmd)
	core.AssertNotEmpty(t, cmd.Short)
}

func TestCmdCommands_AddSetupCommands_Bad(t *core.T) {
	resetSetupCommand(t)
	var root *cli.Command
	core.AssertPanics(t, func() {
		AddSetupCommands(root)
	})
	core.AssertNil(t, root)
}

func TestCmdCommands_AddSetupCommands_Ugly(t *core.T) {
	resetSetupCommand(t)
	root := &cli.Command{Use: "root"}
	root.AddCommand(&cli.Command{Use: "existing"})
	AddSetupCommands(root)

	core.AssertLen(t, root.Commands(), 2)
	core.AssertNotNil(t, setupCommand(root, "setup"))
}
