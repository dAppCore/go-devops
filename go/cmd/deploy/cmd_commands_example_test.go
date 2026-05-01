package deploy

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
)

func ExampleAddDeployCommands() {
	root := &cli.Command{Use: "root"}
	AddDeployCommands(root)
	core.Println(root.Commands()[0].Use)
	// Output: deploy
}
