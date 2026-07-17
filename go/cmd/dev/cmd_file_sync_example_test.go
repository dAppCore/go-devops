package dev

import core "dappco.re/go"

func ExampleAddFileSyncCommand() {
	c := core.New()
	AddFileSyncCommand(c, "dev")
	r := c.Command("dev/sync")
	core.Println(r.OK)
	// Output: true
}
