package dev

import core "dappco.re/go"

func ExampleAddTagCommand() {
	c := core.New()
	AddTagCommand(c)
	r := c.Command("dev/tag")
	core.Println(r.OK)
	// Output: true
}
