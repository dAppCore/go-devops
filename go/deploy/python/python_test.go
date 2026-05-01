package python

import (
	. "dappco.re/go"
	embedpython "github.com/kluctl/go-embed-python/python"
	"sync"
)

type fakePythonRunner struct {
	output []byte
	err    error
}

func (f fakePythonRunner) Output() ([]byte, error) {
	return f.output, f.err
}

func resetPythonHooks(t *T) {
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

func TestPython_Init_Good(t *T) {
	resetPythonHooks(t)
	newEmbeddedPython = func(string) (*embedpython.EmbeddedPython, error) {
		return &embedpython.EmbeddedPython{}, nil
	}

	r := Init()
	AssertTrue(t, r.OK)
	AssertNotNil(t, GetPython())
}

func TestPython_Init_Bad(t *T) {
	resetPythonHooks(t)
	newEmbeddedPython = func(string) (*embedpython.EmbeddedPython, error) {
		return nil, AnError
	}

	r := Init()
	AssertFalse(t, r.OK)
	AssertErrorIs(t, r.Value.(error), AnError)
	AssertNil(t, GetPython())
}

func TestPython_Init_Ugly(t *T) {
	resetPythonHooks(t)
	calls := 0
	newEmbeddedPython = func(string) (*embedpython.EmbeddedPython, error) {
		calls++
		return &embedpython.EmbeddedPython{}, nil
	}

	AssertTrue(t, Init().OK)
	AssertTrue(t, Init().OK)
	AssertEqual(t, 1, calls)
}

func TestPython_GetPython_Good(t *T) {
	resetPythonHooks(t)
	ep = &embedpython.EmbeddedPython{}
	got := GetPython()

	AssertNotNil(t, got)
	AssertTrue(t, got == ep)
}

func TestPython_GetPython_Bad(t *T) {
	resetPythonHooks(t)
	got := GetPython()

	AssertNil(t, got)
	AssertNil(t, ep)
}

func TestPython_GetPython_Ugly(t *T) {
	resetPythonHooks(t)
	newEmbeddedPython = func(string) (*embedpython.EmbeddedPython, error) {
		return &embedpython.EmbeddedPython{}, nil
	}
	RequireTrue(t, Init().OK)

	AssertNotNil(t, GetPython())
	AssertTrue(t, GetPython() == ep)
}

func TestPython_RunScript_Good(t *T) {
	resetPythonHooks(t)
	initRuntime = func() Result { return Ok(nil) }
	pythonCommand = func(args ...string) (pythonRunner, error) {
		AssertLen(t, args, 1)
		return fakePythonRunner{output: []byte("script-ok")}, nil
	}

	out, r := RunScript(Background(), "print('ignored')")
	AssertTrue(t, r.OK)
	AssertEqual(t, "script-ok", out)
}

func TestPython_RunScript_Bad(t *T) {
	resetPythonHooks(t)
	initRuntime = func() Result { return Fail(AnError) }

	out, r := RunScript(Background(), "print('ignored')")
	AssertFalse(t, r.OK)
	AssertErrorIs(t, r.Value.(error), AnError)
	AssertEqual(t, "", out)
}

func TestPython_RunScript_Ugly(t *T) {
	resetPythonHooks(t)
	initRuntime = func() Result { return Ok(nil) }
	pythonCommand = func(args ...string) (pythonRunner, error) {
		AssertLen(t, args, 2)
		return fakePythonRunner{err: AnError}, nil
	}

	out, r := RunScript(Background(), "print('ignored')", "--flag")
	AssertFalse(t, r.OK)
	AssertEqual(t, "", out)
}

func TestPython_RunModule_Good(t *T) {
	resetPythonHooks(t)
	initRuntime = func() Result { return Ok(nil) }
	pythonCommand = func(args ...string) (pythonRunner, error) {
		AssertEqual(t, []string{"-m", "json.tool"}, args)
		return fakePythonRunner{output: []byte("module-ok")}, nil
	}

	out, r := RunModule(Background(), "json.tool")
	AssertTrue(t, r.OK)
	AssertEqual(t, "module-ok", out)
}

func TestPython_RunModule_Bad(t *T) {
	resetPythonHooks(t)
	initRuntime = func() Result { return Fail(AnError) }

	out, r := RunModule(Background(), "json.tool")
	AssertFalse(t, r.OK)
	AssertErrorIs(t, r.Value.(error), AnError)
	AssertEqual(t, "", out)
}

func TestPython_RunModule_Ugly(t *T) {
	resetPythonHooks(t)
	initRuntime = func() Result { return Ok(nil) }
	pythonCommand = func(args ...string) (pythonRunner, error) {
		AssertEqual(t, []string{"-m", "missing.module", "--help"}, args)
		return fakePythonRunner{err: AnError}, nil
	}

	out, r := RunModule(Background(), "missing.module", "--help")
	AssertFalse(t, r.OK)
	AssertEqual(t, "", out)
}

func TestPython_DevOpsPath_Good(t *T) {
	t.Setenv("DEVOPS_PATH", "/tmp/devops")
	path, r := DevOpsPath()

	AssertTrue(t, r.OK)
	AssertEqual(t, "/tmp/devops", path)
}

func TestPython_DevOpsPath_Bad(t *T) {
	t.Setenv("DEVOPS_PATH", "")
	path, r := DevOpsPath()

	AssertTrue(t, r.OK)
	AssertContains(t, path, "Code/DevOps")
}

func TestPython_DevOpsPath_Ugly(t *T) {
	t.Setenv("DEVOPS_PATH", "/tmp/dev ops")
	path, r := DevOpsPath()

	AssertTrue(t, r.OK)
	AssertEqual(t, "/tmp/dev ops", path)
}

func TestPython_CoolifyModulePath_Good(t *T) {
	t.Setenv("DEVOPS_PATH", "/tmp/devops")
	path, r := CoolifyModulePath()

	AssertTrue(t, r.OK)
	AssertEqual(t, "/tmp/devops/playbooks/roles/coolify/module_utils", path)
}

func TestPython_CoolifyModulePath_Bad(t *T) {
	t.Setenv("DEVOPS_PATH", "")
	path, r := CoolifyModulePath()

	AssertTrue(t, r.OK)
	AssertContains(t, path, "playbooks/roles/coolify/module_utils")
}

func TestPython_CoolifyModulePath_Ugly(t *T) {
	t.Setenv("DEVOPS_PATH", "/tmp/dev ops")
	path, r := CoolifyModulePath()

	AssertTrue(t, r.OK)
	AssertContains(t, path, "/tmp/dev ops/")
}

func TestPython_CoolifyScript_Good(t *T) {
	t.Setenv("DEVOPS_PATH", "/tmp/devops")
	script, r := CoolifyScript("https://coolify.example", "token", "list-servers", map[string]any{"limit": 1})

	AssertTrue(t, r.OK)
	AssertContains(t, script, "list-servers")
	AssertContains(t, script, "https://coolify.example")
}

func TestPython_CoolifyScript_Bad(t *T) {
	t.Setenv("DEVOPS_PATH", "/tmp/devops")
	script, r := CoolifyScript("https://coolify.example", "token", "bad", map[string]any{"bad": func() {}})

	AssertFalse(t, r.OK)
	AssertEqual(t, "", script)
}

func TestPython_CoolifyScript_Ugly(t *T) {
	t.Setenv("DEVOPS_PATH", "/tmp/devops")
	script, r := CoolifyScript("", "", "", nil)

	AssertTrue(t, r.OK)
	AssertContains(t, script, "CoolifyClient")
	AssertContains(t, script, "json.loads")
}
