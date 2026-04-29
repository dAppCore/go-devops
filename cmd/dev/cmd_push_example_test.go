package dev

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
)

func ExampleAddPushCommand() {
	root := &cli.Command{Use: "root"}
	AddPushCommand(root)
	core.Println(root.Commands()[0].Use)
	// Output: push
}
