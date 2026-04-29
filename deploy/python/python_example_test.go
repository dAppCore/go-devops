package python

import (
	"sync"

	. "dappco.re/go"
	embedpython "github.com/kluctl/go-embed-python/python"
)

type examplePythonRunner struct {
	out []byte
	err error
}

func (r examplePythonRunner) Output() ([]byte, error) { return r.out, r.err }

func examplePythonRuntime(out string) func() {
	oldInit := initRuntime
	oldCommand := pythonCommand
	initRuntime = func() error { return nil }
	pythonCommand = func(args ...string) (pythonRunner, error) { return examplePythonRunner{out: []byte(out)}, nil }
	return func() { initRuntime = oldInit; pythonCommand = oldCommand }
}

func ExampleInit() {
	oldEP, oldErr, oldNew := ep, initErr, newEmbeddedPython
	once = sync.Once{}
	ep = nil
	initErr = nil
	newEmbeddedPython = func(string) (*embedpython.EmbeddedPython, error) { return &embedpython.EmbeddedPython{}, nil }
	defer func() { once = sync.Once{}; ep = oldEP; initErr = oldErr; newEmbeddedPython = oldNew }()

	err := Init()
	Println(err == nil, GetPython() != nil)
	// Output: true true
}

func ExampleGetPython() {
	oldEP := ep
	ep = &embedpython.EmbeddedPython{}
	defer func() { ep = oldEP }()
	Println(GetPython() != nil)
	// Output: true
}

func ExampleRunScript() {
	cleanup := examplePythonRuntime("script-ok")
	defer cleanup()
	out, err := RunScript(Background(), "print('ok')")
	Println(err == nil, out)
	// Output: true script-ok
}

func ExampleRunModule() {
	cleanup := examplePythonRuntime("module-ok")
	defer cleanup()
	out, err := RunModule(Background(), "json.tool")
	Println(err == nil, out)
	// Output: true module-ok
}

func ExampleDevOpsPath() {
	old := Getenv("DEVOPS_PATH")
	Setenv("DEVOPS_PATH", "/tmp/devops")
	defer Setenv("DEVOPS_PATH", old)
	path, err := DevOpsPath()
	Println(err == nil, path)
	// Output: true /tmp/devops
}

func ExampleCoolifyModulePath() {
	old := Getenv("DEVOPS_PATH")
	Setenv("DEVOPS_PATH", "/tmp/devops")
	defer Setenv("DEVOPS_PATH", old)
	path, err := CoolifyModulePath()
	Println(err == nil, path)
	// Output: true /tmp/devops/playbooks/roles/coolify/module_utils
}

func ExampleCoolifyScript() {
	old := Getenv("DEVOPS_PATH")
	Setenv("DEVOPS_PATH", "/tmp/devops")
	defer Setenv("DEVOPS_PATH", old)
	script, err := CoolifyScript("https://coolify.example", "token", "list-servers", map[string]any{"limit": 1})
	Println(err == nil, Contains(script, "list-servers"))
	// Output: true true
}
