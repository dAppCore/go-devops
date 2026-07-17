// Package deploy implements the `core deploy` command tree — Coolify PaaS
// server/app/database/service listing and raw API calls.
package deploy

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
	"dappco.re/go/i18n"

	_ "dappco.re/go/devops/locales"
)

// commonFlags are declared on every deploy subcommand: Coolify connection
// (--url/--token, falling back to COOLIFY_URL/COOLIFY_TOKEN) plus --json output.
func commonFlags() core.Options {
	return core.NewOptions(
		core.Option{Key: "url", Value: ""},
		core.Option{Key: "token", Value: ""},
		core.Option{Key: "json", Value: false},
	)
}

// AddDeployCommands registers the 'deploy' command and all subcommands.
//
//	c := core.New()
//	if r := deploy.AddDeployCommands(c); !r.OK { return r }
func AddDeployCommands(c *core.Core) core.Result {
	// A bare "deploy" (or "deploy --help") has no leaf Action of its own —
	// core.Cli.Run treats an Action-less command as non-executable and would
	// error rather than showing usage (unlike cobra's automatic group-help
	// behaviour), so this prints the catalog instead of leaving Action nil.
	if r := c.Command("deploy", core.Command{
		Description: i18n.T("cmd.deploy.short"),
		Action: func(core.Options) core.Result {
			cli.PrintHelp()
			return core.Ok(nil)
		},
	}); !r.OK {
		return r
	}
	if r := c.Command("deploy/servers", core.Command{
		Description: "List Coolify servers",
		Flags:       commonFlags(),
		Action:      func(o core.Options) core.Result { return runListServers(o) },
	}); !r.OK {
		return r
	}
	if r := c.Command("deploy/projects", core.Command{
		Description: "List Coolify projects",
		Flags:       commonFlags(),
		Action:      func(o core.Options) core.Result { return runListProjects(o) },
	}); !r.OK {
		return r
	}
	if r := c.Command("deploy/apps", core.Command{
		Description: "List Coolify applications",
		Flags:       commonFlags(),
		Action:      func(o core.Options) core.Result { return runListApps(o) },
	}); !r.OK {
		return r
	}
	if r := c.Command("deploy/databases", core.Command{
		Description: "List Coolify databases",
		Flags:       commonFlags(),
		Action:      func(o core.Options) core.Result { return runListDatabases(o) },
	}); !r.OK {
		return r
	}
	if r := c.Command("deploy/services", core.Command{
		Description: "List Coolify services",
		Flags:       commonFlags(),
		Action:      func(o core.Options) core.Result { return runListServices(o) },
	}); !r.OK {
		return r
	}
	if r := c.Command("deploy/team", core.Command{
		Description: "Show current team info",
		Flags:       commonFlags(),
		Action:      func(o core.Options) core.Result { return runTeam(o) },
	}); !r.OK {
		return r
	}
	// call <operation> [--params=<json>]: the params-json positional from the
	// old `call <operation> [params-json]` cobra signature became a flag —
	// core.Cli.Run keeps only the last bare positional under "_arg", so a
	// second positional would silently overwrite the operation name.
	if r := c.Command("deploy/call", core.Command{
		Description: "Call any Coolify API operation",
		Flags: core.NewOptions(
			core.Option{Key: "url", Value: ""},
			core.Option{Key: "token", Value: ""},
			core.Option{Key: "json", Value: false},
			core.Option{Key: "params", Value: ""},
		),
		Action: func(o core.Options) core.Result { return runCall(o) },
	}); !r.OK {
		return r
	}
	return core.Ok(nil)
}
