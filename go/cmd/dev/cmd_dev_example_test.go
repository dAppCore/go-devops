package dev

import core "github.com/dappcore/go"

func ExampleAddDevCommands() {
	c := core.New()
	AddDevCommands(c)
	r := c.Command("dev")
	core.Println(r.OK)
	// Output: true
}
