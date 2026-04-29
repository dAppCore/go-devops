// cmd_file_sync.go implements safe file synchronization across repos for AI agents.
//
// Usage:
//   core dev sync workflow.yml --to="packages/core-*"
//   core dev sync .github/workflows/ --to="packages/core-*" --message="feat: add CI"
//   core dev sync config.yaml --to="packages/core-*" --dry-run

package dev

import (
	"context"

	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
	"dappco.re/go/i18n"
	coreio "dappco.re/go/io"
	"dappco.re/go/log"
	coreexec "dappco.re/go/process/exec"
	"dappco.re/go/scm/git"
	"dappco.re/go/scm/repos"
)

// File sync command flags
var (
	fileSyncTo       string
	fileSyncMessage  string
	fileSyncCoAuthor string
	fileSyncDryRun   bool
	fileSyncPush     bool
	fileSyncYes      bool
)

// AddFileSyncCommand adds the 'sync' command to dev for file syncing.
func AddFileSyncCommand(parent *cli.Command) {
	syncCmd := &cli.Command{
		Use:   "sync <file-or-dir>",
		Short: i18n.T("cmd.dev.file_sync.short"),
		Long:  i18n.T("cmd.dev.file_sync.long"),
		Args:  cli.MinimumNArgs(1),
		RunE: func(cmd *cli.Command, args []string) error {
			return runFileSync(args[0])
		},
	}

	syncCmd.Flags().StringVar(&fileSyncTo, "to", "", i18n.T("cmd.dev.file_sync.flag.to"))
	syncCmd.Flags().StringVarP(&fileSyncMessage, "message", "m", "", i18n.T("cmd.dev.file_sync.flag.message"))
	syncCmd.Flags().StringVar(&fileSyncCoAuthor, "co-author", "", i18n.T("cmd.dev.file_sync.flag.co_author"))
	syncCmd.Flags().BoolVar(&fileSyncDryRun, "dry-run", false, i18n.T("cmd.dev.file_sync.flag.dry_run"))
	syncCmd.Flags().BoolVar(&fileSyncPush, "push", false, i18n.T("cmd.dev.file_sync.flag.push"))
	syncCmd.Flags().BoolVarP(&fileSyncYes, "yes", "y", false, i18n.T("cmd.dev.file_sync.flag.yes"))

	if err := syncCmd.MarkFlagRequired("to"); err != nil {
		panic(err)
	}

	parent.AddCommand(syncCmd)
}

