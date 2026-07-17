// cmd_github.go implements the 'setup github' command for configuring
// GitHub repositories with organization standards.
//
// Usage:
//   core setup github [flags]
//
// Flags:
//   -r, --repo string    Specific repo to setup
//   -a, --all            Setup all repos in registry
//   -l, --labels         Only sync labels
//   -w, --webhooks       Only sync webhooks
//   -p, --protection     Only sync branch protection
//   -s, --security       Only sync security settings
//   -c, --check          Dry-run: show what would change
//       --config string  Path to github.yaml config
//       --verbose        Show detailed output

package setup

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
	"dappco.re/go/i18n"
	coreio "dappco.re/go/io"
	log "dappco.re/go/log"
	coreexec "dappco.re/go/process/exec"
	"dappco.re/go/scm/repos"
)

// GitHub command flags
var (
	ghRepo       string
	ghAll        bool
	ghLabels     bool
	ghWebhooks   bool
	ghProtection bool
	ghSecurity   bool
	ghCheck      bool
	ghConfigPath string
	ghVerbose    bool
)

// addGitHubCommand adds the 'setup github' command, plus a 'setup gh' alias
// (core.Command has no native Aliases field — see cmd_vm.go's addVMStatusCommand
// for the same multi-path-registration pattern).
func addGitHubCommand(c *core.Core) core.Result {
	flags := core.NewOptions(
		core.Option{Key: "repo", Value: ""},
		core.Option{Key: "all", Value: false},
		core.Option{Key: "labels", Value: false},
		core.Option{Key: "webhooks", Value: false},
		core.Option{Key: "protection", Value: false},
		core.Option{Key: "security", Value: false},
		core.Option{Key: "check", Value: false},
		core.Option{Key: "config", Value: ""},
		core.Option{Key: "verbose", Value: false},
	)
	action := func(o core.Options) core.Result {
		ghRepo = o.String("repo")
		ghAll = o.Bool("all")
		ghLabels = o.Bool("labels")
		ghWebhooks = o.Bool("webhooks")
		ghProtection = o.Bool("protection")
		ghSecurity = o.Bool("security")
		ghCheck = o.Bool("check")
		ghConfigPath = o.String("config")
		ghVerbose = o.Bool("verbose")
		return runGitHubSetup()
	}
	if r := c.Command("setup/github", core.Command{
		Description: i18n.T("cmd.setup.github.short"),
		Flags:       flags,
		Action:      action,
	}); !r.OK {
		return r
	}
	return c.Command("setup/gh", core.Command{
		Description: i18n.T("cmd.setup.github.short"),
		Flags:       flags,
		Action:      action,
	})
}

