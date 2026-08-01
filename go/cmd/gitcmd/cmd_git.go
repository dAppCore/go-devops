// Package gitcmd provides git workflow commands as a root-level command.
//
// Git Operations:
//   - health: Show status across repos
//   - commit: Claude-assisted commit message generation
//   - push: Push repos with unpushed commits
//   - pull: Pull repos that are behind remote
//   - work: Combined status, commit, and push workflow
//
// Safe Operations (for AI agents):
//   - file-sync: Sync files across repos with auto commit/push
//   - apply: Run command across repos with auto commit/push
package gitcmd

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
	"dappco.re/go/devops/cmd/dev"
	"dappco.re/go/i18n"
)

// AddGitCommands registers the 'git' command and all subcommands — the same
// dev-package commands mounted under "dev", remounted under "git".
//
//	c := core.New()
//	if r := gitcmd.AddGitCommands(c); !r.OK { return r }
func AddGitCommands(c *core.Core) core.Result {
	// See cmd/deploy/cmd_commands.go's AddDeployCommands for why the group
	// command needs its own Action (core.Cli.Run has no automatic group-help).
	if r := c.Command("git", core.Command{
		Description: i18n.T("cmd.git.short"),
		Action: func(core.Options) core.Result {
			cli.PrintHelp()
			return core.Ok(nil)
		},
	}); !r.OK {
		return r
	}

	for _, register := range []func(*core.Core, string) core.Result{
		// Shows repo status
		dev.AddHealthCommand,
		dev.AddCommitCommand,
		dev.AddPushCommand,
		dev.AddPullCommand,
		dev.AddWorkCommand,
		// Safe operations for AI agents
		dev.AddFileSyncCommand,
		dev.AddApplyCommand,
	} {
		if r := register(c, "git"); !r.OK {
			return r
		}
	}

	return core.Ok(nil)
}
