// cmd_registry.go implements registry mode for cloning packages.
//
// Registry mode is activated when a repos.yaml exists. It reads the registry
// and clones all (or selected) packages into the configured packages directory.

package setup

import (
	"context"

	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
	"dappco.re/go/devops/cmd/workspace"
	"dappco.re/go/i18n"
	coreio "dappco.re/go/io"
	log "dappco.re/go/log"
	coreexec "dappco.re/go/process/exec"
	"dappco.re/go/scm/repos"
)

// runRegistrySetup loads a registry from path and runs setup.
func runRegistrySetup(ctx context.Context, registryPath, only string, dryRun, all, runBuild bool) (_ coreFailure) {
	reg, err := repos.LoadRegistry(coreio.Local, registryPath)
	if err != nil {
		return log.E("setup.registry", "failed to load registry", err)
	}

	// Check workspace config for default_only if no filter specified
	if only == "" {
		registryDir := core.PathDir(registryPath)
		if wsConfig, err := workspace.LoadConfig(registryDir); err == nil && wsConfig != nil && len(wsConfig.DefaultOnly) > 0 {
			only = core.Join(",", wsConfig.DefaultOnly...)
		}
	}

	return runRegistrySetupWithReg(ctx, reg, registryPath, only, dryRun, all, runBuild)
}

