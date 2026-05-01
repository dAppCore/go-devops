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
			return resultToError(runTag(tagRegistryPath, tagDryRun, tagForce))
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

func runTag(registryPath string, dryRun, force bool) (_ core.Result) {
	ctx := context.Background()

	// Load registry
	reg, _, r := loadRegistryWithConfig(registryPath)
	if !r.OK {
		return r
	}

	// Get topological order (dependencies first)
	ordered, err := reg.TopologicalOrder()
	if err != nil {
		return core.Fail(cli.Wrap(err, "failed to compute dependency order"))
	}

	// Build version bump plan
	var plans []tagPlan

	for _, repo := range ordered {
		if !repo.Exists() || !repo.IsGitRepo() {
			continue
		}

		current, r := latestTag(ctx, repo.Path)
		if !r.OK || current == "" {
			current = "v0.0.0"
		}

		next, r := bumpPatch(current)
		if !r.OK {
			return core.Fail(log.E("dev.tag", core.Sprintf("%s: failed to bump version %s", repo.Name, current), r.Value.(error)))
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
		return core.Ok(nil)
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
		return core.Ok(nil)
	}

	// Confirm unless --force
	if !force {
		cli.Blank()
		if !cli.Confirm(core.Sprintf("Tag and push %d repos?", len(plans))) {
			cli.Text(i18n.T("cli.aborted"))
			return core.Ok(nil)
		}
	}

	cli.Blank()

	// Execute: for each repo in dependency order
	var succeeded, failed int

	for _, p := range plans {
		cli.Print("%s %s → %s\n", dimStyle.Render("▸"), repoNameStyle.Render(p.Name), aheadStyle.Render(p.Next))

		if p.IsGoMod {
			// Step 1: GOWORK=off go get -u ./...
			if r := goGetUpdate(ctx, p.Path); !r.OK {
				cli.Print("  %s go get -u: %s\n", errorStyle.Render("x"), r.Value.(error))
				failed++
				continue
			}

			// Step 2: GOWORK=off go mod tidy
			if r := goModTidy(ctx, p.Path); !r.OK {
				cli.Print("  %s go mod tidy: %s\n", errorStyle.Render("x"), r.Value.(error))
				failed++
				continue
			}

			// Step 3: Commit go.mod/go.sum if changed
			if r := commitGoMod(ctx, p.Path, p.Next); !r.OK {
				cli.Print("  %s commit: %s\n", errorStyle.Render("x"), r.Value.(error))
				failed++
				continue
			}
		}

		// Step 4: Create annotated tag
		if r := createTag(ctx, p.Path, p.Next); !r.OK {
			cli.Print("  %s tag: %s\n", errorStyle.Render("x"), r.Value.(error))
			failed++
			continue
		}

		// Step 5: Push commits and tags
		if r := pushWithTags(ctx, p.Path); !r.OK {
			cli.Print("  %s push: %s\n", errorStyle.Render("x"), r.Value.(error))
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

	return core.Ok(nil)
}

// latestTag returns the latest semver tag in the repo.
func latestTag(ctx context.Context, repoPath string) (string, core.Result) {
	cmd := coreexec.Command(ctx, "git", "describe", "--tags", "--abbrev=0", "--match", "v*").WithDir(repoPath)
	r := cmd.Output()
	if !r.OK {
		return "", r
	}
	return core.Trim(string(r.Value.([]byte))), core.Ok(nil)
}

// bumpPatch increments the patch version of a semver tag.
// "v0.3.1" → "v0.3.2"
func bumpPatch(tag string) (string, core.Result) {
	v := core.TrimPrefix(tag, "v")
	parts := core.Split(v, ".")
	if len(parts) != 3 {
		return "", core.Fail(log.E("dev.tag", core.Sprintf("invalid semver: %s", tag), nil))
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return "", core.Fail(log.E("dev.tag", core.Sprintf("invalid patch version: %s", parts[2]), err))
	}
	return core.Sprintf("v%s.%s.%d", parts[0], parts[1], patch+1), core.Ok(nil)
}

// goGetUpdate runs GOWORK=off go get -u ./... in the repo.
func goGetUpdate(ctx context.Context, repoPath string) (_ core.Result) {
	cmd := coreexec.Command(ctx, "go", "get", "-u", "./...").
		WithDir(repoPath).
		WithEnv(append(core.Environ(), "GOWORK=off"))
	r := cmd.CombinedOutput()
	if !r.OK {
		return core.Fail(log.E("dev.tag", "go get -u failed", r.Value.(error)))
	}
	return core.Ok(nil)
}

// goModTidy runs GOWORK=off go mod tidy in the repo.
func goModTidy(ctx context.Context, repoPath string) (_ core.Result) {
	cmd := coreexec.Command(ctx, "go", "mod", "tidy").
		WithDir(repoPath).
		WithEnv(append(core.Environ(), "GOWORK=off"))
	r := cmd.CombinedOutput()
	if !r.OK {
		return core.Fail(log.E("dev.tag", "go mod tidy failed", r.Value.(error)))
	}
	return core.Ok(nil)
}

// commitGoMod stages and commits go.mod and go.sum if they have changes.
func commitGoMod(ctx context.Context, repoPath, version string) (_ core.Result) {
	// Check if go.mod or go.sum changed (staged or unstaged)
	diffCmd := coreexec.Command(ctx, "git", "diff", "--quiet", "go.mod", "go.sum").WithDir(repoPath)
	modChanged := !diffCmd.Run().OK

	// Also check for untracked go.sum
	lsCmd := coreexec.Command(ctx, "git", "ls-files", "--others", "--exclude-standard", "go.sum").WithDir(repoPath)
	lsResult := lsCmd.Output()
	untrackedSum := lsResult.OK && core.Trim(string(lsResult.Value.([]byte))) != ""

	if !modChanged && !untrackedSum {
		return core.Ok(nil) // No changes
	}

	// Stage go.mod and go.sum
	addCmd := coreexec.Command(ctx, "git", "add", "go.mod", "go.sum").WithDir(repoPath)
	if r := addCmd.CombinedOutput(); !r.OK {
		return core.Fail(log.E("dev.tag", "git add failed", r.Value.(error)))
	}

	// Check if anything is actually staged
	stagedCmd := coreexec.Command(ctx, "git", "diff", "--cached", "--quiet").WithDir(repoPath)
	if stagedCmd.Run().OK {
		return core.Ok(nil) // Nothing staged
	}

	// Commit
	msg := core.Sprintf("chore: sync dependencies for %s\n\nCo-Authored-By: Virgil <virgil@lethean.io>", version)
	commitCmd := coreexec.Command(ctx, "git", "commit", "-m", msg).WithDir(repoPath)
	if r := commitCmd.CombinedOutput(); !r.OK {
		return core.Fail(log.E("dev.tag", "git commit failed", r.Value.(error)))
	}
	return core.Ok(nil)
}

// createTag creates an annotated tag.
func createTag(ctx context.Context, repoPath, tag string) (_ core.Result) {
	cmd := coreexec.Command(ctx, "git", "tag", "-a", tag, "-m", tag).WithDir(repoPath)
	if r := cmd.CombinedOutput(); !r.OK {
		return core.Fail(log.E("dev.tag", "git tag failed", r.Value.(error)))
	}
	return core.Ok(nil)
}

// pushWithTags pushes commits and tags to the remote.
// Uses interactive mode to support SSH passphrase prompts.
func pushWithTags(ctx context.Context, repoPath string) (_ core.Result) {
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
