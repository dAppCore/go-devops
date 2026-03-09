package deploy

import (
	"forge.lthn.ai/core/cli/pkg/cli"
)

func init() {
	cli.RegisterCommands(AddDeployCommands)
}

// AddDeployCommands registers the 'deploy' command and all subcommands.
func AddDeployCommands(root *cli.Command) {
	root.AddCommand(Cmd)
}
