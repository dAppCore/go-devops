package dev

import (
	"dappco.re/go/i18n"
	"dappco.re/go/cli/pkg/cli"
)

// addAPICommands adds the 'api' command and its subcommands to the given parent command.
func addAPICommands(parent *cli.Command) {
	// Create the 'api' command
	apiCmd := &cli.Command{
		Use:   "api",
		Short: i18n.T("cmd.dev.api.short"),
	}
	parent.AddCommand(apiCmd)

	// Add the 'sync' command to 'api'
	addSyncCommand(apiCmd)

	// Add the 'test-gen' command to 'api'
	addTestGenCommand(apiCmd)
}
