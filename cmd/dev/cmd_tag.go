package dev

import (
	"context"
	"strconv"

	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
	"dappco.re/go/i18n"
	log "dappco.re/go/log"
	coreexec "dappco.re/go/process/exec"
)

// Tag command flags
var (
	tagRegistryPath string
	tagDryRun       bool
	tagForce        bool
)

// AddTagCommand adds the 'tag' command to the given parent command.
func AddTagCommand(parent *cli.Command) {
	tagCmd := &cli.Command{
		Use:   "tag",
		Short: i18n.T("cmd.dev.tag.short"),
		Long:  i18n.T("cmd.dev.tag.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			return runTag(tagRegistryPath, tagDryRun, tagForce)
		},
	}

	tagCmd.Flags().StringVar(&tagRegistryPath, "registry", "", i18n.T("common.flag.registry"))
	tagCmd.Flags().BoolVar(&tagDryRun, "dry-run", false, i18n.T("cmd.dev.tag.flag.dry_run"))
	tagCmd.Flags().BoolVarP(&tagForce, "force", "f", false, i18n.T("cmd.dev.tag.flag.force"))

	parent.AddCommand(tagCmd)
}

// tagPlan holds the version bump plan for a single repo.
type tagPlan struct {
	Name    string
	Path    string
	Current string // current tag (e.g. "v0.3.1")
	Next    string // next tag (e.g. "v0.3.2")
	IsGoMod bool   // whether the repo has a go.mod
}

func runTag(registryPath string, dryRun, force bool) (_ coreFailure) {
	ctx := context.Background()

	// Load registry
	reg, _, err := loadRegistryWithConfig(registryPath)
	if err != nil {
		return err
	}

	// Get topological order (dependencies first)
	ordered, err := reg.TopologicalOrder()
	if err != nil {
		return cli.Wrap(err, "failed to compute dependency order")
	}

	// Build version bump plan
	var plans []tagPlan

	for _, repo := range ordered {
		if !repo.Exists() || !repo.IsGitRepo() {
			continue
		}

		current, err := latestTag(ctx, repo.Path)
		if err != nil || current == "" {
			current = "v0.0.0"
		}

		next, err := bumpPatch(current)
		if err != nil {
			return log.E("dev.tag", core.Sprintf("%s: failed to bump version %s", repo.Name, current), err)
		}

		hasGoMod := fileExists(core.PathJoin(repo.Path, "go.mod"))

		plans = append(plans, tagPlan{
			Name:    repo.Name,
			Path:    repo.Path,
			Current: current,
			Next:    next,
			IsGoMod: hasGoMod,
		})
	}

	if len(plans) == 0 {
		cli.Text(i18n.T("cmd.dev.no_git_repos"))
		return nil
	}

	// Show plan
	cli.Print("\n%s\n\n", cli.TitleStyle.Render("Tag plan (dependency order)"))

	nameWidth := 4
	for _, p := range plans {
		if len(p.Name) > nameWidth {
			nameWidth = len(p.Name)
		}
	}

	for _, p := range plans {
		paddedName := cli.Sprintf("%-*s", nameWidth, p.Name)
		cli.Print("  %s  %s → %s\n",
			repoNameStyle.Render(paddedName),
			dimStyle.Render(p.Current),
			aheadStyle.Render(p.Next),
		)
	}

	if dryRun {
		cli.Blank()
		cli.Text("Dry run — no changes made.")
		return nil
	}

	// Confirm unless --force
	if !force {
		cli.Blank()
		if !cli.Confirm(core.Sprintf("Tag and push %d repos?", len(plans))) {
			cli.Text(i18n.T("cli.aborted"))
			return nil
		}
	}

	cli.Blank()

	// Execute: for each repo in dependency order
	var succeeded, failed int

	for _, p := range plans {
		cli.Print("%s %s → %s\n", dimStyle.Render("▸"), repoNameStyle.Render(p.Name), aheadStyle.Render(p.Next))

		if p.IsGoMod {
			// Step 1: GOWORK=off go get -u ./...
			if err := goGetUpdate(ctx, p.Path); err != nil {
				cli.Print("  %s go get -u: %s\n", errorStyle.Render("x"), err)
				failed++
				continue
			}

			// Step 2: GOWORK=off go mod tidy
			if err := goModTidy(ctx, p.Path); err != nil {
				cli.Print("  %s go mod tidy: %s\n", errorStyle.Render("x"), err)
				failed++
				continue
			}

			// Step 3: Commit go.mod/go.sum if changed
			if err := commitGoMod(ctx, p.Path, p.Next); err != nil {
				cli.Print("  %s commit: %s\n", errorStyle.Render("x"), err)
				failed++
				continue
			}
		}

		// Step 4: Create annotated tag
		if err := createTag(ctx, p.Path, p.Next); err != nil {
			cli.Print("  %s tag: %s\n", errorStyle.Render("x"), err)
			failed++
			continue
		}

		// Step 5: Push commits and tags
		if err := pushWithTags(ctx, p.Path); err != nil {
			cli.Print("  %s push: %s\n", errorStyle.Render("x"), err)
			failed++
			continue
		}

		cli.Print("  %s %s\n", successStyle.Render("v"), p.Next)
		succeeded++
	}

	// Summary
	cli.Blank()
	cli.Print("%s", successStyle.Render(core.Sprintf("%d tagged", succeeded)))
	if failed > 0 {
		cli.Print(", %s", errorStyle.Render(i18n.T("common.count.failed", map[string]any{"Count": failed})))
	}
	cli.Blank()

	return nil
}

