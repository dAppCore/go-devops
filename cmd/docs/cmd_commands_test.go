package docs

import (
	. "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
)

func TestCmdCommands_AddDocsCommands_Good(t *T) {
	root := &cli.Command{Use: "root"}
	AddDocsCommands(root)
	commands := root.Commands()

	AssertLen(t, commands, 1)
	AssertEqual(t, "docs", commands[0].Use)
}

func TestCmdCommands_AddDocsCommands_Bad(t *T) {
	var root *cli.Command
	AssertPanics(t, func() {
		AddDocsCommands(root)
	})
	AssertNil(t, root)
}

func TestCmdCommands_AddDocsCommands_Ugly(t *T) {
	root := &cli.Command{Use: "root"}
	root.AddCommand(&cli.Command{Use: "existing"})
	AddDocsCommands(root)

	foundExisting := false
	foundDocs := false
	for _, cmd := range root.Commands() {
		foundExisting = foundExisting || cmd.Use == "existing"
		foundDocs = foundDocs || cmd.Use == "docs"
	}
	AssertTrue(t, foundExisting)
	AssertTrue(t, foundDocs)
}
