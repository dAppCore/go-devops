// Package setup provides workspace setup and bootstrap commands.
package setup

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
	"dappco.re/go/i18n"
)

// Style aliases from shared package
var (
	repoNameStyle = cli.RepoStyle
	successStyle  = cli.SuccessStyle
	errorStyle    = cli.ErrorStyle
	warningStyle  = cli.WarningStyle
	dimStyle      = cli.DimStyle
)

// Default organization and devops repo for bootstrap
const (
	defaultOrg      = "host-uk"
	devopsRepo      = "core-devops"
	devopsReposYaml = "repos.yaml"
)

// AddSetupCommand adds the 'setup' command (itself executable — the
// registry/bootstrap orchestrator) plus its 'repo', 'github'/'gh', and 'ci'
// subcommands.
//
//	c := core.New()
//	if r := setup.AddSetupCommand(c); !r.OK { return r }
func AddSetupCommand(c *core.Core) core.Result {
	if r := c.Command("setup", core.Command{
		Description: i18n.T("cmd.setup.short"),
		Flags: core.NewOptions(
			core.Option{Key: "registry", Value: ""},
			core.Option{Key: "only", Value: ""},
			core.Option{Key: "dry-run", Value: false},
			core.Option{Key: "all", Value: false},
			core.Option{Key: "name", Value: ""},
			core.Option{Key: "build", Value: false},
		),
		Action: func(o core.Options) core.Result {
			return runSetupOrchestrator(o.String("registry"), o.String("only"), o.Bool("dry-run"), o.Bool("all"), o.String("name"), o.Bool("build"))
		},
	}); !r.OK {
		return r
	}
	if r := addRepoCommand(c); !r.OK {
		return r
	}
	if r := addGitHubCommand(c); !r.OK {
		return r
	}
	return addSetupCICommand(c)
}
