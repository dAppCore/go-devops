package dev

import core "dappco.re/go"

func ExampleAddCommitCommand() {
	c := core.New()
	AddCommitCommand(c, "dev")
	r := c.Command("dev/commit")
	core.Println(r.OK)
	// Output: true
}
