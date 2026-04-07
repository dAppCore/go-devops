package dev

import (
	"context"
	"os"
	"os/exec"

	"dappco.re/go/agent/pkg/lib"
	"dappco.re/go/core"
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
	return core.Result{}
}

// doCommit shells out to claude for AI-assisted commit.
func doCommit(ctx context.Context, repoPath string, allowEdit bool) error {
	prompt := ""
	if r := lib.Prompt("commit"); r.OK {
		prompt = r.Value.(string)
	}

	tools := "Bash,Read,Glob,Grep"
	if allowEdit {
		tools = "Bash,Read,Write,Edit,Glob,Grep"
	}

	cmd := exec.CommandContext(ctx, "claude", "-p", prompt, "--allowedTools", tools)
	cmd.Dir = repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
