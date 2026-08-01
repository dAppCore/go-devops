package dev

import core "dappco.re/go"

func TestCmdTag_AddTagCommand_Good(t *core.T) {
	c := core.New()
	r := AddTagCommand(c)
	core.AssertTrue(t, r.OK)

	cmd := c.Command("dev/tag")
	core.AssertTrue(t, cmd.OK)
	core.AssertNotNil(t, cmd.Value.(*core.Command).Action)
}

func TestCmdTag_AddTagCommand_Bad(t *core.T) {
	var c *core.Core
	core.AssertPanics(t, func() {
		AddTagCommand(c)
	})
	core.AssertNil(t, c)
}

func TestCmdTag_AddTagCommand_Ugly(t *core.T) {
	c := core.New()
	core.AssertTrue(t, AddTagCommand(c).OK)

	// Re-registering onto the same Core hits the duplicate-executable guard.
	r := AddTagCommand(c)
	core.AssertFalse(t, r.OK)
}