func runFileSync(source string) (_ coreFailure) {
	ctx := context.Background()

	// Security: Reject path traversal attempts
	if core.Contains(source, "..") {
		return log.E("dev.sync", "path traversal not allowed", nil)
	}

	// Validate source exists
	sourceInfoResult := core.Stat(source)
	if !sourceInfoResult.OK {
		return log.E("dev.sync", i18n.T("cmd.dev.file_sync.error.source_not_found", map[string]any{"Path": source}), sourceInfoResult.Value.(error))
	}
	sourceInfo := sourceInfoResult.Value.(core.FsFileInfo)

	// Find target repos
	targetRepos, err := resolveTargetRepos(fileSyncTo)
	if err != nil {
		return err
	}

	if len(targetRepos) == 0 {
		return cli.Err("%s", i18n.T("cmd.dev.file_sync.error.no_targets"))
	}

	// Show plan
	cli.Print("%s: %s\n", dimStyle.Render(i18n.T("cmd.dev.file_sync.source")), source)
	cli.Print("%s: %d repos\n", dimStyle.Render(i18n.T("cmd.dev.file_sync.targets")), len(targetRepos))
	if fileSyncDryRun {
		cli.Print("%s\n", warningStyle.Render(i18n.T("cmd.dev.file_sync.dry_run_mode")))
	}
	cli.Blank()

	if !fileSyncDryRun && !fileSyncYes {
		cli.Print("%s\n", warningStyle.Render(i18n.T("cmd.dev.file_sync.warning")))
		cli.Blank()
		if !cli.Confirm(i18n.T("cmd.dev.file_sync.confirm")) {
			cli.Text(i18n.T("cli.aborted"))
			return nil
		}
		cli.Blank()
	}

	var succeeded, skipped, failed int

	for _, repo := range targetRepos {
		repoName := core.PathBase(repo.Path)

		if fileSyncDryRun {
			cli.Print("  %s %s\n", dimStyle.Render("[dry-run]"), repoName)
			succeeded++
			continue
		}

		// Step 1: Pull latest (safe sync)
		if err := safePull(ctx, repo.Path); err != nil {
			cli.Print("  %s %s: pull failed: %s\n", errorStyle.Render("x"), repoName, err)
			failed++
			continue
		}

		// Step 2: Copy file(s)
		destPath := core.PathJoin(repo.Path, source)
		if sourceInfo.IsDir() {
			if err := copyDir(source, destPath); err != nil {
				cli.Print("  %s %s: copy failed: %s\n", errorStyle.Render("x"), repoName, err)
				failed++
				continue
			}
		} else {
			// Ensure dir exists
			if err := coreio.Local.EnsureDir(core.PathDir(destPath)); err != nil {
				cli.Print("  %s %s: copy failed: %s\n", errorStyle.Render("x"), repoName, err)
				failed++
				continue
			}
			if err := coreio.Copy(coreio.Local, source, coreio.Local, destPath); err != nil {
				cli.Print("  %s %s: copy failed: %s\n", errorStyle.Render("x"), repoName, err)
				failed++
				continue
			}
		}

		// Step 3: Check if anything changed
		statuses := git.Status(ctx, git.StatusOptions{
			Paths: []string{repo.Path},
			Names: map[string]string{repo.Path: repoName},
		})
		if len(statuses) == 0 || !statuses[0].IsDirty() {
			cli.Print("  %s %s: %s\n", dimStyle.Render("-"), repoName, i18n.T("cmd.dev.file_sync.no_changes"))
			skipped++
			continue
		}

		// Step 4: Commit if message provided
		if fileSyncMessage != "" {
			commitMsg := fileSyncMessage
			if fileSyncCoAuthor != "" {
				commitMsg += "\n\nCo-Authored-By: " + fileSyncCoAuthor
			}

			if err := gitAddCommit(ctx, repo.Path, source, commitMsg); err != nil {
				cli.Print("  %s %s: commit failed: %s\n", errorStyle.Render("x"), repoName, err)
				failed++
				continue
			}

			// Step 5: Push if requested
			if fileSyncPush {
				if err := safePush(ctx, repo.Path); err != nil {
					cli.Print("  %s %s: push failed: %s\n", errorStyle.Render("x"), repoName, err)
					failed++
					continue
				}
			}
		}

		cli.Print("  %s %s\n", successStyle.Render("v"), repoName)
		succeeded++
	}

	// Summary
	cli.Blank()
	cli.Print("%s: ", i18n.T("cmd.dev.file_sync.summary"))
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

// resolveTargetRepos resolves the --to pattern to actual repos
func resolveTargetRepos(pattern string) ([]*repos.Repo, coreFailure) {
	// Load registry
	registryPath, err := repos.FindRegistry(coreio.Local)
	if err != nil {
		return nil, log.E("dev.sync", "failed to find registry", err)
	}

	registry, err := repos.LoadRegistry(coreio.Local, registryPath)
	if err != nil {
		return nil, log.E("dev.sync", "failed to load registry", err)
	}

	// Match pattern against repo names
	var matched []*repos.Repo
	patterns := splitPatterns(pattern)
	for _, repo := range registry.Repos {
		for _, candidate := range patterns {
			if matchGlob(repo.Name, candidate) || matchGlob(repo.Path, candidate) {
				matched = append(matched, repo)
				break
			}
		}
	}

	return matched, nil
}

// splitPatterns normalises comma-separated glob patterns.
func splitPatterns(pattern string) []string {
	raw := core.Split(pattern, ",")
	out := make([]string, 0, len(raw))

	for _, p := range raw {
		p = core.Trim(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}

	return out
}

// matchGlob performs simple glob matching with * wildcards
func matchGlob(s, pattern string) bool {
	// Handle exact match and simple glob patterns.
	if s == pattern {
		return true
	}

	matchResult := core.PathMatch(pattern, s)
	if matchResult.OK {
		return matchResult.Value.(bool)
	}

	// Fallback to legacy wildcard rules for invalid glob patterns.
	// Handle * at end
	if core.HasSuffix(pattern, "*") {
		prefix := core.TrimSuffix(pattern, "*")
		return core.HasPrefix(s, prefix)
	}

	// Handle * at start
	if core.HasPrefix(pattern, "*") {
		suffix := core.TrimPrefix(pattern, "*")
		return core.HasSuffix(s, suffix)
	}

	// Handle * in middle
	if core.Contains(pattern, "*") {
		parts := core.SplitN(pattern, "*", 2)
		return core.HasPrefix(s, parts[0]) && core.HasSuffix(s, parts[1])
	}

	return false
}

// safePull pulls with rebase, handling errors gracefully
func safePull(ctx context.Context, path string) (_ coreFailure) {
	// Check if we have upstream
	_, err := gitCommandQuiet(ctx, path, "rev-parse", "--abbrev-ref", "@{u}")
	if err != nil {
		// No upstream set, skip pull
		return nil
	}

	return git.Pull(ctx, path)
}

// safePush pushes with automatic pull-rebase on rejection
func safePush(ctx context.Context, path string) (_ coreFailure) {
	err := git.Push(ctx, path)
	if err == nil {
		return nil
	}

	// If non-fast-forward, try pull and push again
	if git.IsNonFastForward(err) {
		if pullErr := git.Pull(ctx, path); pullErr != nil {
			return pullErr
		}
		return git.Push(ctx, path)
	}

	return err
}

// gitAddCommit stages and commits a file/directory
func gitAddCommit(ctx context.Context, repoPath, filePath, message string) (_ coreFailure) {
	// Stage the file(s)
	if _, err := gitCommandQuiet(ctx, repoPath, "add", filePath); err != nil {
		return err
	}

	// Commit
	_, err := gitCommandQuiet(ctx, repoPath, "commit", "-m", message)
	return err
}

// gitCommandQuiet runs a git command without output
func gitCommandQuiet(ctx context.Context, dir string, args ...string) (string, coreFailure) {
	cmd := coreexec.Command(ctx, "git", args...).WithDir(dir)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", cli.Err("%s", core.Trim(string(output)))
	}
	return string(output), nil
}

// copyDir recursively copies a directory
func copyDir(src, dst string) (_ coreFailure) {
	entries, err := coreio.Local.List(src)
	if err != nil {
		return err
	}

	if err := coreio.Local.EnsureDir(dst); err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := core.PathJoin(src, entry.Name())
		dstPath := core.PathJoin(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := coreio.Copy(coreio.Local, srcPath, coreio.Local, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}
