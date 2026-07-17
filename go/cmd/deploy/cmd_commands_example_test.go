package deploy

import core "dappco.re/go"

func ExampleAddDeployCommands() {
	c := core.New()
	AddDeployCommands(c)
	r := c.Command("deploy/servers")
	core.Println(r.OK)
	// Output: true
}
