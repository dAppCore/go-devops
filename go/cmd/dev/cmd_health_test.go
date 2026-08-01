package dev

import core "dappco.re/go"

func TestCmdHealth_AddHealthCommand_Good(t *core.T) {
	c := core.New()
	r := AddHealthCommand(c, "dev")
	core.AssertTrue(t, r.OK)

	cmd := c.Command("dev/health")
	core.AssertTrue(t, cmd.OK)
	core.AssertNotNil(t, cmd.Value.(*core.Command).Action)
}

func TestCmdHealth_AddHealthCommand_Bad(t *core.T) {
	var c *core.Core
	core.AssertPanics(t, func() {
		AddHealthCommand(c, "dev")
	})
	core.AssertNil(t, c)
}

func TestCmdHealth_AddHealthCommand_Ugly(t *core.T) {
	c := core.New()
	core.AssertTrue(t, AddHealthCommand(c, "dev").OK)

	// Same command, different prefix (mirrors AddGitCommands mounting it
	// under "git" too) — registers cleanly at the new path.
	r := AddHealthCommand(c, "git")
	core.AssertTrue(t, r.OK)
	core.AssertTrue(t, c.Command("git/health").OK)
}
