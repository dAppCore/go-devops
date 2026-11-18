package dev

import core "github.com/dappcore/go"

func ExampleAddHealthCommand() {
	c := core.New()
	AddHealthCommand(c, "dev")
	r := c.Command("dev/health")
	core.Println(r.OK)
	// Output: true
}
