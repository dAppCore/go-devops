// cmd_apply.go implements safe command/script execution across repos for AI agents.
//
// Usage:
//   core dev apply --command="sed -i 's/old/new/g' README.md"
//   core dev apply --script="./scripts/update-version.sh"
//   core dev apply --command="..." --commit --message="chore: update"

package dev

import (
	"context"
	"sort"

	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
	"dappco.re/go/i18n"
	"dappco.re/go/io"
	log "dappco.re/go/log"
	coreexec "dappco.re/go/process/exec"
	"dappco.re/go/scm/git"
	"dappco.re/go/scm/repos"
)

// Apply command flags
var (
	applyCommand  string
	applyScript   string
	applyRepos    string
	applyCommit   bool
	applyMessage  string
	applyCoAuthor string
	applyDryRun   bool
	applyPush     bool
	applyContinue bool // Continue on error
	applyYes      bool // Skip confirmation prompt
)

// AddApplyCommand adds the 'apply' command to dev.
func AddApplyCommand(parent *cli.Command) {
	applyCmd := &cli.Command{
		Use:   "apply",
		Short: i18n.T("cmd.dev.apply.short"),
		Long:  i18n.T("cmd.dev.apply.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			return runApply()
		},
	}

	applyCmd.Flags().StringVar(&applyCommand, "command", "", i18n.T("cmd.dev.apply.flag.command"))
	applyCmd.Flags().StringVar(&applyScript, "script", "", i18n.T("cmd.dev.apply.flag.script"))
	applyCmd.Flags().StringVar(&applyRepos, "repos", "", i18n.T("cmd.dev.apply.flag.repos"))
	applyCmd.Flags().BoolVar(&applyCommit, "commit", false, i18n.T("cmd.dev.apply.flag.commit"))
	applyCmd.Flags().StringVarP(&applyMessage, "message", "m", "", i18n.T("cmd.dev.apply.flag.message"))
	applyCmd.Flags().StringVar(&applyCoAuthor, "co-author", "", i18n.T("cmd.dev.apply.flag.co_author"))
	applyCmd.Flags().BoolVar(&applyDryRun, "dry-run", false, i18n.T("cmd.dev.apply.flag.dry_run"))
	applyCmd.Flags().BoolVar(&applyPush, "push", false, i18n.T("cmd.dev.apply.flag.push"))
	applyCmd.Flags().BoolVar(&applyContinue, "continue", false, i18n.T("cmd.dev.apply.flag.continue"))
	applyCmd.Flags().BoolVarP(&applyYes, "yes", "y", false, i18n.T("cmd.dev.apply.flag.yes"))

	parent.AddCommand(applyCmd)
}

func runApply() (_ coreFailure) {
	ctx := context.Background()

	// Validate inputs
	if applyCommand == "" && applyScript == "" {
		return log.E("dev.apply", i18n.T("cmd.dev.apply.error.no_command"), nil)
	}
	if applyCommand != "" && applyScript != "" {
		return log.E("dev.apply", i18n.T("cmd.dev.apply.error.both_command_script"), nil)
	}
	if applyCommit && applyMessage == "" {
		return log.E("dev.apply", i18n.T("cmd.dev.apply.error.commit_needs_message"), nil)
	}

	// Validate script exists
	if applyScript != "" {
		if !io.Local.IsFile(applyScript) {
			return core.E("dev.apply", "script not found: "+applyScript, nil) // Error mismatch? IsFile returns bool
		}
	}

	// Get target repos
	targetRepos, err := getApplyTargetRepos()
	if err != nil {
		return err
	}

	if len(targetRepos) == 0 {
		return log.E("dev.apply", i18n.T("cmd.dev.apply.error.no_repos"), nil)
	}

	// Show plan
	action := applyCommand
	if applyScript != "" {
		action = applyScript
	}
	cli.Print("%s: %s\n", dimStyle.Render(i18n.T("cmd.dev.apply.action")), action)
	cli.Print("%s: %d repos\n", dimStyle.Render(i18n.T("cmd.dev.apply.targets")), len(targetRepos))
	if applyDryRun {
		cli.Print("%s\n", warningStyle.Render(i18n.T("cmd.dev.apply.dry_run_mode")))
	}
	cli.Blank()

	// Require confirmation unless --yes or --dry-run
	if !applyYes && !applyDryRun {
		cli.Print("%s\n", warningStyle.Render(i18n.T("cmd.dev.apply.warning")))
		cli.Blank()

		if !cli.Confirm(i18n.T("cmd.dev.apply.confirm"), cli.Required()) {
			cli.Print("%s\n", dimStyle.Render(i18n.T("cmd.dev.apply.cancelled")))
			return nil
		}
		cli.Blank()
	}

	var succeeded, skipped, failed int

	for _, repo := range targetRepos {
		repoName := core.PathBase(repo.Path)

		if applyDryRun {
			cli.Print("  %s %s\n", dimStyle.Render("[dry-run]"), repoName)
			succeeded++
			continue
		}

		// Step 1: Run command or script
		var cmdErr error
		if applyCommand != "" {
			cmdErr = runCommandInRepo(ctx, repo.Path, applyCommand)
		} else {
			cmdErr = runScriptInRepo(ctx, repo.Path, applyScript)
		}

		if cmdErr != nil {
			cli.Print("  %s %s: %s\n", errorStyle.Render("x"), repoName, cmdErr)
			failed++
			if !applyContinue {
				return cli.Err("%s", i18n.T("cmd.dev.apply.error.command_failed"))
			}
			continue
		}

		// Step 2: Check if anything changed
		statuses := git.Status(ctx, git.StatusOptions{
			Paths: []string{repo.Path},
			Names: map[string]string{repo.Path: repoName},
		})
		if len(statuses) == 0 || !statuses[0].IsDirty() {
			cli.Print("  %s %s: %s\n", dimStyle.Render("-"), repoName, i18n.T("cmd.dev.apply.no_changes"))
			skipped++
			continue
		}

		// Step 3: Commit if requested
		if applyCommit {
			commitMsg := applyMessage
			if applyCoAuthor != "" {
				commitMsg += "\n\nCo-Authored-By: " + applyCoAuthor
			}

			// Stage all changes
			if _, err := gitCommandQuiet(ctx, repo.Path, "add", "-A"); err != nil {
				cli.Print("  %s %s: stage failed: %s\n", errorStyle.Render("x"), repoName, err)
				failed++
				if !applyContinue {
					return err
				}
				continue
			}

			// Commit
			if _, err := gitCommandQuiet(ctx, repo.Path, "commit", "-m", commitMsg); err != nil {
				cli.Print("  %s %s: commit failed: %s\n", errorStyle.Render("x"), repoName, err)
				failed++
				if !applyContinue {
					return err
				}
				continue
			}

			// Step 4: Push if requested
			if applyPush {
				if err := safePush(ctx, repo.Path); err != nil {
					cli.Print("  %s %s: push failed: %s\n", errorStyle.Render("x"), repoName, err)
					failed++
					if !applyContinue {
						return err
					}
					continue
				}
			}
		}

		cli.Print("  %s %s\n", successStyle.Render("v"), repoName)
		succeeded++
	}

	// Summary
	cli.Blank()
	cli.Print("%s: ", i18n.T("cmd.dev.apply.summary"))
	if succeeded > 0 {
		cli.Print("%s", successStyle.Render(i18n.T("common.count.succeeded", map[string]any{"Count": succeeded})))
	}
	if skipped > 0 {
		if succeeded > 0 {
			cli.Print(", ")
		}
		cli.Print("%s", dimStyle.Render(i18n.T("common.count.skipped", map[string]any{"Count": skipped})))
	}
	if failed > 0 {
		if succeeded > 0 || skipped > 0 {
			cli.Print(", ")
		}
		cli.Print("%s", errorStyle.Render(i18n.T("common.count.failed", map[string]any{"Count": failed})))
	}
	cli.Blank()

	return nil
}

