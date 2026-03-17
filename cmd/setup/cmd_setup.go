// Package setup provides workspace setup and bootstrap commands.
package setup

import (
	"forge.lthn.ai/core/cli/pkg/cli"
	"forge.lthn.ai/core/go-i18n"
)

// Style aliases from shared package
var (
	repoNameStyle = cli.RepoStyle
	successStyle  = cli.SuccessStyle
	errorStyle    = cli.ErrorStyle
	warningStyle  = cli.WarningStyle
	dimStyle      = cli.DimStyle
)

// Default organization and devops repo for bootstrap
const (
	defaultOrg      = "host-uk"
	devopsRepo      = "core-devops"
	devopsReposYaml = "repos.yaml"
)

// Setup command flags
var (
	registryPath string
	only         string
	dryRun       bool
	all          bool
	name         string
	build        bool
)

var setupCmd = &cli.Command{
	Use: "setup",
	RunE: func(cmd *cli.Command, args []string) error {
		return runSetupOrchestrator(registryPath, only, dryRun, all, name, build)
	},
}

func initSetupFlags() {
	setupCmd.Flags().StringVar(&registryPath, "registry", "", i18n.T("cmd.setup.flag.registry"))
	setupCmd.Flags().StringVar(&only, "only", "", i18n.T("cmd.setup.flag.only"))
	setupCmd.Flags().BoolVar(&dryRun, "dry-run", false, i18n.T("cmd.setup.flag.dry_run"))
	setupCmd.Flags().BoolVar(&all, "all", false, i18n.T("cmd.setup.flag.all"))
	setupCmd.Flags().StringVar(&name, "name", "", i18n.T("cmd.setup.flag.name"))
	setupCmd.Flags().BoolVar(&build, "build", false, i18n.T("cmd.setup.flag.build"))
}

// AddSetupCommand adds the 'setup' command to the given parent command.
func AddSetupCommand(root *cli.Command) {
	initSetupFlags()
	addGitHubCommand(setupCmd)
	root.AddCommand(setupCmd)
}
