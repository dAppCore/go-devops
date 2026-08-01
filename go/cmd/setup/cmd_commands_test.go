package setup

import core "dappco.re/go"

func TestCmdCommands_AddSetupCommands_Good(t *core.T) {
	c := core.New()
	r := AddSetupCommands(c)
	core.AssertTrue(t, r.OK)

	cmd := c.Command("setup")
	core.AssertTrue(t, cmd.OK)
	core.AssertNotEmpty(t, cmd.Value.(*core.Command).Description)
}

func TestCmdCommands_AddSetupCommands_Bad(t *core.T) {
	var c *core.Core
	core.AssertPanics(t, func() {
		AddSetupCommands(c)
	})
	core.AssertNil(t, c)
}

func TestCmdCommands_AddSetupCommands_Ugly(t *core.T) {
	c := core.New()
	core.AssertTrue(t, AddSetupCommands(c).OK)

	// Re-registering onto the same Core hits the duplicate-executable guard.
	r := AddSetupCommands(c)
	core.AssertFalse(t, r.OK)
}
