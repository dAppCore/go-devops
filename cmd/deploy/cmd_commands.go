package deploy

import (
	"forge.lthn.ai/core/cli/pkg/cli"

	_ "forge.lthn.ai/core/go-devops/locales"
)

func init() {
	cli.RegisterCommands(AddDeployCommands)
}

// AddDeployCommands registers the 'deploy' command and all subcommands.
func AddDeployCommands(root *cli.Command) {
	setDeployI18n()
	root.AddCommand(Cmd)
}
