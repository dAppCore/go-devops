package setup

import core "dappco.re/go"

func ExampleAddSetupCommand() {
	c := core.New()
	AddSetupCommand(c)
	r := c.Command("setup")
	core.Println(r.OK)
	// Output: true
}
