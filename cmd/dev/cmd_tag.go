package dev

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/go-i18n"
	log "forge.lthn.ai/core/go-log"
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

func runTag(registryPath string, dryRun, force bool) error {
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
			return log.E("dev.tag", fmt.Sprintf("%s: failed to bump version %s", repo.Name, current), err)
		}

		hasGoMod := fileExists(filepath.Join(repo.Path, "go.mod"))

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
		if !cli.Confirm(fmt.Sprintf("Tag and push %d repos?", len(plans))) {
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
	cli.Print("%s", successStyle.Render(fmt.Sprintf("%d tagged", succeeded)))
	if failed > 0 {
		cli.Print(", %s", errorStyle.Render(i18n.T("common.count.failed", map[string]any{"Count": failed})))
	}
	cli.Blank()

	return nil
}

// latestTag returns the latest semver tag in the repo.
func latestTag(ctx context.Context, repoPath string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "describe", "--tags", "--abbrev=0", "--match", "v*")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// bumpPatch increments the patch version of a semver tag.
// "v0.3.1" → "v0.3.2"
func bumpPatch(tag string) (string, error) {
	v := strings.TrimPrefix(tag, "v")
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return "", log.E("dev.tag", fmt.Sprintf("invalid semver: %s", tag), nil)
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return "", log.E("dev.tag", fmt.Sprintf("invalid patch version: %s", parts[2]), nil)
	}
	return fmt.Sprintf("v%s.%s.%d", parts[0], parts[1], patch+1), nil
}

// goGetUpdate runs GOWORK=off go get -u ./... in the repo.
func goGetUpdate(ctx context.Context, repoPath string) error {
	cmd := exec.CommandContext(ctx, "go", "get", "-u", "./...")
	cmd.Dir = repoPath
	cmd.Env = append(os.Environ(), "GOWORK=off")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return log.E("dev.tag", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// goModTidy runs GOWORK=off go mod tidy in the repo.
func goModTidy(ctx context.Context, repoPath string) error {
	cmd := exec.CommandContext(ctx, "go", "mod", "tidy")
	cmd.Dir = repoPath
	cmd.Env = append(os.Environ(), "GOWORK=off")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return log.E("dev.tag", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// commitGoMod stages and commits go.mod and go.sum if they have changes.
func commitGoMod(ctx context.Context, repoPath, version string) error {
	// Check if go.mod or go.sum changed (staged or unstaged)
	diffCmd := exec.CommandContext(ctx, "git", "diff", "--quiet", "go.mod", "go.sum")
	diffCmd.Dir = repoPath
	modChanged := diffCmd.Run() != nil

	// Also check for untracked go.sum
	lsCmd := exec.CommandContext(ctx, "git", "ls-files", "--others", "--exclude-standard", "go.sum")
	lsCmd.Dir = repoPath
	lsOut, _ := lsCmd.Output()
	untrackedSum := strings.TrimSpace(string(lsOut)) != ""

	if !modChanged && !untrackedSum {
		return nil // No changes
	}

	// Stage go.mod and go.sum
	addCmd := exec.CommandContext(ctx, "git", "add", "go.mod", "go.sum")
	addCmd.Dir = repoPath
	if out, err := addCmd.CombinedOutput(); err != nil {
		return log.E("dev.tag", "git add: "+strings.TrimSpace(string(out)), err)
	}

	// Check if anything is actually staged
	stagedCmd := exec.CommandContext(ctx, "git", "diff", "--cached", "--quiet")
	stagedCmd.Dir = repoPath
	if stagedCmd.Run() == nil {
		return nil // Nothing staged
	}

	// Commit
	msg := fmt.Sprintf("chore: sync dependencies for %s\n\nCo-Authored-By: Virgil <virgil@lethean.io>", version)
	commitCmd := exec.CommandContext(ctx, "git", "commit", "-m", msg)
	commitCmd.Dir = repoPath
	if out, err := commitCmd.CombinedOutput(); err != nil {
		return log.E("dev.tag", "git commit: "+strings.TrimSpace(string(out)), err)
	}
	return nil
}

// createTag creates an annotated tag.
func createTag(ctx context.Context, repoPath, tag string) error {
	cmd := exec.CommandContext(ctx, "git", "tag", "-a", tag, "-m", tag)
	cmd.Dir = repoPath
	if out, err := cmd.CombinedOutput(); err != nil {
		return log.E("dev.tag", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// pushWithTags pushes commits and tags to the remote.
// Uses interactive mode to support SSH passphrase prompts.
func pushWithTags(ctx context.Context, repoPath string) error {
	cmd := exec.CommandContext(ctx, "git", "push", "--follow-tags")
	cmd.Dir = repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// fileExists checks if a file exists at the given path.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
