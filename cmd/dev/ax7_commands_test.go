package dev

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
)

func ax7Command(root *cli.Command, use string) *cli.Command {
	for _, cmd := range root.Commands() {
		if cmd.Use == use || core.HasPrefix(cmd.Use, use+" ") {
			return cmd
		}
	}
	return nil
}

func TestAX7_AddDevCommands_Good(t *core.T) {
	root := &cli.Command{Use: "root"}
	AddDevCommands(root)
	devCmd := ax7Command(root, "dev")

	core.AssertNotNil(t, devCmd)
	core.AssertGreaterOrEqual(t, len(devCmd.Commands()), 10)
}

func TestAX7_AddDevCommands_Bad(t *core.T) {
	var root *cli.Command
	core.AssertPanics(t, func() {
		AddDevCommands(root)
	})
	core.AssertNil(t, root)
}

func TestAX7_AddDevCommands_Ugly(t *core.T) {
	root := &cli.Command{Use: "root"}
	root.AddCommand(&cli.Command{Use: "existing"})
	AddDevCommands(root)

	core.AssertLen(t, root.Commands(), 2)
	core.AssertNotNil(t, ax7Command(root, "dev"))
}

func TestAX7_AddApplyCommand_Good(t *core.T) {
	root := &cli.Command{Use: "root"}
	AddApplyCommand(root)
	cmd := ax7Command(root, "apply")

	core.AssertNotNil(t, cmd)
	core.AssertNotNil(t, cmd.Flag("command"))
}

func TestAX7_AddApplyCommand_Bad(t *core.T) {
	var root *cli.Command
	core.AssertPanics(t, func() {
		AddApplyCommand(root)
	})
	core.AssertNil(t, root)
}

func TestAX7_AddApplyCommand_Ugly(t *core.T) {
	root := &cli.Command{Use: "root"}
	AddApplyCommand(root)
	AddApplyCommand(root)

	core.AssertLen(t, root.Commands(), 2)
	core.AssertNotNil(t, ax7Command(root, "apply"))
}

func TestAX7_AddFileSyncCommand_Good(t *core.T) {
	root := &cli.Command{Use: "root"}
	AddFileSyncCommand(root)
	cmd := ax7Command(root, "sync")

	core.AssertNotNil(t, cmd)
	core.AssertNotNil(t, cmd.Flag("to"))
}

func TestAX7_AddFileSyncCommand_Bad(t *core.T) {
	var root *cli.Command
	core.AssertPanics(t, func() {
		AddFileSyncCommand(root)
	})
	core.AssertNil(t, root)
}

func TestAX7_AddFileSyncCommand_Ugly(t *core.T) {
	root := &cli.Command{Use: "root"}
	AddFileSyncCommand(root)
	AddFileSyncCommand(root)

	core.AssertLen(t, root.Commands(), 2)
	core.AssertNotNil(t, ax7Command(root, "sync"))
}

func TestAX7_AddPushCommand_Good(t *core.T) {
	root := &cli.Command{Use: "root"}
	AddPushCommand(root)
	cmd := ax7Command(root, "push")

	core.AssertNotNil(t, cmd)
	core.AssertNotNil(t, cmd.Flag("force"))
}

func TestAX7_AddPushCommand_Bad(t *core.T) {
	var root *cli.Command
	core.AssertPanics(t, func() {
		AddPushCommand(root)
	})
	core.AssertNil(t, root)
}

func TestAX7_AddPushCommand_Ugly(t *core.T) {
	root := &cli.Command{Use: "root"}
	root.AddCommand(&cli.Command{Use: "existing"})
	AddPushCommand(root)

	core.AssertLen(t, root.Commands(), 2)
	core.AssertNotNil(t, ax7Command(root, "push"))
}

func TestAX7_AddWorkCommand_Good(t *core.T) {
	root := &cli.Command{Use: "root"}
	AddWorkCommand(root)
	cmd := ax7Command(root, "work")

	core.AssertNotNil(t, cmd)
	core.AssertNotNil(t, cmd.Flag("status"))
}

