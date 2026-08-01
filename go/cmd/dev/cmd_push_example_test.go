package dev

import core "dappco.re/go"

func ExampleAddPushCommand() {
	c := core.New()
	AddPushCommand(c, "dev")
	r := c.Command("dev/push")
	core.Println(r.OK)
	// Output: true
}
