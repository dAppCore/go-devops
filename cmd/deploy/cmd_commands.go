package deploy

import (
	"dappco.re/go/core/cli/pkg/cli"

	_ "dappco.re/go/core/devops/locales"
)

func init() {
	cli.RegisterCommands(AddDeployCommands)
}

// AddDeployCommands registers the 'deploy' command and all subcommands.
func AddDeployCommands(root *cli.Command) {
	setDeployI18n()
	root.AddCommand(Cmd)
}
