package gitcmd

import core "dappco.re/go"

func TestCmdGit_AddGitCommands_Good(t *core.T) {
	c := core.New()
	r := AddGitCommands(c)
	core.AssertTrue(t, r.OK)

	core.AssertTrue(t, c.Command("git").OK)

	gitSubcommands := 0
	for _, path := range c.Commands() {
		if core.HasPrefix(path, "git/") {
			gitSubcommands++
		}
	}
	core.AssertGreaterOrEqual(t, gitSubcommands, 7)
}

func TestCmdGit_AddGitCommands_Bad(t *core.T) {
	var c *core.Core
	core.AssertPanics(t, func() {
		AddGitCommands(c)
	})
	core.AssertNil(t, c)
}

func TestCmdGit_AddGitCommands_Ugly(t *core.T) {
	c := core.New()
	core.AssertTrue(t, AddGitCommands(c).OK)

	// Re-registering onto the same Core hits the duplicate-executable guard.
	r := AddGitCommands(c)
	core.AssertFalse(t, r.OK)
}
