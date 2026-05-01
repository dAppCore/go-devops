package python

import (
	"context"
	"sync"

	core "dappco.re/go"
	"dappco.re/go/log"
	"github.com/kluctl/go-embed-python/python"
)

type pythonRunner interface {
	Output() ([]byte, error)
}

var (
	once    sync.Once
	ep      *python.EmbeddedPython
	initErr error

	newEmbeddedPython = python.NewEmbeddedPython
	initRuntime       = Init
	pythonCommand     = func(args ...string) (pythonRunner, error) {
		return ep.PythonCmd(args...)
	}
)

// Init initializes the embedded Python runtime.
func Init() (_ core.Result) {
	once.Do(func() {
		ep, initErr = newEmbeddedPython("core-deploy")
	})
	if initErr != nil {
		return core.Fail(initErr)
	}
	return core.Ok(ep)
}

// GetPython returns the embedded Python instance.
func GetPython() *python.EmbeddedPython {
	return ep
}

// RunScript runs a Python script with the given code and returns stdout.
func RunScript(ctx context.Context, code string, args ...string) (string, core.Result) {
	if r := initRuntime(); !r.OK {
		return "", r
	}

	// Write code to temp file
	tmpResult := core.MkdirTemp("", "core-python-*")
	if !tmpResult.OK {
		return "", core.Fail(log.E("python", "create temp dir", tmpResult.Value.(error)))
	}
	tmpDir := tmpResult.Value.(string)
	defer func() {
		if r := core.RemoveAll(tmpDir); !r.OK {
			log.Warn("failed to remove temporary Python script directory", "script_dir", tmpDir, "error", r.Error())
		}
	}()

	tmpPath := core.PathJoin(tmpDir, "script.py")
	if r := core.WriteFile(tmpPath, []byte(code), 0o600); !r.OK {
		return "", core.Fail(log.E("python", "write script", r.Value.(error)))
	}

	// Build args: script path + any additional args
	cmdArgs := append([]string{tmpPath}, args...)

	// Get the command
	cmd, err := pythonCommand(cmdArgs...)
	if err != nil {
		return "", core.Fail(log.E("python", "create command", err))
	}

	// Run with context
	output, err := cmd.Output()
	if err != nil {
		return "", core.Fail(log.E("python", "run script", err))
	}

	return string(output), core.Ok(nil)
}

// RunModule runs a Python module (python -m module_name).
func RunModule(ctx context.Context, module string, args ...string) (string, core.Result) {
	if r := initRuntime(); !r.OK {
		return "", r
	}

	cmdArgs := append([]string{"-m", module}, args...)
	cmd, err := pythonCommand(cmdArgs...)
	if err != nil {
		return "", core.Fail(log.E("python", "create command", err))
	}

	output, err := cmd.Output()
	if err != nil {
		return "", core.Fail(log.E("python", "run module "+module, err))
	}

	return string(output), core.Ok(nil)
}

// DevOpsPath returns the path to the DevOps repo.
func DevOpsPath() (string, core.Result) {
	if path := core.Getenv("DEVOPS_PATH"); path != "" {
		return path, core.Ok(nil)
	}
	r := core.UserHomeDir()
	if !r.OK {
		return "", core.Fail(log.E("python", "get user home", r.Value.(error)))
	}
	return core.PathJoin(r.Value.(string), "Code", "DevOps"), core.Ok(nil)
}

// CoolifyModulePath returns the path to the Coolify module_utils.
func CoolifyModulePath() (string, core.Result) {
	path, r := DevOpsPath()
	if !r.OK {
		return "", r
	}
	return core.PathJoin(path, "playbooks", "roles", "coolify", "module_utils"), core.Ok(nil)
}

// CoolifyScript generates Python code to call the Coolify API.
func CoolifyScript(baseURL, apiToken, operation string, params map[string]any) (string, core.Result) {
	paramsResult := core.JSONMarshal(params)
	if !paramsResult.OK {
		return "", core.Fail(log.E("python", "marshal params", paramsResult.Value.(error)))
	}

	modulePath, r := CoolifyModulePath()
	if !r.OK {
		return "", r
	}

	return core.Sprintf(`
import sys
import json
sys.path.insert(0, %q)

from swagger.coolify_api import CoolifyClient

client = CoolifyClient(
    base_url=%q,
    api_token=%q,
    timeout=30,
    verify_ssl=True,
)

params = json.loads(%q)
result = client._call(%q, params, check_response=False)
print(json.dumps(result))
`, modulePath, baseURL, apiToken, string(paramsResult.Value.([]byte)), operation), core.Ok(nil)
}
