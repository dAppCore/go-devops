package dev

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
)

func ExampleAddPullCommand() {
	root := &cli.Command{Use: "root"}
	AddPullCommand(root)
	core.Println(root.Commands()[0].Use)
	// Output: pull
}
