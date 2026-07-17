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
	cli.WithAppName("devops")
	cli.Main(
		cli.WithCommands("deploy", deploycmd.AddDeployCommands),
		cli.WithCommands("dev", devcmd.AddDevCommands),
		cli.WithCommands("docs", docscmd.AddDocsCommands),
		cli.WithCommands("git", gitcmd.AddGitCommands),
		cli.WithCommands("setup", setupcmd.AddSetupCommands),
		addPlaybookSmokeCommand,
	)
}

// addPlaybookSmokeCommand adds the 'playbook-smoke [dir]' command — not part
// of any cmd/ package, previously a bespoke cobra command wired directly
// into this test driver's root group.
func addPlaybookSmokeCommand(c *core.Core) core.Result {
	return c.Command("playbook-smoke", core.Command{
		Description: "Validate bundled playbook YAML can be decoded",
		Action: func(o core.Options) core.Result {
			return runPlaybookSmoke(o.String("_arg"))
		},
	})
}

func runPlaybookSmoke(dirArg string) (_ core.Result) {
	dir := "playbooks"
	if dirArg != "" {
		dir = dirArg
	}

	count := 0
	walkResult := core.PathWalkDir(dir, func(path string, entry core.FsDirEntry, err error) error {
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
	if !walkResult.OK {
		return core.Fail(core.Errorf("walk %s: %w", dir, walkResult.Value.(error)))
	}
	if count == 0 {
		return core.Fail(core.Errorf("no playbook YAML files found in %s", dir))
	}

	if result := core.WriteString(core.Stdout(), core.Sprintf("playbook smoke passed: %d YAML files decoded\n", count)); !result.OK {
		return result
	}
	return core.Ok(nil)
}

func isYAML(path string) bool {
	ext := core.Lower(core.PathExt(path))
	return ext == ".yaml" || ext == ".yml"
}
