package setup

import core "github.com/dappcore/go"

func ExampleAddSetupCommand() {
	c := core.New()
	AddSetupCommand(c)
	r := c.Command("setup")
	core.Println(r.OK)
	// Output: true
}
