package dev

import (
	"context"

	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
	"dappco.re/go/i18n"
	coreio "dappco.re/go/io"
	"dappco.re/go/scm/git"
)

// AddCommitCommand adds the 'commit' command under prefix (e.g. "dev" or "git").
//
//	c := core.New()
//	if r := dev.AddCommitCommand(c, "dev"); !r.OK { return r }
func AddCommitCommand(c *core.Core, prefix string) core.Result {
	return c.Command(prefix+"/commit", core.Command{
		Description: i18n.T("cmd.dev.commit.short"),
		Flags: core.NewOptions(
			core.Option{Key: "registry", Value: ""},
			core.Option{Key: "all", Value: false},
		),
		Action: func(o core.Options) core.Result {
			return runCommit(o.String("registry"), o.Bool("all"))
		},
	})
}

func runCommit(registryPath string, all bool) (_ core.Result) {
	ctx := context.Background()
	cwd := "."
	if cwdResult := core.Getwd(); cwdResult.OK {
		cwd = cwdResult.Value.(string)
	}

	// Check if current directory is a git repo (single-repo mode)
	if registryPath == "" && isGitRepo(cwd) {
		return runCommitSingleRepo(ctx, cwd, all)
	}

	// Multi-repo mode: find or use provided registry
	reg, regDir, r := loadRegistryWithConfig(registryPath)
	if !r.OK {
		return r
	}
	registryPath = regDir // Use resolved registry directory for relative paths

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

	// Find dirty repos
	var dirtyRepos []git.RepoStatus
	for _, s := range statuses {
		if s.Error == nil && s.IsDirty() {
			dirtyRepos = append(dirtyRepos, s)
		}
	}

	if len(dirtyRepos) == 0 {
		cli.Text(i18n.T("cmd.dev.no_changes"))
		return core.Ok(nil)
	}

	// Show dirty repos
	cli.Print("\n%s\n\n", i18n.T("cmd.dev.repos_with_changes", map[string]any{"Count": len(dirtyRepos)}))
	for _, s := range dirtyRepos {
		cli.Print("  %s: ", repoNameStyle.Render(s.Name))
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
	}

	// Confirm unless --all
	if !all {
		cli.Blank()
		if !cli.Confirm(i18n.T("cmd.dev.confirm_claude_commit")) {
			cli.Text(i18n.T("cli.aborted"))
			return core.Ok(nil)
		}
	}

	cli.Blank()

	// Commit each dirty repo
	var succeeded, failed int
	for _, s := range dirtyRepos {
		cli.Print("%s %s\n", dimStyle.Render(i18n.T("cmd.dev.committing")), s.Name)

		if r := doCommit(ctx, s.Path, false); !r.OK {
			cli.Print("  %s %s\n", errorStyle.Render("x"), r.Error())
			failed++
		} else {
			cli.Print("  %s %s\n", successStyle.Render("v"), i18n.T("cmd.dev.committed"))
			succeeded++
		}
		cli.Blank()
	}

	// Summary
	cli.Print("%s", successStyle.Render(i18n.T("cmd.dev.done_succeeded", map[string]any{"Count": succeeded})))
	if failed > 0 {
		cli.Print(", %s", errorStyle.Render(i18n.T("common.count.failed", map[string]any{"Count": failed})))
	}
	cli.Blank()

	return core.Ok(nil)
}

// isGitRepo checks if a directory is a git repository.
func isGitRepo(path string) bool {
	gitDir := path + "/.git"
	_, err := coreio.Local.List(gitDir)
	return err == nil
}

// runCommitSingleRepo handles commit for a single repo (current directory).
func runCommitSingleRepo(ctx context.Context, repoPath string, all bool) (_ core.Result) {
	repoName := core.PathBase(repoPath)

	// Get status
	statuses := git.Status(ctx, git.StatusOptions{
		Paths: []string{repoPath},
		Names: map[string]string{repoPath: repoName},
	})

	if len(statuses) == 0 || statuses[0].Error != nil {
		if len(statuses) > 0 && statuses[0].Error != nil {
			return core.Fail(statuses[0].Error)
		}
		return cli.Err("failed to get repo status")
	}

	s := statuses[0]
	if !s.IsDirty() {
		cli.Text(i18n.T("cmd.dev.no_changes"))
		return core.Ok(nil)
	}

	// Show status
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

	// Confirm unless --all
	if !all {
		cli.Blank()
		if !cli.Confirm(i18n.T("cmd.dev.confirm_claude_commit")) {
			cli.Text(i18n.T("cli.aborted"))
			return core.Ok(nil)
		}
	}

	cli.Blank()

	// Commit
	if r := doCommit(ctx, repoPath, false); !r.OK {
		cli.Print("  %s %s\n", errorStyle.Render("x"), r.Error())
		return r
	}
	cli.Print("  %s %s\n", successStyle.Render("v"), i18n.T("cmd.dev.committed"))
	return core.Ok(nil)
}
