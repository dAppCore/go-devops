// AX-10 CLI driver for go-devops.
//
//	task -d tests/cli/devops
//	go run ./tests/cli/devops dev --help
package main

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
	deploycmd "dappco.re/go/devops/cmd/deploy"
	devcmd "dappco.re/go/devops/cmd/dev"
	docscmd "dappco.re/go/devops/cmd/docs"
	gitcmd "dappco.re/go/devops/cmd/gitcmd"
	setupcmd "dappco.re/go/devops/cmd/setup"

	"gopkg.in/yaml.v3"
)

func main() {
	root := cli.NewGroup("devops", "DevOps CLI artifact test driver", "")
	devcmd.AddDevCommands(root)
	deploycmd.AddDeployCommands(root)
	docscmd.AddDocsCommands(root)
	gitcmd.AddGitCommands(root)
	setupcmd.AddSetupCommands(root)
	root.AddCommand(playbookSmokeCommand())
	root.SetArgs(core.Args()[1:])

	if err := root.Execute(); err != nil {
		core.Print(core.Stderr(), "%v", err)
		core.Exit(1)
	}
}

func playbookSmokeCommand() *cli.Command {
	return &cli.Command{
		Use:   "playbook-smoke [dir]",
		Short: "Validate bundled playbook YAML can be decoded",
		Args:  cli.RangeArgs(0, 1),
		RunE:  runPlaybookSmoke,
	}
}

func runPlaybookSmoke(cmd *cli.Command, args []string) (_ coreFailure) {
	dir := "playbooks"
	if len(args) > 0 {
		dir = args[0]
	}

	count := 0
	err := core.PathWalkDir(dir, func(path string, entry core.FsDirEntry, err error) error {
		if err != nil {
			return core.Errorf("%s: %w", path, err)
		}
		if entry.IsDir() || !isYAML(path) {
			return nil
		}

		rawResult := core.ReadFile(path)
		if !rawResult.OK {
			return core.Errorf("%s: %w", path, rawResult.Value.(error))
		}
		raw := rawResult.Value.([]byte)

		var document any
		if err := yaml.Unmarshal(raw, &document); err != nil {
			return core.Errorf("%s: %w", path, err)
		}
		count++
		return nil
	})
	if err != nil {
		return core.Errorf("walk %s: %w", dir, err)
	}
	if count == 0 {
		return core.Errorf("no playbook YAML files found in %s", dir)
	}

	if result := core.WriteString(cmd.OutOrStdout(), core.Sprintf("playbook smoke passed: %d YAML files decoded\n", count)); !result.OK {
		return result.Value.(error)
	}
	return nil
}

func isYAML(path string) bool {
	ext := core.Lower(core.PathExt(path))
	return ext == ".yaml" || ext == ".yml"
}
