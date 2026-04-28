package python

import (
	"os/exec"
	"sync"

	. "dappco.re/go"
	embedpython "github.com/kluctl/go-embed-python/python"
)

func ax7ResetPythonHooks(t *T) {
	oldEP := ep
	oldInitErr := initErr
	oldNew := newEmbeddedPython
	oldInitRuntime := initRuntime
	oldPythonCommand := pythonCommand

	once = sync.Once{}
	ep = nil
	initErr = nil

	t.Cleanup(func() {
		once = sync.Once{}
		ep = oldEP
		initErr = oldInitErr
		newEmbeddedPython = oldNew
		initRuntime = oldInitRuntime
		pythonCommand = oldPythonCommand
	})
}

func TestAX7_Init_Good(t *T) {
	ax7ResetPythonHooks(t)
	newEmbeddedPython = func(string) (*embedpython.EmbeddedPython, error) {
		return &embedpython.EmbeddedPython{}, nil
	}

	err := Init()
	AssertNoError(t, err)
	AssertNotNil(t, GetPython())
}

func TestAX7_Init_Bad(t *T) {
	ax7ResetPythonHooks(t)
	newEmbeddedPython = func(string) (*embedpython.EmbeddedPython, error) {
		return nil, AnError
	}

	err := Init()
	AssertErrorIs(t, err, AnError)
	AssertNil(t, GetPython())
}

func TestAX7_Init_Ugly(t *T) {
	ax7ResetPythonHooks(t)
	calls := 0
	newEmbeddedPython = func(string) (*embedpython.EmbeddedPython, error) {
		calls++
		return &embedpython.EmbeddedPython{}, nil
	}

	AssertNoError(t, Init())
	AssertNoError(t, Init())
	AssertEqual(t, 1, calls)
}

func TestAX7_GetPython_Good(t *T) {
	ax7ResetPythonHooks(t)
	ep = &embedpython.EmbeddedPython{}
	got := GetPython()

	AssertNotNil(t, got)
	AssertTrue(t, got == ep)
}

func TestAX7_GetPython_Bad(t *T) {
	ax7ResetPythonHooks(t)
	got := GetPython()

	AssertNil(t, got)
	AssertNil(t, ep)
}

func TestAX7_GetPython_Ugly(t *T) {
	ax7ResetPythonHooks(t)
	newEmbeddedPython = func(string) (*embedpython.EmbeddedPython, error) {
		return &embedpython.EmbeddedPython{}, nil
	}
	RequireNoError(t, Init())

	AssertNotNil(t, GetPython())
	AssertTrue(t, GetPython() == ep)
}

func TestAX7_RunScript_Good(t *T) {
	ax7ResetPythonHooks(t)
	initRuntime = func() error { return nil }
	pythonCommand = func(args ...string) (*exec.Cmd, error) {
		AssertLen(t, args, 1)
		return exec.Command("sh", "-c", "printf script-ok"), nil
	}

	out, err := RunScript(Background(), "print('ignored')")
	AssertNoError(t, err)
	AssertEqual(t, "script-ok", out)
}

func TestAX7_RunScript_Bad(t *T) {
	ax7ResetPythonHooks(t)
	initRuntime = func() error { return AnError }

	out, err := RunScript(Background(), "print('ignored')")
	AssertErrorIs(t, err, AnError)
	AssertEqual(t, "", out)
}

func TestAX7_RunScript_Ugly(t *T) {
	ax7ResetPythonHooks(t)
	initRuntime = func() error { return nil }
	pythonCommand = func(args ...string) (*exec.Cmd, error) {
		AssertLen(t, args, 2)
		return exec.Command("sh", "-c", "printf script-failed >&2; exit 7"), nil
	}

	out, err := RunScript(Background(), "print('ignored')", "--flag")
	AssertError(t, err)
	AssertEqual(t, "", out)
}

