package setup

import core "github.com/dappcore/go"

func ExampleAddSetupCommands() {
	c := core.New()
	AddSetupCommands(c)
	r := c.Command("setup")
	core.Println(r.OK)
	// Output: true
}