// runRegistrySetupWithReg runs setup with an already-loaded registry.
func runRegistrySetupWithReg(ctx context.Context, reg *repos.Registry, registryPath, only string, dryRun, all, runBuild bool) (_ coreFailure) {
	cli.Print("%s %s\n", dimStyle.Render(i18n.Label("registry")), registryPath)
	cli.Print("%s %s\n", dimStyle.Render(i18n.T("cmd.setup.org_label")), reg.Org)

	registryDir := core.PathDir(registryPath)

	// Determine base path for cloning
	basePath := reg.BasePath
	if basePath == "" {
		// Load workspace config to see if packages_dir is set (ignore errors, fall back to default)
		wsConfig, _ := workspace.LoadConfig(registryDir)
		if wsConfig != nil && wsConfig.PackagesDir != "" {
			basePath = wsConfig.PackagesDir
		} else {
			basePath = "./packages"
		}
	}

	// Expand ~
	if core.HasPrefix(basePath, "~/") {
		homeResult := core.UserHomeDir()
		if homeResult.OK {
			basePath = core.PathJoin(homeResult.Value.(string), basePath[2:])
		}
	}

	// Resolve relative to registry location
	if !core.PathIsAbs(basePath) {
		basePath = core.PathJoin(registryDir, basePath)
	}

	cli.Print("%s %s\n", dimStyle.Render(i18n.Label("target")), basePath)

	// Parse type filter
	var typeFilter []string
	if only != "" {
		for _, t := range core.Split(only, ",") {
			typeFilter = append(typeFilter, core.Trim(t))
		}
		cli.Print("%s %s\n", dimStyle.Render(i18n.Label("filter")), only)
	}

	// Ensure base path exists
	if !dryRun {
		if err := coreio.Local.EnsureDir(basePath); err != nil {
			return log.E("setup.registry", "failed to create packages directory", err)
		}
	}

	// Get all available repos
	allRepos := reg.List()

	// Determine which repos to clone
	var toClone []*repos.Repo
	var skipped, exists int

	// Use wizard in interactive mode, unless --all specified
	useWizard := isTerminal() && !all && !dryRun

	if useWizard {
		selected, err := runPackageWizard(reg, typeFilter)
		if err != nil {
			return log.E("setup.registry", "wizard error", err)
		}

		// Build set of selected repos
		selectedSet := make(map[string]bool)
		for _, name := range selected {
			selectedSet[name] = true
		}

		// Filter repos based on selection
		for _, repo := range allRepos {
			if !selectedSet[repo.Name] {
				skipped++
				continue
			}

			// Check if already exists
			repoPath := core.PathJoin(basePath, repo.Name)
			// Check .git dir existence via Exists
			if coreio.Local.Exists(core.PathJoin(repoPath, ".git")) {
				exists++
				continue
			}

			toClone = append(toClone, repo)
		}
	} else {
		// Non-interactive: filter by type
		typeFilterSet := make(map[string]bool)
		for _, t := range typeFilter {
			typeFilterSet[t] = true
		}

		for _, repo := range allRepos {
			// Skip if type filter doesn't match (when filter is specified)
			if len(typeFilterSet) > 0 && !typeFilterSet[repo.Type] {
				skipped++
				continue
			}

			// Skip if clone: false
			if repo.Clone != nil && !*repo.Clone {
				skipped++
				continue
			}

			// Check if already exists
			repoPath := core.PathJoin(basePath, repo.Name)
			if coreio.Local.Exists(core.PathJoin(repoPath, ".git")) {
				exists++
				continue
			}

			toClone = append(toClone, repo)
		}
	}

	// Summary
	cli.Blank()
	cli.Print("%s, %s, %s\n",
		i18n.T("cmd.setup.to_clone", map[string]any{"Count": len(toClone)}),
		i18n.T("cmd.setup.exist", map[string]any{"Count": exists}),
		i18n.T("common.count.skipped", map[string]any{"Count": skipped}))

	if len(toClone) == 0 {
		cli.Print("\n%s\n", i18n.T("cmd.setup.nothing_to_clone"))
		return nil
	}

	if dryRun {
		cli.Print("\n%s\n", i18n.T("cmd.setup.would_clone_list"))
		for _, repo := range toClone {
			cli.Print("  %s (%s)\n", repoNameStyle.Render(repo.Name), repo.Type)
		}
		return nil
	}

	// Confirm in interactive mode
	if useWizard {
		confirmed, err := confirmClone(len(toClone), basePath)
		if err != nil {
			return err
		}
		if !confirmed {
			cli.Text(i18n.T("cmd.setup.cancelled"))
			return nil
		}
	}

	// Clone repos
	cli.Blank()
	var succeeded, failed int

	for _, repo := range toClone {
		cli.Print("  %s %s... ", dimStyle.Render(i18n.T("common.status.cloning")), repo.Name)

		repoPath := core.PathJoin(basePath, repo.Name)

		err := gitClone(ctx, reg.Org, repo.Name, repoPath)
		if err != nil {
			cli.Print("%s\n", errorStyle.Render("x "+err.Error()))
			failed++
		} else {
			cli.Print("%s\n", successStyle.Render(i18n.T("cmd.setup.done")))
			succeeded++
		}
	}

	// Summary
	cli.Blank()
	cli.Print("%s %s", successStyle.Render(i18n.Label("done")), i18n.T("cmd.setup.cloned_count", map[string]any{"Count": succeeded}))
	if failed > 0 {
		cli.Print(", %s", errorStyle.Render(i18n.T("i18n.count.failed", failed)))
	}
	if exists > 0 {
		cli.Print(", %s", i18n.T("cmd.setup.already_exist_count", map[string]any{"Count": exists}))
	}
	cli.Blank()

	// Run build if requested
	if runBuild && succeeded > 0 {
		cli.Blank()
		cli.Print("%s %s\n", dimStyle.Render(">>"), i18n.ProgressSubject("run", "build"))
		buildCmd := coreexec.Command(ctx, "core", "build").
			WithDir(basePath).
			WithStdout(core.Stdout()).
			WithStderr(core.Stderr())
		if err := commandResultError(buildCmd.Run()); err != nil {
			return log.E("setup.registry", i18n.T("i18n.fail.run", "build"), err)
		}
	}

	return nil
}

// gitClone clones a repository using gh CLI or git.
func gitClone(ctx context.Context, org, repo, path string) (_ coreFailure) {
	// Try gh clone first with HTTPS (works without SSH keys)
	if cli.GhAuthenticated() {
		// Use HTTPS URL directly to bypass git_protocol config
		httpsURL := core.Sprintf("https://github.com/%s/%s.git", org, repo)
		cmd := coreexec.Command(ctx, "gh", "repo", "clone", httpsURL, path)
		output, err := cmd.CombinedOutput()
		if err == nil {
			return nil
		}
		errStr := core.Trim(string(output))
		// Only fall through to SSH if it's an auth error
		if !core.Contains(errStr, "Permission denied") &&
			!core.Contains(errStr, "could not read") {
			return log.E("setup.registry", errStr, nil)
		}
	}

	// Fallback to git clone via SSH
	url := core.Sprintf("git@github.com:%s/%s.git", org, repo)
	cmd := coreexec.Command(ctx, "git", "clone", url, path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return log.E("setup.registry", core.Trim(string(output)), nil)
	}
	return nil
}
