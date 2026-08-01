package dev

import core "dappco.re/go"

func TestCmdPull_AddPullCommand_Good(t *core.T) {
	c := core.New()
	r := AddPullCommand(c, "dev")
	core.AssertTrue(t, r.OK)

	cmd := c.Command("dev/pull")
	core.AssertTrue(t, cmd.OK)
	core.AssertNotNil(t, cmd.Value.(*core.Command).Action)
}

func TestCmdPull_AddPullCommand_Bad(t *core.T) {
	var c *core.Core
	core.AssertPanics(t, func() {
		AddPullCommand(c, "dev")
	})
	core.AssertNil(t, c)
}

func TestCmdPull_AddPullCommand_Ugly(t *core.T) {
	c := core.New()
	core.AssertTrue(t, AddPullCommand(c, "dev").OK)

	r := AddPullCommand(c, "git")
	core.AssertTrue(t, r.OK)
	core.AssertTrue(t, c.Command("git/pull").OK)
}
