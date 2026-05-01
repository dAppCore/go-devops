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

// addGitHubCommand adds the 'github' subcommand to the setup command.
func addGitHubCommand(parent *cli.Command) {
	ghCmd := &cli.Command{
		Use:     "github",
		Aliases: []string{"gh"},
		Short:   i18n.T("cmd.setup.github.short"),
		Long:    i18n.T("cmd.setup.github.long"),
		RunE: func(cmd *cli.Command, args []string) error {
			return resultError(runGitHubSetup())
		},
	}

	ghCmd.Flags().StringVarP(&ghRepo, "repo", "r", "", i18n.T("cmd.setup.github.flag.repo"))
	ghCmd.Flags().BoolVarP(&ghAll, "all", "a", false, i18n.T("cmd.setup.github.flag.all"))
	ghCmd.Flags().BoolVarP(&ghLabels, "labels", "l", false, i18n.T("cmd.setup.github.flag.labels"))
	ghCmd.Flags().BoolVarP(&ghWebhooks, "webhooks", "w", false, i18n.T("cmd.setup.github.flag.webhooks"))
	ghCmd.Flags().BoolVarP(&ghProtection, "protection", "p", false, i18n.T("cmd.setup.github.flag.protection"))
	ghCmd.Flags().BoolVarP(&ghSecurity, "security", "s", false, i18n.T("cmd.setup.github.flag.security"))
	ghCmd.Flags().BoolVarP(&ghCheck, "check", "c", false, i18n.T("cmd.setup.github.flag.check"))
	ghCmd.Flags().StringVar(&ghConfigPath, "config", "", i18n.T("cmd.setup.github.flag.config"))
	ghCmd.Flags().BoolVarP(&ghVerbose, "verbose", "v", false, i18n.T("common.flag.verbose"))

	parent.AddCommand(ghCmd)
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
