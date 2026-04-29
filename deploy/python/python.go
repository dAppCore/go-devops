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
func Init() (_ coreFailure) {
	once.Do(func() {
		ep, initErr = newEmbeddedPython("core-deploy")
	})
	return initErr
}

// GetPython returns the embedded Python instance.
func GetPython() *python.EmbeddedPython {
	return ep
}

// RunScript runs a Python script with the given code and returns stdout.
func RunScript(ctx context.Context, code string, args ...string) (string, coreFailure) {
	if err := initRuntime(); err != nil {
		return "", err
	}

	// Write code to temp file
	tmpResult := core.MkdirTemp("", "core-python-*")
	if !tmpResult.OK {
		return "", log.E("python", "create temp dir", nil)
	}
	tmpDir := tmpResult.Value.(string)
	defer func() {
		if r := core.RemoveAll(tmpDir); !r.OK {
			log.Warn("failed to remove temporary Python script directory", "script_dir", tmpDir, "error", r.Error())
		}
	}()

	tmpPath := core.PathJoin(tmpDir, "script.py")
	if r := core.WriteFile(tmpPath, []byte(code), 0o600); !r.OK {
		return "", log.E("python", "write script", r.Value.(error))
	}

	// Build args: script path + any additional args
	cmdArgs := append([]string{tmpPath}, args...)

	// Get the command
	cmd, err := pythonCommand(cmdArgs...)
	if err != nil {
		return "", log.E("python", "create command", err)
	}

	// Run with context
	output, err := cmd.Output()
	if err != nil {
		return "", log.E("python", "run script", err)
	}

	return string(output), nil
}

// RunModule runs a Python module (python -m module_name).
func RunModule(ctx context.Context, module string, args ...string) (string, coreFailure) {
	if err := initRuntime(); err != nil {
		return "", err
	}

	cmdArgs := append([]string{"-m", module}, args...)
	cmd, err := pythonCommand(cmdArgs...)
	if err != nil {
		return "", log.E("python", "create command", err)
	}

	output, err := cmd.Output()
	if err != nil {
		return "", log.E("python", "run module "+module, err)
	}

	return string(output), nil
}

// DevOpsPath returns the path to the DevOps repo.
func DevOpsPath() (string, coreFailure) {
	if path := core.Getenv("DEVOPS_PATH"); path != "" {
		return path, nil
	}
	r := core.UserHomeDir()
	if !r.OK {
		return "", log.E("python", "get user home", r.Value.(error))
	}
	return core.PathJoin(r.Value.(string), "Code", "DevOps"), nil
}

// CoolifyModulePath returns the path to the Coolify module_utils.
func CoolifyModulePath() (string, coreFailure) {
	path, err := DevOpsPath()
	if err != nil {
		return "", err
	}
	return core.PathJoin(path, "playbooks", "roles", "coolify", "module_utils"), nil
}

// CoolifyScript generates Python code to call the Coolify API.
func CoolifyScript(baseURL, apiToken, operation string, params map[string]any) (string, coreFailure) {
	paramsResult := core.JSONMarshal(params)
	if !paramsResult.OK {
		return "", log.E("python", "marshal params", paramsResult.Value.(error))
	}

	modulePath, err := CoolifyModulePath()
	if err != nil {
		return "", err
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
`, modulePath, baseURL, apiToken, string(paramsResult.Value.([]byte)), operation), nil
}
