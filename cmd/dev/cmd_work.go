package dev

import (
	"cmp"
	"context"
	"slices"

	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
	"dappco.re/go/i18n"
	"dappco.re/go/scm/git"
)

// Work command flags
var (
	workStatusOnly   bool
	workAutoCommit   bool
	workRegistryPath string
)

// AddWorkCommand adds the 'work' command to the given parent command.
func AddWorkCommand(parent *cli.Command) {
	workCmd := &cli.Command{
		Use:   "work",
		Short: i18n.T("cmd.dev.work.short"),
		Long:  i18n.T("cmd.dev.work.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			return resultToError(runWork(workRegistryPath, workStatusOnly, workAutoCommit))
		},
	}

	workCmd.Flags().BoolVar(&workStatusOnly, "status", false, i18n.T("cmd.dev.work.flag.status"))
	workCmd.Flags().BoolVar(&workAutoCommit, "commit", false, i18n.T("cmd.dev.work.flag.commit"))
	workCmd.Flags().StringVar(&workRegistryPath, "registry", "", i18n.T("common.flag.registry"))

	parent.AddCommand(workCmd)
}

func runWork(registryPath string, statusOnly, autoCommit bool) (_ core.Result) {
	ctx := context.Background()

	// Build worker bundle with required services
	bundle, err := NewWorkBundle(WorkBundleOptions{
		RegistryPath: registryPath,
	})
	if !err.OK {
		return err
	}

	// Start services (registers handlers)
	if r := bundle.Start(ctx); !r.OK {
		return r
	}
	defer func() {
		if r := bundle.Stop(ctx); !r.OK {
			cli.Print("  %s %s\n", errorStyle.Render("x"), r.Value.(error))
		}
	}()

	// Load registry and get paths
	reg, _, r := loadRegistryWithConfig(registryPath)
	if !r.OK {
		return r
	}

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

	// Query git status directly
	statuses := git.Status(ctx, git.StatusOptions{
		Paths: paths,
		Names: names,
	})

	// Sort by repo name for consistent output
	slices.SortFunc(statuses, func(a, b git.RepoStatus) int {
		return cmp.Compare(a.Name, b.Name)
	})

	// Display status table
	printStatusTable(statuses)

	// Collect dirty and ahead repos
	var dirtyRepos []git.RepoStatus
	var aheadRepos []git.RepoStatus

	for _, s := range statuses {
		if s.Error != nil {
			continue
		}
		if s.IsDirty() {
			dirtyRepos = append(dirtyRepos, s)
		}
		if s.HasUnpushed() {
			aheadRepos = append(aheadRepos, s)
		}
	}

	// Auto-commit dirty repos if requested
	if autoCommit && len(dirtyRepos) > 0 {
		cli.Blank()
		cli.Print("%s\n", cli.TitleStyle.Render(i18n.T("cmd.dev.commit.committing")))
		cli.Blank()

		for _, s := range dirtyRepos {
			r := doCommit(ctx, s.Path, false)
			if !r.OK {
				cli.Print("  %s %s: %s\n", errorStyle.Render("x"), s.Name, r.Value.(error))
			} else {
				cli.Print("  %s %s\n", successStyle.Render("v"), s.Name)
			}
		}

		// Re-query status after commits
		statuses = git.Status(ctx, git.StatusOptions{
			Paths: paths,
			Names: names,
		})

		// Rebuild ahead repos list
		aheadRepos = nil
		for _, s := range statuses {
			if s.Error == nil && s.HasUnpushed() {
				aheadRepos = append(aheadRepos, s)
			}
		}
	}

	// If status only, we're done
	if statusOnly {
		if len(dirtyRepos) > 0 && !autoCommit {
			cli.Blank()
			cli.Print("%s\n", dimStyle.Render(i18n.T("cmd.dev.work.use_commit_flag")))
		}
		return core.Ok(nil)
	}

	// Push repos with unpushed commits
	if len(aheadRepos) == 0 {
		cli.Blank()
		cli.Text(i18n.T("cmd.dev.work.all_up_to_date"))
		return core.Ok(nil)
	}

	cli.Blank()
	cli.Print("%s\n", i18n.T("common.count.repos_unpushed", map[string]any{"Count": len(aheadRepos)}))
	for _, s := range aheadRepos {
		cli.Print("  %s: %s\n", s.Name, i18n.T("common.count.commits", map[string]any{"Count": s.Ahead}))
	}

	cli.Blank()
	if !cli.Confirm(i18n.T("cmd.dev.push.confirm")) {
		cli.Text(i18n.T("cli.aborted"))
		return core.Ok(nil)
	}

	cli.Blank()

	// Push each repo directly
	var divergedRepos []git.RepoStatus

	for _, s := range aheadRepos {
		err := git.Push(ctx, s.Path)
		if err != nil {
			if git.IsNonFastForward(err) {
				cli.Print("  %s %s: %s\n", warningStyle.Render("!"), s.Name, i18n.T("cmd.dev.push.diverged"))
				divergedRepos = append(divergedRepos, s)
			} else {
				cli.Print("  %s %s: %s\n", errorStyle.Render("x"), s.Name, err)
			}
		} else {
			cli.Print("  %s %s\n", successStyle.Render("v"), s.Name)
		}
	}

	// Handle diverged repos - offer to pull and retry
	if len(divergedRepos) > 0 {
		cli.Blank()
		cli.Print("%s\n", i18n.T("cmd.dev.push.diverged_help"))
		if cli.Confirm(i18n.T("cmd.dev.push.pull_and_retry")) {
			cli.Blank()
			for _, s := range divergedRepos {
				cli.Print("  %s %s...\n", dimStyle.Render("↓"), s.Name)

				// Pull directly
				err := git.Pull(ctx, s.Path)
				if err != nil {
					cli.Print("  %s %s: %s\n", errorStyle.Render("x"), s.Name, err)
					continue
				}

				cli.Print("  %s %s...\n", dimStyle.Render("↑"), s.Name)

				// Push directly
				err = git.Push(ctx, s.Path)
				if err != nil {
					cli.Print("  %s %s: %s\n", errorStyle.Render("x"), s.Name, err)
					continue
				}

				cli.Print("  %s %s\n", successStyle.Render("v"), s.Name)
			}
		}
	}

	return core.Ok(nil)
}

