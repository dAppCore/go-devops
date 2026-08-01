package dev

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
	"dappco.re/go/i18n"
)

// addAPICommands adds the 'api' group and its 'sync'/'test-gen' subcommands.
func addAPICommands(c *core.Core) core.Result {
	// See cmd/deploy/cmd_commands.go's AddDeployCommands for why the group
	// command needs its own Action (core.Cli.Run has no automatic group-help).
	if r := c.Command("dev/api", core.Command{
		Description: i18n.T("cmd.dev.api.short"),
		Action: func(core.Options) core.Result {
			cli.PrintHelp()
			return core.Ok(nil)
		},
	}); !r.OK {
		return r
	}
	if r := addSyncCommand(c); !r.OK {
		return r
	}
	return addTestGenCommand(c)
}
