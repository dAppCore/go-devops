package dev

import (
	"context"

	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
	"dappco.re/go/i18n"
	"dappco.re/go/scm/git"
)

// Push command flags
var (
	pushRegistryPath string
	pushForce        bool
)

// AddPushCommand adds the 'push' command to the given parent command.
func AddPushCommand(parent *cli.Command) {
	pushCmd := &cli.Command{
		Use:   "push",
		Short: i18n.T("cmd.dev.push.short"),
		Long:  i18n.T("cmd.dev.push.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			return resultToError(runPush(pushRegistryPath, pushForce))
		},
	}

	pushCmd.Flags().StringVar(&pushRegistryPath, "registry", "", i18n.T("common.flag.registry"))
	pushCmd.Flags().BoolVarP(&pushForce, "force", "f", false, i18n.T("cmd.dev.push.flag.force"))

	parent.AddCommand(pushCmd)
}

func runPush(registryPath string, force bool) (_ core.Result) {
	ctx := context.Background()
	cwd := "."
	if cwdResult := core.Getwd(); cwdResult.OK {
		cwd = cwdResult.Value.(string)
	}

	// Check if current directory is a git repo (single-repo mode)
	if registryPath == "" && isGitRepo(cwd) {
		return runPushSingleRepo(ctx, cwd, force)
	}

	// Multi-repo mode: find or use provided registry
	reg, _, r := loadRegistryWithConfig(registryPath)
	if !r.OK {
		return r
	}

	// Build paths and names for git operations
	var paths []string
	names := make(map[string]string)

	for _, repo := range reg.List() {
		if repo.IsGitRepo() {
			paths = append(paths, repo.Path)
			names[repo.Path] = repo.Name
		}
	}

	if len(paths) == 0 {
		cli.Text(i18n.T("cmd.dev.no_git_repos"))
		return core.Ok(nil)
	}

	// Get status for all repos
	statuses := git.Status(ctx, git.StatusOptions{
		Paths: paths,
		Names: names,
	})

	// Find repos with unpushed commits
	var aheadRepos []git.RepoStatus
	for _, s := range statuses {
		if s.Error == nil && s.HasUnpushed() {
			aheadRepos = append(aheadRepos, s)
		}
	}

	if len(aheadRepos) == 0 {
		cli.Text(i18n.T("cmd.dev.push.all_up_to_date"))
		return core.Ok(nil)
	}

	// Show repos to push
	cli.Print("\n%s\n\n", i18n.T("common.count.repos_unpushed", map[string]any{"Count": len(aheadRepos)}))
	totalCommits := 0
	for _, s := range aheadRepos {
		cli.Print("  %s: %s\n",
			repoNameStyle.Render(s.Name),
			aheadStyle.Render(i18n.T("common.count.commits", map[string]any{"Count": s.Ahead})),
		)
		totalCommits += s.Ahead
	}

	// Confirm unless --force
	if !force {
		cli.Blank()
		if !cli.Confirm(i18n.T("cmd.dev.push.confirm_push", map[string]any{"Commits": totalCommits, "Repos": len(aheadRepos)})) {
			cli.Text(i18n.T("cli.aborted"))
			return core.Ok(nil)
		}
	}

	cli.Blank()

	// Push sequentially (SSH passphrase needs interaction)
	var pushPaths []string
	for _, s := range aheadRepos {
		pushPaths = append(pushPaths, s.Path)
	}

	results := git.PushMultiple(ctx, pushPaths, names)

	var succeeded, failed int
	var divergedRepos []git.PushResult

	for _, r := range results {
		if r.Success {
			cli.Print("  %s %s\n", successStyle.Render("v"), r.Name)
			succeeded++
		} else {
			// Check if this is a non-fast-forward error (diverged branch)
			if git.IsNonFastForward(r.Error) {
				cli.Print("  %s %s: %s\n", warningStyle.Render("!"), r.Name, i18n.T("cmd.dev.push.diverged"))
				divergedRepos = append(divergedRepos, r)
			} else {
				cli.Print("  %s %s: %s\n", errorStyle.Render("x"), r.Name, r.Error)
			}
			failed++
		}
	}

	// Handle diverged repos - offer to pull and retry
	if len(divergedRepos) > 0 {
		cli.Blank()
		cli.Print("%s\n", i18n.T("cmd.dev.push.diverged_help"))
		if cli.Confirm(i18n.T("cmd.dev.push.pull_and_retry")) {
			cli.Blank()
			for _, r := range divergedRepos {
				cli.Print("  %s %s...\n", dimStyle.Render("↓"), r.Name)
				if err := git.Pull(ctx, r.Path); err != nil {
					cli.Print("  %s %s: %s\n", errorStyle.Render("x"), r.Name, err)
					continue
				}
				cli.Print("  %s %s...\n", dimStyle.Render("↑"), r.Name)
				if err := git.Push(ctx, r.Path); err != nil {
					cli.Print("  %s %s: %s\n", errorStyle.Render("x"), r.Name, err)
					continue
				}
				cli.Print("  %s %s\n", successStyle.Render("v"), r.Name)
				succeeded++
				failed--
			}
		}
	}

	// Summary
	cli.Blank()
	cli.Print("%s", successStyle.Render(i18n.T("cmd.dev.push.done_pushed", map[string]any{"Count": succeeded})))
	if failed > 0 {
		cli.Print(", %s", errorStyle.Render(i18n.T("common.count.failed", map[string]any{"Count": failed})))
	}
	cli.Blank()

	return core.Ok(nil)
}

