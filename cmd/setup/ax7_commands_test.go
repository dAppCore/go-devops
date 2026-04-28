package setup

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
)

func ax7ResetSetupCommand(t *core.T) {
	original := setupCmd
	setupCmd = &cli.Command{
		Use: "setup",
		RunE: func(cmd *cli.Command, args []string) error {
			return nil
		},
	}
	t.Cleanup(func() { setupCmd = original })
}

func ax7SetupCommand(root *cli.Command, use string) *cli.Command {
	for _, cmd := range root.Commands() {
		if cmd.Use == use || core.HasPrefix(cmd.Use, use+" ") {
			return cmd
		}
	}
	return nil
}

func TestAX7_AddSetupCommand_Good(t *core.T) {
	ax7ResetSetupCommand(t)
	root := &cli.Command{Use: "root"}
	AddSetupCommand(root)
	cmd := ax7SetupCommand(root, "setup")

	core.AssertNotNil(t, cmd)
	core.AssertNotNil(t, cmd.Flag("registry"))
}

func TestAX7_AddSetupCommand_Bad(t *core.T) {
	ax7ResetSetupCommand(t)
	var root *cli.Command
	core.AssertPanics(t, func() {
		AddSetupCommand(root)
	})
	core.AssertNil(t, root)
}

func TestAX7_AddSetupCommand_Ugly(t *core.T) {
	ax7ResetSetupCommand(t)
	root := &cli.Command{Use: "root"}
	root.AddCommand(&cli.Command{Use: "existing"})
	AddSetupCommand(root)

	core.AssertLen(t, root.Commands(), 2)
	core.AssertNotNil(t, ax7SetupCommand(root, "setup"))
}

func TestAX7_AddSetupCommands_Good(t *core.T) {
	ax7ResetSetupCommand(t)
	root := &cli.Command{Use: "root"}
	AddSetupCommands(root)
	cmd := ax7SetupCommand(root, "setup")

	core.AssertNotNil(t, cmd)
	core.AssertNotEmpty(t, cmd.Short)
}

func TestAX7_AddSetupCommands_Bad(t *core.T) {
	ax7ResetSetupCommand(t)
	var root *cli.Command
	core.AssertPanics(t, func() {
		AddSetupCommands(root)
	})
	core.AssertNil(t, root)
}

func TestAX7_AddSetupCommands_Ugly(t *core.T) {
	ax7ResetSetupCommand(t)
	root := &cli.Command{Use: "root"}
	root.AddCommand(&cli.Command{Use: "existing"})
	AddSetupCommands(root)

	core.AssertLen(t, root.Commands(), 2)
	core.AssertNotNil(t, ax7SetupCommand(root, "setup"))
}