func printStatusTable(statuses []git.RepoStatus) {
	// Calculate column widths
	nameWidth := 4 // "Repo"
	for _, s := range statuses {
		if len(s.Name) > nameWidth {
			nameWidth = len(s.Name)
		}
	}

	// Print header with fixed-width formatting
	cli.Print("%-*s  %8s  %9s  %6s  %5s\n",
		nameWidth,
		cli.TitleStyle.Render(i18n.Label("repo")),
		cli.TitleStyle.Render(i18n.T("cmd.dev.work.table_modified")),
		cli.TitleStyle.Render(i18n.T("cmd.dev.work.table_untracked")),
		cli.TitleStyle.Render(i18n.T("cmd.dev.work.table_staged")),
		cli.TitleStyle.Render(i18n.T("cmd.dev.work.table_ahead")),
	)

	// Print separator
	cli.Text(repeatStatusSeparator(nameWidth + 2 + 10 + 11 + 8 + 7))

	// Print rows
	for _, s := range statuses {
		if s.Error != nil {
			paddedName := cli.Sprintf("%-*s", nameWidth, s.Name)
			cli.Print("%s  %s\n",
				repoNameStyle.Render(paddedName),
				errorStyle.Render(i18n.T("cmd.dev.work.error_prefix")+" "+s.Error.Error()),
			)
			continue
		}

		// Style numbers based on values
		modStr := cli.Sprintf("%d", s.Modified)
		if s.Modified > 0 {
			modStr = dirtyStyle.Render(modStr)
		} else {
			modStr = cleanStyle.Render(modStr)
		}

		untrackedStr := cli.Sprintf("%d", s.Untracked)
		if s.Untracked > 0 {
			untrackedStr = dirtyStyle.Render(untrackedStr)
		} else {
			untrackedStr = cleanStyle.Render(untrackedStr)
		}

		stagedStr := cli.Sprintf("%d", s.Staged)
		if s.Staged > 0 {
			stagedStr = aheadStyle.Render(stagedStr)
		} else {
			stagedStr = cleanStyle.Render(stagedStr)
		}

		aheadStr := cli.Sprintf("%d", s.Ahead)
		if s.Ahead > 0 {
			aheadStr = aheadStyle.Render(aheadStr)
		} else {
			aheadStr = cleanStyle.Render(aheadStr)
		}

		// Pad name before styling to avoid ANSI code length issues
		paddedName := cli.Sprintf("%-*s", nameWidth, s.Name)
		cli.Print("%s  %8s  %9s  %6s  %5s\n",
			repoNameStyle.Render(paddedName),
			modStr,
			untrackedStr,
			stagedStr,
			aheadStr,
		)
	}
}

func repeatStatusSeparator(width int) string {
	parts := make([]string, width)
	for i := range parts {
		parts[i] = "-"
	}
	return core.Join("", parts...)
}
