package dev

import core "github.com/dappcore/go"

func ExampleAddPullCommand() {
	c := core.New()
	AddPullCommand(c, "dev")
	r := c.Command("dev/pull")
	core.Println(r.OK)
	// Output: true
}
