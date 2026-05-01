package deploy

import core "dappco.re/go"

func TestCmdDeploy_Cmd_Good(t *core.T) {
	core.AssertEqual(t, "deploy", Cmd.Use)
	core.AssertNotNil(t, Cmd.PersistentFlags())
	core.AssertGreaterOrEqual(t, len(Cmd.Commands()), 1)
}

func TestCmdDeploy_Cmd_Bad(t *core.T) {
	original := Cmd.Use
	Cmd.Use = ""
	t.Cleanup(func() { Cmd.Use = original })

	core.AssertEqual(t, "", Cmd.Use)
	core.AssertNotNil(t, Cmd.PersistentFlags())
}

func TestCmdDeploy_Cmd_Ugly(t *core.T) {
	commands := Cmd.Commands()

	core.AssertGreaterOrEqual(t, len(commands), 1)
	core.AssertNotNil(t, Cmd)
}
