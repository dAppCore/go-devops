package dev

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
)

func ExampleAddWorkCommand() {
	root := &cli.Command{Use: "root"}
	AddWorkCommand(root)
	core.Println(root.Commands()[0].Use)
	// Output: work
}
