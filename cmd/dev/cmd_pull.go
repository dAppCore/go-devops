package dev

import (
	"context"

	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
	"dappco.re/go/i18n"
	coreexec "dappco.re/go/process/exec"
	"dappco.re/go/scm/git"
)

// Pull command flags
var (
	pullRegistryPath string
	pullAll          bool
)

// AddPullCommand adds the 'pull' command to the given parent command.
func AddPullCommand(parent *cli.Command) {
	pullCmd := &cli.Command{
		Use:   "pull",
		Short: i18n.T("cmd.dev.pull.short"),
		Long:  i18n.T("cmd.dev.pull.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			return resultToError(runPull(pullRegistryPath, pullAll))
		},
	}

	pullCmd.Flags().StringVar(&pullRegistryPath, "registry", "", i18n.T("common.flag.registry"))
	pullCmd.Flags().BoolVar(&pullAll, "all", false, i18n.T("cmd.dev.pull.flag.all"))

	parent.AddCommand(pullCmd)
}

func runPull(registryPath string, all bool) (_ core.Result) {
	ctx := context.Background()

	// Find or use provided registry
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

	// Find repos to pull
	var toPull []git.RepoStatus
	for _, s := range statuses {
		if s.Error != nil {
			continue
		}
		if all || s.HasUnpulled() {
			toPull = append(toPull, s)
		}
	}

	if len(toPull) == 0 {
		cli.Text(i18n.T("cmd.dev.pull.all_up_to_date"))
		return core.Ok(nil)
	}

	// Show what we're pulling
	if all {
		cli.Print("\n%s\n\n", i18n.T("cmd.dev.pull.pulling_repos", map[string]any{"Count": len(toPull)}))
	} else {
		cli.Print("\n%s\n\n", i18n.T("cmd.dev.pull.repos_behind", map[string]any{"Count": len(toPull)}))
		for _, s := range toPull {
			cli.Print("  %s: %s\n",
				repoNameStyle.Render(s.Name),
				dimStyle.Render(i18n.T("cmd.dev.pull.commits_behind", map[string]any{"Count": s.Behind})),
			)
		}
		cli.Blank()
	}

	// Pull each repo
	var succeeded, failed int
	for _, s := range toPull {
		cli.Print("  %s %s... ", dimStyle.Render(i18n.T("cmd.dev.pull.pulling")), s.Name)

		r := gitPull(ctx, s.Path)
		if !r.OK {
			cli.Print("%s\n", errorStyle.Render("x "+r.Error()))
			failed++
		} else {
			cli.Print("%s\n", successStyle.Render("v"))
			succeeded++
		}
	}

	// Summary
	cli.Blank()
	cli.Print("%s", successStyle.Render(i18n.T("cmd.dev.pull.done_pulled", map[string]any{"Count": succeeded})))
	if failed > 0 {
		cli.Print(", %s", errorStyle.Render(i18n.T("common.count.failed", map[string]any{"Count": failed})))
	}
	cli.Blank()

	return core.Ok(nil)
}

func gitPull(ctx context.Context, path string) (_ core.Result) {
	cmd := coreexec.Command(ctx, "git", "pull", "--ff-only").WithDir(path)
	r := cmd.CombinedOutput()
	if !r.OK {
		return core.Fail(cli.Err("%s", r.Value.(error).Error()))
	}
	return core.Ok(nil)
}
