// Package docs provides documentation management commands for multi-repo workspaces.
//
// Commands:
//   - list: Scan repos for README.md, CLAUDE.md, CHANGELOG.md, docs/
//   - sync: Copy docs/ files from all repos to core-php/docs/packages/
//
// Works with repos.yaml to discover repositories and sync documentation
// to a central location for unified documentation builds.
package docs

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
	"dappco.re/go/i18n"

	_ "dappco.re/go/devops/locales"
)

// AddDocsCommands registers the 'docs' command and all subcommands.
//
//	c := core.New()
//	if r := docs.AddDocsCommands(c); !r.OK { return r }
func AddDocsCommands(c *core.Core) core.Result {
	// See cmd/deploy/cmd_commands.go's AddDeployCommands for why the group
	// command needs its own Action (core.Cli.Run has no automatic group-help).
	if r := c.Command("docs", core.Command{
		Description: i18n.T("cmd.docs.short"),
		Action: func(core.Options) core.Result {
			cli.PrintHelp()
			return core.Ok(nil)
		},
	}); !r.OK {
		return r
	}
	if r := addDocsListCommand(c); !r.OK {
		return r
	}
	return addDocsSyncCommand(c)
}
