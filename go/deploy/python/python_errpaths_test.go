package python

import (
	. "dappco.re/go"
)

// TestPython_RunScript_CommandCreateFails covers the pythonCommand
// construction-error arm of RunScript (python.go: cmd, err := pythonCommand;
// if err != nil) which the existing Ugly test does not reach — there the
// command constructs and only Output() fails.
func TestPython_RunScript_CommandCreateFails(t *T) {
	resetPythonHooks(t)
	initRuntime = func() Result { return Ok(nil) }
	pythonCommand = func(args ...string) (pythonRunner, error) {
		return nil, AnError
	}

	out, r := RunScript(Background(), "print('x')")
	AssertFalse(t, r.OK)
	AssertEqual(t, "", out)
}

// TestPython_RunModule_CommandCreateFails covers the matching
// construction-error arm of RunModule.
func TestPython_RunModule_CommandCreateFails(t *T) {
	resetPythonHooks(t)
	initRuntime = func() Result { return Ok(nil) }
	pythonCommand = func(args ...string) (pythonRunner, error) {
		return nil, AnError
	}

	out, r := RunModule(Background(), "json.tool")
	AssertFalse(t, r.OK)
	AssertEqual(t, "", out)
}

// TestPython_CoolifyScript_MarshalFails covers the JSON marshal-error arm
// of CoolifyScript by passing a params map containing a value that the
// JSON encoder cannot serialise (a channel).
func TestPython_CoolifyScript_MarshalFails(t *T) {
	t.Setenv("DEVOPS_PATH", "/tmp/devops")
	params := map[string]any{"bad": make(chan int)}

	script, r := CoolifyScript("https://coolify.example", "tok", "list-servers", params)
	AssertFalse(t, r.OK)
	AssertEqual(t, "", script)
}
