package dev

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
)

func ExampleAddDevCommands() {
	root := &cli.Command{Use: "root"}
	AddDevCommands(root)
	core.Println(root.Commands()[0].Use)
	// Output: dev
}
