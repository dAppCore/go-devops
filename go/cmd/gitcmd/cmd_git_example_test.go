package gitcmd

import core "dappco.re/go"

func ExampleAddGitCommands() {
	c := core.New()
	AddGitCommands(c)
	r := c.Command("git")
	core.Println(r.OK)
	// Output: true
}
