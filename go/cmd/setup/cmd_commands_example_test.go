package setup

import (
	. "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
)

func ExampleAddSetupCommands() {
	root := &cli.Command{Use: "root"}
	AddSetupCommands(root)
	Println(root.Commands()[0].Use)
	// Output: setup
}
