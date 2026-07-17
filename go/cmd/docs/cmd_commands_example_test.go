package docs

import core "dappco.re/go"

func ExampleAddDocsCommands() {
	c := core.New()
	AddDocsCommands(c)
	r := c.Command("docs")
	core.Println(r.OK)
	// Output: true
}
