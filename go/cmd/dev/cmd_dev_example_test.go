package dev

import core "dappco.re/go"

func ExampleAddDevCommands() {
	c := core.New()
	AddDevCommands(c)
	r := c.Command("dev")
	core.Println(r.OK)
	// Output: true
}
