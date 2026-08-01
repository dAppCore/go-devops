package dev

import core "dappco.re/go"

func TestCmdPush_AddPushCommand_Good(t *core.T) {
	c := core.New()
	r := AddPushCommand(c, "dev")
	core.AssertTrue(t, r.OK)

	cmd := c.Command("dev/push")
	core.AssertTrue(t, cmd.OK)
	core.AssertNotNil(t, cmd.Value.(*core.Command).Action)
}

func TestCmdPush_AddPushCommand_Bad(t *core.T) {
	var c *core.Core
	core.AssertPanics(t, func() {
		AddPushCommand(c, "dev")
	})
	core.AssertNil(t, c)
}

func TestCmdPush_AddPushCommand_Ugly(t *core.T) {
	c := core.New()
	core.AssertTrue(t, AddPushCommand(c, "dev").OK)

	r := AddPushCommand(c, "git")
	core.AssertTrue(t, r.OK)
	core.AssertTrue(t, c.Command("git/push").OK)
}
