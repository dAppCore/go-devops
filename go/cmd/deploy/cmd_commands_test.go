package deploy

import core "dappco.re/go"

func TestCmdCommands_AddDeployCommands_Good(t *core.T) {
	c := core.New()
	r := AddDeployCommands(c)
	core.AssertTrue(t, r.OK)

	root := c.Command("deploy")
	core.AssertTrue(t, root.OK)

	servers := c.Command("deploy/servers")
	core.AssertTrue(t, servers.OK)
	core.AssertNotNil(t, servers.Value.(*core.Command).Action)
}

func TestCmdCommands_AddDeployCommands_Bad(t *core.T) {
	var c *core.Core
	core.AssertPanics(t, func() {
		AddDeployCommands(c)
	})
	core.AssertNil(t, c)
}

func TestCmdCommands_AddDeployCommands_Ugly(t *core.T) {
	c := core.New()
	core.AssertTrue(t, AddDeployCommands(c).OK)

	// Re-registering onto the same Core hits the duplicate-executable guard.
	r := AddDeployCommands(c)
	core.AssertFalse(t, r.OK)
}