// latestTag returns the latest semver tag in the repo.
func latestTag(ctx context.Context, repoPath string) (string, coreFailure) {
	cmd := coreexec.Command(ctx, "git", "describe", "--tags", "--abbrev=0", "--match", "v*").WithDir(repoPath)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return core.Trim(string(out)), nil
}

// bumpPatch increments the patch version of a semver tag.
// "v0.3.1" → "v0.3.2"
func bumpPatch(tag string) (string, coreFailure) {
	v := core.TrimPrefix(tag, "v")
	parts := core.Split(v, ".")
	if len(parts) != 3 {
		return "", log.E("dev.tag", core.Sprintf("invalid semver: %s", tag), nil)
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return "", log.E("dev.tag", core.Sprintf("invalid patch version: %s", parts[2]), nil)
	}
	return core.Sprintf("v%s.%s.%d", parts[0], parts[1], patch+1), nil
}

// goGetUpdate runs GOWORK=off go get -u ./... in the repo.
func goGetUpdate(ctx context.Context, repoPath string) (_ coreFailure) {
	cmd := coreexec.Command(ctx, "go", "get", "-u", "./...").
		WithDir(repoPath).
		WithEnv(append(core.Environ(), "GOWORK=off"))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return log.E("dev.tag", core.Trim(string(out)), err)
	}
	return nil
}

// goModTidy runs GOWORK=off go mod tidy in the repo.
func goModTidy(ctx context.Context, repoPath string) (_ coreFailure) {
	cmd := coreexec.Command(ctx, "go", "mod", "tidy").
		WithDir(repoPath).
		WithEnv(append(core.Environ(), "GOWORK=off"))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return log.E("dev.tag", core.Trim(string(out)), err)
	}
	return nil
}

// commitGoMod stages and commits go.mod and go.sum if they have changes.
func commitGoMod(ctx context.Context, repoPath, version string) (_ coreFailure) {
	// Check if go.mod or go.sum changed (staged or unstaged)
	diffCmd := coreexec.Command(ctx, "git", "diff", "--quiet", "go.mod", "go.sum").WithDir(repoPath)
	modChanged := !diffCmd.Run().OK

	// Also check for untracked go.sum
	lsCmd := coreexec.Command(ctx, "git", "ls-files", "--others", "--exclude-standard", "go.sum").WithDir(repoPath)
	lsOut, _ := lsCmd.Output()
	untrackedSum := core.Trim(string(lsOut)) != ""

	if !modChanged && !untrackedSum {
		return nil // No changes
	}

	// Stage go.mod and go.sum
	addCmd := coreexec.Command(ctx, "git", "add", "go.mod", "go.sum").WithDir(repoPath)
	if out, err := addCmd.CombinedOutput(); err != nil {
		return log.E("dev.tag", "git add: "+core.Trim(string(out)), err)
	}

	// Check if anything is actually staged
	stagedCmd := coreexec.Command(ctx, "git", "diff", "--cached", "--quiet").WithDir(repoPath)
	if stagedCmd.Run().OK {
		return nil // Nothing staged
	}

	// Commit
	msg := core.Sprintf("chore: sync dependencies for %s\n\nCo-Authored-By: Virgil <virgil@lethean.io>", version)
	commitCmd := coreexec.Command(ctx, "git", "commit", "-m", msg).WithDir(repoPath)
	if out, err := commitCmd.CombinedOutput(); err != nil {
		return log.E("dev.tag", "git commit: "+core.Trim(string(out)), err)
	}
	return nil
}

// createTag creates an annotated tag.
func createTag(ctx context.Context, repoPath, tag string) (_ coreFailure) {
	cmd := coreexec.Command(ctx, "git", "tag", "-a", tag, "-m", tag).WithDir(repoPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return log.E("dev.tag", core.Trim(string(out)), err)
	}
	return nil
}

// pushWithTags pushes commits and tags to the remote.
// Uses interactive mode to support SSH passphrase prompts.
func pushWithTags(ctx context.Context, repoPath string) (_ coreFailure) {
	cmd := coreexec.Command(ctx, "git", "push", "--follow-tags").
		WithDir(repoPath).
		WithStdout(core.Stdout()).
		WithStderr(core.Stderr()).
		WithStdin(core.Stdin())
	return resultError(cmd.Run())
}

// fileExists checks if a file exists at the given path.
func fileExists(path string) bool {
	return core.Stat(path).OK
}
