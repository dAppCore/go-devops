package dev

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
)

func ExampleAddCommitCommand() {
	root := &cli.Command{Use: "root"}
	AddCommitCommand(root)
	core.Println(root.Commands()[0].Use)
	// Output: commit
}
