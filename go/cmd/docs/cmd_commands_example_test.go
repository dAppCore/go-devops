package docs

import (
	. "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
)

func ExampleAddDocsCommands() {
	root := &cli.Command{Use: "root"}
	AddDocsCommands(root)
	Println(root.Commands()[0].Use)
	// Output: docs
}
