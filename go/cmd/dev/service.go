package dev

import (
	"context"

	core "dappco.re/go"
	"dappco.re/go/agent/pkg/lib"
	coreexec "dappco.re/go/process/exec"
)

// ServiceOptions for configuring the dev service.
//
//	dev.ServiceOptions{
//	    RegistryPath: "/etc/core/dev/registry.yaml",
//	}
type ServiceOptions struct {
	// RegistryPath is the on-disk path for the dev workflow repository
	// registry. Empty → caller-default resolution applies.
	RegistryPath string
}

// Service provides dev workflow orchestration as a Core service.
//
// Usage example: `svc := core.MustServiceFor[*dev.Service](c, "dev")`
type Service struct {
	*core.ServiceRuntime[ServiceOptions]
}

// NewService returns a factory that constructs a *Service from the
// supplied options and registers it under "dev" via core.WithService.
//
//	core.WithService(dev.NewService(dev.ServiceOptions{
//	    RegistryPath: "/etc/core/dev/registry.yaml",
//	}))
func NewService(opts ServiceOptions) func(*core.Core) core.Result {
	return func(c *core.Core) core.Result {
		return core.Ok(&Service{
			ServiceRuntime: core.NewServiceRuntime(c, opts),
		})
	}
}

// Register wires the dev service into the Core with empty
// ServiceOptions — the imperative-style alternative to NewService.
//
//	core.New(core.WithService(dev.Register))
func Register(c *core.Core) core.Result {
	return NewService(ServiceOptions{})(c)
}

func (s *Service) handleAction(_ *core.Core, _ core.Message) core.Result {
	return core.Ok(nil)
}

// doCommit shells out to claude for AI-assisted commit.
func doCommit(ctx context.Context, repoPath string, allowEdit bool) (_ core.Result) {
	prompt := ""
	if r := lib.Prompt("commit"); r.OK {
		value, ok := r.Value.(string)
		if !ok {
			return core.Fail(core.E("dev.commit", "commit prompt was not a string", nil))
		}
		prompt = value
	}

	tools := "Bash,Read,Glob,Grep"
	if allowEdit {
		tools = "Bash,Read,Write,Edit,Glob,Grep"
	}

	cmd := coreexec.Command(ctx, "claude", "-p", prompt, "--allowedTools", tools).
		WithDir(repoPath).
		WithStdout(core.Stdout()).
		WithStderr(core.Stderr()).
		WithStdin(core.Stdin())

	return resultError(cmd.Run())
}
