package dev

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
)

func ExampleAddTagCommand() {
	root := &cli.Command{Use: "root"}
	AddTagCommand(root)
	core.Println(root.Commands()[0].Use)
	// Output: tag
}
