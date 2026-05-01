package dev

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
)

func ExampleAddFileSyncCommand() {
	root := &cli.Command{Use: "root"}
	AddFileSyncCommand(root)
	core.Println(root.Commands()[0].Use)
	// Output: sync <file-or-dir>
}
