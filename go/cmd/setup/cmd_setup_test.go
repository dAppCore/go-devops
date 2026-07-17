package setup

import core "dappco.re/go"

func TestCmdSetup_AddSetupCommand_Good(t *core.T) {
	c := core.New()
	r := AddSetupCommand(c)
	core.AssertTrue(t, r.OK)

	cmd := c.Command("setup")
	core.AssertTrue(t, cmd.OK)
	setupCmd := cmd.Value.(*core.Command)
	core.AssertNotNil(t, setupCmd.Action)
	core.AssertTrue(t, setupCmd.Flags.Has("registry"))
}

func TestCmdSetup_AddSetupCommand_Bad(t *core.T) {
	var c *core.Core
	core.AssertPanics(t, func() {
		AddSetupCommand(c)
	})
	core.AssertNil(t, c)
}

func TestCmdSetup_AddSetupCommand_Ugly(t *core.T) {
	c := core.New()
	core.AssertTrue(t, AddSetupCommand(c).OK)

	// Re-registering onto the same Core hits the duplicate-executable guard.
	r := AddSetupCommand(c)
	core.AssertFalse(t, r.OK)
}
