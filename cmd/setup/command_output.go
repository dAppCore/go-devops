package setup

import (
	core "dappco.re/go"
	coreexec "dappco.re/go/process/exec"
)

func commandCombinedOutput(cmd *coreexec.Cmd) ([]byte, core.Result) {
	buf := core.NewBuffer()
	r := cmd.WithStdout(buf).WithStderr(buf).Run()
	output := buf.Bytes()
	if !r.OK {
		text := core.Trim(string(output))
		if text != "" {
			return output, core.Fail(core.Errorf("%s", text))
		}
		return output, r
	}
	return output, core.Ok(nil)
}
