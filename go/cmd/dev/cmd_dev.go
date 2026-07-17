// Package dev provides multi-repo development workflow commands.
//
// Git Operations:
//   - work: Combined status, commit, and push workflow
//   - health: Quick health check across all repos
//   - commit: Claude-assisted commit message generation
//   - push: Push repos with unpushed commits
//   - pull: Pull repos that are behind remote
//
// Forge Integration (uses Forgejo/Gitea API):
//   - issues: List open issues across repos
//   - reviews: List PRs needing review
//   - ci: Check CI workflow status
//   - impact: Analyse dependency impact of changes
//
// CI/Workflow Management:
//   - workflow list: Show table of repos vs workflows
//   - workflow sync: Copy workflow template to all repos
//
// API Tools:
//   - api sync: Synchronize public service APIs
//   - api test-gen: Generate compile-time API test stubs
//
// Dev Environment (VM management):
//   - install: Download dev environment image
//   - boot: Start dev environment VM
//   - stop: Stop dev environment VM
//   - status: Check dev VM status
//   - shell: Open shell in dev VM
//   - serve: Mount project and start dev server
//   - test: Run tests in dev environment
//   - claude: Start sandboxed Claude session
//   - update: Check for and apply updates
package dev

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
	"dappco.re/go/i18n"

	_ "dappco.re/go/devops/locales"
)

// Style aliases from shared package
var (
	successStyle  = cli.SuccessStyle
	errorStyle    = cli.ErrorStyle
	warningStyle  = cli.WarningStyle
	dimStyle      = cli.DimStyle
	valueStyle    = cli.ValueStyle
	headerStyle   = cli.HeaderStyle
	repoNameStyle = cli.RepoStyle
)

// Table styles for status display (extends shared styles with cell padding)
var (
	dirtyStyle = cli.NewStyle().Foreground(cli.ColourRed500)
	aheadStyle = cli.NewStyle().Foreground(cli.ColourAmber500)
	cleanStyle = cli.NewStyle().Foreground(cli.ColourGreen500)
)

// AddDevCommands registers the 'dev' command and all subcommands.
//
//	c := core.New()
//	if r := dev.AddDevCommands(c); !r.OK { return r }
func AddDevCommands(c *core.Core) core.Result {
	// See cmd/deploy/cmd_commands.go's AddDeployCommands for why the group
	// command needs its own Action (core.Cli.Run has no automatic group-help).
	if r := c.Command("dev", core.Command{
		Description: i18n.T("cmd.dev.short"),
		Action: func(core.Options) core.Result {
			cli.PrintHelp()
			return core.Ok(nil)
		},
	}); !r.OK {
		return r
	}

	// Git operations (also available under 'core git')
	for _, register := range []func(*core.Core, string) core.Result{
		AddWorkCommand,
		AddHealthCommand,
		AddCommitCommand,
		AddPushCommand,
		AddPullCommand,
		AddFileSyncCommand,
		AddApplyCommand,
	} {
		if r := register(c, "dev"); !r.OK {
			return r
		}
	}

	if r := AddTagCommand(c); !r.OK {
		return r
	}

	// GitHub integration, CI/workflow management, API tools, dev environment
	for _, register := range []func(*core.Core) core.Result{
		addIssuesCommand,
		addReviewsCommand,
		addCICommand,
		addImpactCommand,
		addWorkflowCommands,
		addAPICommands,
		addVMCommands,
	} {
		if r := register(c); !r.OK {
			return r
		}
	}

	return core.Ok(nil)
}
