package dev

import core "dappco.re/go"

func ExampleAddHealthCommand() {
	c := core.New()
	AddHealthCommand(c, "dev")
	r := c.Command("dev/health")
	core.Println(r.OK)
	// Output: true
}
