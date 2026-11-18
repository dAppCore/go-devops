package docs

import core "github.com/dappcore/go"

func ExampleAddDocsCommands() {
	c := core.New()
	AddDocsCommands(c)
	r := c.Command("docs")
	core.Println(r.OK)
	// Output: true
}
