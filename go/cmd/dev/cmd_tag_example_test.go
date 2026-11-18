package dev

import core "github.com/dappcore/go"

func ExampleAddTagCommand() {
	c := core.New()
	AddTagCommand(c)
	r := c.Command("dev/tag")
	core.Println(r.OK)
	// Output: true
}
