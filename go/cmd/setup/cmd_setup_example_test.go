package setup

import (
	. "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
)

func ExampleAddSetupCommand() {
	original := setupCmd
	setupCmd = &cli.Command{Use: "setup"}
	defer func() { setupCmd = original }()

	root := &cli.Command{Use: "root"}
	AddSetupCommand(root)
	Println(root.Commands()[0].Use)
	// Output: setup
}