func runGitHubSetup() (_ core.Result) {
	// Check gh is available
	if r := coreexec.Command(core.Background(), "gh", "--version").Run(); !r.OK {
		return core.Fail(log.E("setup.github", i18n.T("error.gh_not_found"), nil))
	}

	// Check gh is authenticated
	if !cli.GhAuthenticated() {
		return core.Fail(log.E("setup.github", i18n.T("cmd.setup.github.error.not_authenticated"), nil))
	}

	// Find registry
	registryPath, err := repos.FindRegistry(coreio.Local)
	if err != nil {
		return core.Fail(cli.Wrap(err, i18n.T("error.registry_not_found")))
	}

	reg, err := repos.LoadRegistry(coreio.Local, registryPath)
	if err != nil {
		return core.Fail(cli.Wrap(err, "failed to load registry"))
	}

	registryDir := core.PathDir(registryPath)

	// Find GitHub config
	configPath, r := FindGitHubConfig(registryDir, ghConfigPath)
	if !r.OK {
		return core.Fail(cli.Wrap(r.Value.(error), i18n.T("cmd.setup.github.error.config_not_found")))
	}

	config, r := LoadGitHubConfig(configPath)
	if !r.OK {
		return core.Fail(cli.Wrap(r.Value.(error), "failed to load GitHub config"))
	}

	if r := config.Validate(); !r.OK {
		return core.Fail(cli.Wrap(r.Value.(error), "invalid GitHub config"))
	}

	// Print header
	cli.Print("%s %s\n", dimStyle.Render(i18n.Label("registry")), registryPath)
	cli.Print("%s %s\n", dimStyle.Render(i18n.Label("config")), configPath)

	if ghCheck {
		cli.Print("%s\n", warningStyle.Render(i18n.T("cmd.setup.github.dry_run_mode")))
	}

	// Determine which repos to process
	var reposToProcess []*repos.Repo

	// Reject conflicting flags
	if ghRepo != "" && ghAll {
		return core.Fail(log.E("setup.github", i18n.T("cmd.setup.github.error.conflicting_flags"), nil))
	}

	if ghRepo != "" {
		// Single repo mode
		repo, ok := reg.Get(ghRepo)
		if !ok {
			return core.Fail(log.E("setup.github", i18n.T("error.repo_not_found", map[string]any{"Name": ghRepo}), nil))
		}
		reposToProcess = []*repos.Repo{repo}
	} else if ghAll {
		// All repos mode
		reposToProcess = reg.List()
	} else {
		// No repos specified
		cli.Print("\n%s\n", i18n.T("cmd.setup.github.no_repos_specified"))
		cli.Print("  %s\n", i18n.T("cmd.setup.github.usage_hint"))
		return core.Ok(nil)
	}

	// Determine which operations to run
	runAll := !ghLabels && !ghWebhooks && !ghProtection && !ghSecurity
	runLabels := runAll || ghLabels
	runWebhooks := runAll || ghWebhooks
	runProtection := runAll || ghProtection
	runSecurity := runAll || ghSecurity

	// Process each repo
	aggregate := NewAggregate()

	for i, repo := range reposToProcess {
		repoFullName := cli.Sprintf("%s/%s", reg.Org, repo.Name)

		// Show progress
		cli.Print("\033[2K\r%s %d/%d %s",
			dimStyle.Render(i18n.T("common.progress.checking")),
			i+1, len(reposToProcess), repo.Name)

		changes := NewChangeSet(repo.Name)

		// Sync labels
		if runLabels {
			labelChanges, r := SyncLabels(repoFullName, config, ghCheck)
			if !r.OK {
				cli.Print("\033[2K\r")
				cli.Print("%s %s: %s\n", errorStyle.Render(cli.Glyph(":cross:")), repo.Name, r.Error())
				aggregate.Add(changes) // Preserve partial results
				continue
			}
			changes.Changes = append(changes.Changes, labelChanges.Changes...)
		}

		// Sync webhooks
		if runWebhooks {
			webhookChanges, r := SyncWebhooks(repoFullName, config, ghCheck)
			if !r.OK {
				cli.Print("\033[2K\r")
				cli.Print("%s %s: %s\n", errorStyle.Render(cli.Glyph(":cross:")), repo.Name, r.Error())
				aggregate.Add(changes) // Preserve partial results
				continue
			}
			changes.Changes = append(changes.Changes, webhookChanges.Changes...)
		}

		// Sync branch protection
		if runProtection {
			protectionChanges, r := SyncBranchProtection(repoFullName, config, ghCheck)
			if !r.OK {
				cli.Print("\033[2K\r")
				cli.Print("%s %s: %s\n", errorStyle.Render(cli.Glyph(":cross:")), repo.Name, r.Error())
				aggregate.Add(changes) // Preserve partial results
				continue
			}
			changes.Changes = append(changes.Changes, protectionChanges.Changes...)
		}

		// Sync security settings
		if runSecurity {
			securityChanges, r := SyncSecuritySettings(repoFullName, config, ghCheck)
			if !r.OK {
				cli.Print("\033[2K\r")
				cli.Print("%s %s: %s\n", errorStyle.Render(cli.Glyph(":cross:")), repo.Name, r.Error())
				aggregate.Add(changes) // Preserve partial results
				continue
			}
			changes.Changes = append(changes.Changes, securityChanges.Changes...)
		}

		aggregate.Add(changes)
	}

	// Clear progress line
	cli.Print("\033[2K\r")

	// Print results
	for _, cs := range aggregate.Sets {
		cs.Print(ghVerbose || ghCheck)
	}

	// Print summary
	aggregate.PrintSummary()

	// Suggest permission fix if needed
	if ghCheck {
		cli.Print("\n%s\n", i18n.T("cmd.setup.github.run_without_check"))
	}

	return core.Ok(nil)
}
