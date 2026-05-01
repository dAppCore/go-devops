package setup

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
)

func resetSetupCommand(t *core.T) {
	original := setupCmd
	setupCmd = &cli.Command{
		Use: "setup",
		RunE: func(cmd *cli.Command, args []string) error {
			return nil
		},
	}
	t.Cleanup(func() { setupCmd = original })
}

func setupCommand(root *cli.Command, use string) *cli.Command {
	for _, cmd := range root.Commands() {
		if cmd.Use == use || core.HasPrefix(cmd.Use, use+" ") {
			return cmd
		}
	}
	return nil
}

func TestCmdSetup_AddSetupCommand_Good(t *core.T) {
	resetSetupCommand(t)
	root := &cli.Command{Use: "root"}
	AddSetupCommand(root)
	cmd := setupCommand(root, "setup")

	core.AssertNotNil(t, cmd)
	core.AssertNotNil(t, cmd.Flag("registry"))
}

func TestCmdSetup_AddSetupCommand_Bad(t *core.T) {
	resetSetupCommand(t)
	var root *cli.Command
	core.AssertPanics(t, func() {
		AddSetupCommand(root)
	})
	core.AssertNil(t, root)
}

func TestCmdSetup_AddSetupCommand_Ugly(t *core.T) {
	resetSetupCommand(t)
	root := &cli.Command{Use: "root"}
	root.AddCommand(&cli.Command{Use: "existing"})
	AddSetupCommand(root)

	core.AssertLen(t, root.Commands(), 2)
	core.AssertNotNil(t, setupCommand(root, "setup"))
}
