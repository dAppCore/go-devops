package dev

import core "github.com/dappcore/go"

func ExampleAddWorkCommand() {
	c := core.New()
	AddWorkCommand(c, "dev")
	r := c.Command("dev/work")
	core.Println(r.OK)
	// Output: true
}