// getApplyTargetRepos gets repos to apply command to
func getApplyTargetRepos() ([]*repos.Repo, coreFailure) {
	// Load registry
	registryPath, err := repos.FindRegistry(io.Local)
	if err != nil {
		return nil, log.E("dev.apply", "failed to find registry", err)
	}

	registry, err := repos.LoadRegistry(io.Local, registryPath)
	if err != nil {
		return nil, log.E("dev.apply", "failed to load registry", err)
	}

	return filterTargetRepos(registry, applyRepos), nil
}

// filterTargetRepos selects repos by exact name/path or glob pattern.
func filterTargetRepos(registry *repos.Registry, selection string) []*repos.Repo {
	repoNames := make([]string, 0, len(registry.Repos))
	for name := range registry.Repos {
		repoNames = append(repoNames, name)
	}
	sort.Strings(repoNames)

	if selection == "" {
		matched := make([]*repos.Repo, 0, len(repoNames))
		for _, name := range repoNames {
			matched = append(matched, registry.Repos[name])
		}
		return matched
	}

	patterns := splitPatterns(selection)
	var matched []*repos.Repo

	for _, name := range repoNames {
		repo := registry.Repos[name]
		for _, candidate := range patterns {
			if matchGlob(repo.Name, candidate) || matchGlob(repo.Path, candidate) {
				matched = append(matched, repo)
				break
			}
		}
	}

	return matched
}

// runCommandInRepo runs a shell command in a repo directory
func runCommandInRepo(ctx context.Context, repoPath, command string) (_ coreFailure) {
	// Use shell to execute command
	var cmd *coreexec.Cmd
	if isWindows() {
		cmd = coreexec.Command(ctx, "cmd", "/C", command)
	} else {
		cmd = coreexec.Command(ctx, "sh", "-c", command)
	}
	cmd = cmd.WithDir(repoPath).WithStdout(core.Stdout()).WithStderr(core.Stderr())

	return resultError(cmd.Run())
}

// runScriptInRepo runs a script in a repo directory
func runScriptInRepo(ctx context.Context, repoPath, scriptPath string) (_ coreFailure) {
	// Get absolute path to script
	absResult := core.PathAbs(scriptPath)
	if !absResult.OK {
		return absResult.Value.(error)
	}
	absScript := absResult.Value.(string)

	var cmd *coreexec.Cmd
	if isWindows() {
		cmd = coreexec.Command(ctx, "cmd", "/C", absScript)
	} else {
		// Execute script directly to honor shebang
		cmd = coreexec.Command(ctx, absScript)
	}
	cmd = cmd.WithDir(repoPath).WithStdout(core.Stdout()).WithStderr(core.Stderr())

	return resultError(cmd.Run())
}

// isWindows returns true if running on Windows
func isWindows() bool {
	return core.PathSeparator == '\\'
}
