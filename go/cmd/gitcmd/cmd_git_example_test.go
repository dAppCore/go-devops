package gitcmd

import (
	. "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
)

func ExampleAddGitCommands() {
	root := &cli.Command{Use: "root"}
	AddGitCommands(root)
	Println(root.Commands()[0].Use)
	// Output: git
}
