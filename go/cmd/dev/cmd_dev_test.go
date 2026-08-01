package dev

import core "dappco.re/go"

func TestCmdDev_AddDevCommands_Good(t *core.T) {
	c := core.New()
	r := AddDevCommands(c)
	core.AssertTrue(t, r.OK)

	devCmd := c.Command("dev")
	core.AssertTrue(t, devCmd.OK)
	core.AssertGreaterOrEqual(t, len(c.Commands()), 10)
}

func TestCmdDev_AddDevCommands_Bad(t *core.T) {
	var c *core.Core
	core.AssertPanics(t, func() {
		AddDevCommands(c)
	})
	core.AssertNil(t, c)
}

func TestCmdDev_AddDevCommands_Ugly(t *core.T) {
	c := core.New()
	core.AssertTrue(t, AddDevCommands(c).OK)

	// Re-registering onto the same Core hits the duplicate-executable guard.
	r := AddDevCommands(c)
	core.AssertFalse(t, r.OK)
}
