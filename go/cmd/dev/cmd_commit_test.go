package dev

import core "dappco.re/go"

func TestCmdCommit_AddCommitCommand_Good(t *core.T) {
	c := core.New()
	r := AddCommitCommand(c, "dev")
	core.AssertTrue(t, r.OK)

	cmd := c.Command("dev/commit")
	core.AssertTrue(t, cmd.OK)
	core.AssertNotNil(t, cmd.Value.(*core.Command).Action)
}

func TestCmdCommit_AddCommitCommand_Bad(t *core.T) {
	var c *core.Core
	core.AssertPanics(t, func() {
		AddCommitCommand(c, "dev")
	})
	core.AssertNil(t, c)
}

func TestCmdCommit_AddCommitCommand_Ugly(t *core.T) {
	c := core.New()
	core.AssertTrue(t, AddCommitCommand(c, "dev").OK)

	r := AddCommitCommand(c, "git")
	core.AssertTrue(t, r.OK)
	core.AssertTrue(t, c.Command("git/commit").OK)
}
