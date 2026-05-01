package dev

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
)

func ExampleAddHealthCommand() {
	root := &cli.Command{Use: "root"}
	AddHealthCommand(root)
	core.Println(root.Commands()[0].Use)
	// Output: health
}
