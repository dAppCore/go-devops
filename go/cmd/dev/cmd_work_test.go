package dev

import core "dappco.re/go"

func TestCmdWork_AddWorkCommand_Good(t *core.T) {
	c := core.New()
	r := AddWorkCommand(c, "dev")
	core.AssertTrue(t, r.OK)

	cmd := c.Command("dev/work")
	core.AssertTrue(t, cmd.OK)
	core.AssertNotNil(t, cmd.Value.(*core.Command).Action)
}

func TestCmdWork_AddWorkCommand_Bad(t *core.T) {
	var c *core.Core
	core.AssertPanics(t, func() {
		AddWorkCommand(c, "dev")
	})
	core.AssertNil(t, c)
}

func TestCmdWork_AddWorkCommand_Ugly(t *core.T) {
	c := core.New()
	core.AssertTrue(t, AddWorkCommand(c, "dev").OK)

	r := AddWorkCommand(c, "git")
	core.AssertTrue(t, r.OK)
	core.AssertTrue(t, c.Command("git/work").OK)
}
