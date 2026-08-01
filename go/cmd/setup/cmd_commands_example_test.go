package setup

import core "dappco.re/go"

func ExampleAddSetupCommands() {
	c := core.New()
	AddSetupCommands(c)
	r := c.Command("setup")
	core.Println(r.OK)
	// Output: true
}
