package docs

import core "dappco.re/go"

func TestCmdCommands_AddDocsCommands_Good(t *core.T) {
	c := core.New()
	r := AddDocsCommands(c)
	core.AssertTrue(t, r.OK)

	root := c.Command("docs")
	core.AssertTrue(t, root.OK)

	list := c.Command("docs/list")
	core.AssertTrue(t, list.OK)
	core.AssertNotNil(t, list.Value.(*core.Command).Action)
}

func TestCmdCommands_AddDocsCommands_Bad(t *core.T) {
	var c *core.Core
	core.AssertPanics(t, func() {
		AddDocsCommands(c)
	})
	core.AssertNil(t, c)
}

func TestCmdCommands_AddDocsCommands_Ugly(t *core.T) {
	c := core.New()
	core.AssertTrue(t, AddDocsCommands(c).OK)

	// Re-registering onto the same Core hits the duplicate-executable guard.
	r := AddDocsCommands(c)
	core.AssertFalse(t, r.OK)
}