func TestAX7_AddWorkCommand_Bad(t *core.T) {
	var root *cli.Command
	core.AssertPanics(t, func() {
		AddWorkCommand(root)
	})
	core.AssertNil(t, root)
}

func TestAX7_AddWorkCommand_Ugly(t *core.T) {
	root := &cli.Command{Use: "root"}
	AddWorkCommand(root)
	AddWorkCommand(root)

	core.AssertLen(t, root.Commands(), 2)
	core.AssertNotNil(t, ax7Command(root, "work"))
}

func TestAX7_AddCommitCommand_Good(t *core.T) {
	root := &cli.Command{Use: "root"}
	AddCommitCommand(root)
	cmd := ax7Command(root, "commit")

	core.AssertNotNil(t, cmd)
	core.AssertNotNil(t, cmd.Flag("all"))
}

func TestAX7_AddCommitCommand_Bad(t *core.T) {
	var root *cli.Command
	core.AssertPanics(t, func() {
		AddCommitCommand(root)
	})
	core.AssertNil(t, root)
}

func TestAX7_AddCommitCommand_Ugly(t *core.T) {
	root := &cli.Command{Use: "root"}
	root.AddCommand(&cli.Command{Use: "existing"})
	AddCommitCommand(root)

	core.AssertLen(t, root.Commands(), 2)
	core.AssertNotNil(t, ax7Command(root, "commit"))
}

func TestAX7_AddHealthCommand_Good(t *core.T) {
	root := &cli.Command{Use: "root"}
	AddHealthCommand(root)
	cmd := ax7Command(root, "health")

	core.AssertNotNil(t, cmd)
	core.AssertNotNil(t, cmd.Flag("verbose"))
}

func TestAX7_AddHealthCommand_Bad(t *core.T) {
	var root *cli.Command
	core.AssertPanics(t, func() {
		AddHealthCommand(root)
	})
	core.AssertNil(t, root)
}

func TestAX7_AddHealthCommand_Ugly(t *core.T) {
	root := &cli.Command{Use: "root"}
	AddHealthCommand(root)
	AddHealthCommand(root)

	core.AssertLen(t, root.Commands(), 2)
	core.AssertNotNil(t, ax7Command(root, "health"))
}

func TestAX7_AddTagCommand_Good(t *core.T) {
	root := &cli.Command{Use: "root"}
	AddTagCommand(root)
	cmd := ax7Command(root, "tag")

	core.AssertNotNil(t, cmd)
	core.AssertNotNil(t, cmd.Flag("dry-run"))
}

func TestAX7_AddTagCommand_Bad(t *core.T) {
	var root *cli.Command
	core.AssertPanics(t, func() {
		AddTagCommand(root)
	})
	core.AssertNil(t, root)
}

func TestAX7_AddTagCommand_Ugly(t *core.T) {
	root := &cli.Command{Use: "root"}
	root.AddCommand(&cli.Command{Use: "existing"})
	AddTagCommand(root)

	core.AssertLen(t, root.Commands(), 2)
	core.AssertNotNil(t, ax7Command(root, "tag"))
}

func TestAX7_AddPullCommand_Good(t *core.T) {
	root := &cli.Command{Use: "root"}
	AddPullCommand(root)
	cmd := ax7Command(root, "pull")

	core.AssertNotNil(t, cmd)
	core.AssertNotNil(t, cmd.Flag("all"))
}

func TestAX7_AddPullCommand_Bad(t *core.T) {
	var root *cli.Command
	core.AssertPanics(t, func() {
		AddPullCommand(root)
	})
	core.AssertNil(t, root)
}

func TestAX7_AddPullCommand_Ugly(t *core.T) {
	root := &cli.Command{Use: "root"}
	AddPullCommand(root)
	AddPullCommand(root)

	core.AssertLen(t, root.Commands(), 2)
	core.AssertNotNil(t, ax7Command(root, "pull"))
}