// runPushSingleRepo handles push for a single repo (current directory).
func runPushSingleRepo(ctx context.Context, repoPath string, force bool) (_ core.Result) {
	repoName := core.PathBase(repoPath)

	// Get status
	statuses := git.Status(ctx, git.StatusOptions{
		Paths: []string{repoPath},
		Names: map[string]string{repoPath: repoName},
	})

	if len(statuses) == 0 {
		return core.Fail(cli.Err("failed to get repo status"))
	}

	s := statuses[0]
	if s.Error != nil {
		return core.Fail(s.Error)
	}

	if !s.HasUnpushed() {
		// Check if there are uncommitted changes
		if s.IsDirty() {
			cli.Print("%s: ", repoNameStyle.Render(s.Name))
			if s.Modified > 0 {
				cli.Print("%s ", dirtyStyle.Render(i18n.T("cmd.dev.modified", map[string]any{"Count": s.Modified})))
			}
			if s.Untracked > 0 {
				cli.Print("%s ", dirtyStyle.Render(i18n.T("cmd.dev.untracked", map[string]any{"Count": s.Untracked})))
			}
			if s.Staged > 0 {
				cli.Print("%s ", aheadStyle.Render(i18n.T("cmd.dev.staged", map[string]any{"Count": s.Staged})))
			}
			cli.Blank()
			cli.Blank()
			if cli.Confirm(i18n.T("cmd.dev.push.uncommitted_changes_commit")) {
				cli.Blank()
				// Use edit-enabled commit if only untracked files (may need .gitignore fix)
				var r core.Result
				if s.Modified == 0 && s.Staged == 0 && s.Untracked > 0 {
					r = doCommit(ctx, repoPath, true)
				} else {
					r = runCommitSingleRepo(ctx, repoPath, false)
				}
				if !r.OK {
					return r
				}
				// Re-check - only push if Claude created commits
				newStatuses := git.Status(ctx, git.StatusOptions{
					Paths: []string{repoPath},
					Names: map[string]string{repoPath: repoName},
				})
				if len(newStatuses) > 0 && newStatuses[0].HasUnpushed() {
					return runPushSingleRepo(ctx, repoPath, force)
				}
			}
			return core.Ok(nil)
		}
		cli.Text(i18n.T("cmd.dev.push.all_up_to_date"))
		return core.Ok(nil)
	}

	// Show commits to push
	cli.Print("%s: %s\n", repoNameStyle.Render(s.Name),
		aheadStyle.Render(i18n.T("common.count.commits", map[string]any{"Count": s.Ahead})))

	// Confirm unless --force
	if !force {
		cli.Blank()
		if !cli.Confirm(i18n.T("cmd.dev.push.confirm_push", map[string]any{"Commits": s.Ahead, "Repos": 1})) {
			cli.Text(i18n.T("cli.aborted"))
			return core.Ok(nil)
		}
	}

	cli.Blank()

	// Push
	err := git.Push(ctx, repoPath)
	if err != nil {
		if git.IsNonFastForward(err) {
			cli.Print("  %s %s: %s\n", warningStyle.Render("!"), repoName, i18n.T("cmd.dev.push.diverged"))
			cli.Blank()
			cli.Print("%s\n", i18n.T("cmd.dev.push.diverged_help"))
			if cli.Confirm(i18n.T("cmd.dev.push.pull_and_retry")) {
				cli.Blank()
				cli.Print("  %s %s...\n", dimStyle.Render("↓"), repoName)
				if pullErr := git.Pull(ctx, repoPath); pullErr != nil {
					cli.Print("  %s %s: %s\n", errorStyle.Render("x"), repoName, pullErr)
					return core.Fail(pullErr)
				}
				cli.Print("  %s %s...\n", dimStyle.Render("↑"), repoName)
				if pushErr := git.Push(ctx, repoPath); pushErr != nil {
					cli.Print("  %s %s: %s\n", errorStyle.Render("x"), repoName, pushErr)
					return core.Fail(pushErr)
				}
				cli.Print("  %s %s\n", successStyle.Render("v"), repoName)
				return core.Ok(nil)
			}
		}
		cli.Print("  %s %s: %s\n", errorStyle.Render("x"), repoName, err)
		return core.Fail(err)
	}

	cli.Print("  %s %s\n", successStyle.Render("v"), repoName)
	return core.Ok(nil)
}
