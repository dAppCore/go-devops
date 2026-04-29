package dev

import (
	"context"

	core "dappco.re/go"
	"dappco.re/go/agent/pkg/lib"
	coreexec "dappco.re/go/process/exec"
)

// ServiceOptions for configuring the dev service.
type ServiceOptions struct {
	RegistryPath string
}

// Service provides dev workflow orchestration as a Core service.
type Service struct {
	*core.ServiceRuntime[ServiceOptions]
}

func (s *Service) handleAction(_ *core.Core, _ core.Message) core.Result {
	return core.Ok(nil)
}

// doCommit shells out to claude for AI-assisted commit.
func doCommit(ctx context.Context, repoPath string, allowEdit bool) (_ coreFailure) {
	prompt := ""
	if r := lib.Prompt("commit"); r.OK {
		value, ok := r.Value.(string)
		if !ok {
			return core.E("dev.commit", "commit prompt was not a string", nil)
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
