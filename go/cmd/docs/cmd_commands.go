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
	"github.com/dappcore/cli/go/pkg/cli"
	core "github.com/dappcore/go"
	"github.com/dappcore/go-i18n/go"

	_ "github.com/dappcore/go-devops/go/locales"
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
