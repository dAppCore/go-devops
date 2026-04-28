package deploy

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
)

func TestAX7_AddDeployCommands_Good(t *core.T) {
	root := &cli.Command{Use: "root"}
	AddDeployCommands(root)
	commands := root.Commands()

	core.AssertLen(t, commands, 1)
	core.AssertEqual(t, "deploy", commands[0].Use)
}

func TestAX7_AddDeployCommands_Bad(t *core.T) {
	var root *cli.Command
	core.AssertPanics(t, func() {
		AddDeployCommands(root)
	})
	core.AssertNil(t, root)
}

func TestAX7_AddDeployCommands_Ugly(t *core.T) {
	root := &cli.Command{Use: "root"}
	root.AddCommand(&cli.Command{Use: "existing"})
	AddDeployCommands(root)

	foundExisting := false
	foundDeploy := false
	for _, cmd := range root.Commands() {
		foundExisting = foundExisting || cmd.Use == "existing"
		foundDeploy = foundDeploy || cmd.Use == "deploy"
	}
	core.AssertLen(t, root.Commands(), 2)
	core.AssertTrue(t, foundExisting)
	core.AssertTrue(t, foundDeploy)
}
