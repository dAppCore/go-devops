// AX-10 CLI driver for go-devops.
//
//	task -d tests/cli/devops
//	go run ./tests/cli/devops dev --help
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	root.SetArgs(os.Args[1:])

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
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

func runPlaybookSmoke(cmd *cli.Command, args []string) error {
	dir := "playbooks"
	if len(args) > 0 {
		dir = args[0]
	}

	count := 0
	err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if entry.IsDir() || !isYAML(path) {
			return nil
		}

		raw, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}

		var document any
		if err := yaml.Unmarshal(raw, &document); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		count++
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk %s: %w", dir, err)
	}
	if count == 0 {
		return fmt.Errorf("no playbook YAML files found in %s", dir)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "playbook smoke passed: %d YAML files decoded\n", count)
	return nil
}

func isYAML(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yaml" || ext == ".yml"
}
