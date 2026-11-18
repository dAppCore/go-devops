package dev

import core "github.com/dappcore/go"

func ExampleAddApplyCommand() {
	c := core.New()
	AddApplyCommand(c, "dev")
	r := c.Command("dev/apply")
	core.Println(r.OK)
	// Output: true
}
