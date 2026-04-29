package dev

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
)

func ExampleAddApplyCommand() {
	root := &cli.Command{Use: "root"}
	AddApplyCommand(root)
	core.Println(root.Commands()[0].Use)
	// Output: apply
}