func TestAX7_RunModule_Good(t *T) {
	ax7ResetPythonHooks(t)
	initRuntime = func() error { return nil }
	pythonCommand = func(args ...string) (*exec.Cmd, error) {
		AssertEqual(t, []string{"-m", "json.tool"}, args)
		return exec.Command("sh", "-c", "printf module-ok"), nil
	}

	out, err := RunModule(Background(), "json.tool")
	AssertNoError(t, err)
	AssertEqual(t, "module-ok", out)
}

func TestAX7_RunModule_Bad(t *T) {
	ax7ResetPythonHooks(t)
	initRuntime = func() error { return AnError }

	out, err := RunModule(Background(), "json.tool")
	AssertErrorIs(t, err, AnError)
	AssertEqual(t, "", out)
}

func TestAX7_RunModule_Ugly(t *T) {
	ax7ResetPythonHooks(t)
	initRuntime = func() error { return nil }
	pythonCommand = func(args ...string) (*exec.Cmd, error) {
		AssertEqual(t, []string{"-m", "missing.module", "--help"}, args)
		return exec.Command("sh", "-c", "exit 9"), nil
	}

	out, err := RunModule(Background(), "missing.module", "--help")
	AssertError(t, err)
	AssertEqual(t, "", out)
}

func TestAX7_DevOpsPath_Good(t *T) {
	t.Setenv("DEVOPS_PATH", "/tmp/devops")
	path, err := DevOpsPath()

	AssertNoError(t, err)
	AssertEqual(t, "/tmp/devops", path)
}

func TestAX7_DevOpsPath_Bad(t *T) {
	t.Setenv("DEVOPS_PATH", "")
	path, err := DevOpsPath()

	AssertNoError(t, err)
	AssertContains(t, path, "Code/DevOps")
}

func TestAX7_DevOpsPath_Ugly(t *T) {
	t.Setenv("DEVOPS_PATH", "/tmp/dev ops")
	path, err := DevOpsPath()

	AssertNoError(t, err)
	AssertEqual(t, "/tmp/dev ops", path)
}

func TestAX7_CoolifyModulePath_Good(t *T) {
	t.Setenv("DEVOPS_PATH", "/tmp/devops")
	path, err := CoolifyModulePath()

	AssertNoError(t, err)
	AssertEqual(t, "/tmp/devops/playbooks/roles/coolify/module_utils", path)
}

func TestAX7_CoolifyModulePath_Bad(t *T) {
	t.Setenv("DEVOPS_PATH", "")
	path, err := CoolifyModulePath()

	AssertNoError(t, err)
	AssertContains(t, path, "playbooks/roles/coolify/module_utils")
}

func TestAX7_CoolifyModulePath_Ugly(t *T) {
	t.Setenv("DEVOPS_PATH", "/tmp/dev ops")
	path, err := CoolifyModulePath()

	AssertNoError(t, err)
	AssertContains(t, path, "/tmp/dev ops/")
}

func TestAX7_CoolifyScript_Good(t *T) {
	t.Setenv("DEVOPS_PATH", "/tmp/devops")
	script, err := CoolifyScript("https://coolify.example", "token", "list-servers", map[string]any{"limit": 1})

	AssertNoError(t, err)
	AssertContains(t, script, "list-servers")
	AssertContains(t, script, "https://coolify.example")
}

func TestAX7_CoolifyScript_Bad(t *T) {
	t.Setenv("DEVOPS_PATH", "/tmp/devops")
	script, err := CoolifyScript("https://coolify.example", "token", "bad", map[string]any{"bad": func() {}})

	AssertError(t, err)
	AssertEqual(t, "", script)
}

func TestAX7_CoolifyScript_Ugly(t *T) {
	t.Setenv("DEVOPS_PATH", "/tmp/devops")
	script, err := CoolifyScript("", "", "", nil)

	AssertNoError(t, err)
	AssertContains(t, script, "CoolifyClient")
	AssertContains(t, script, "json.loads")
}
