package dev

import (
	core "dappco.re/go"
	"slices"
	"testing"
)

func TestAddFileSyncCommand_Good(t *testing.T) {
	c := core.New()
	if r := AddDevCommands(c); !r.OK {
		t.Fatalf("AddDevCommands: %v", r.Error())
	}

	r := c.Command("dev/sync")
	if !r.OK {
		t.Fatal("expected sync command")
	}
	syncCmd := r.Value.(*core.Command)
	if syncCmd.Action == nil {
		t.Fatal("expected sync command to be executable")
	}

	// core.Command.Flags carries declared keys + defaults only — cobra's
	// shorthand ("-y" for "--yes") has no equivalent in the new model.
	if !syncCmd.Flags.Has("yes") {
		t.Fatal("expected yes flag")
	}
	if !syncCmd.Flags.Has("dry-run") {
		t.Fatal("expected dry-run flag")
	}
	if !syncCmd.Flags.Has("push") {
		t.Fatal("expected push flag")
	}
}

func TestSplitPatterns_Good(t *testing.T) {
	patterns := splitPatterns("packages/core-*,  apps/* ,services/*,")
	want := []string{"packages/core-*", "apps/*", "services/*"}
	if !slices.Equal(patterns, want) {
		t.Fatalf("patterns = %v, want %v", patterns, want)
	}
}

func TestMatchGlob_Good(t *testing.T) {
	trueCases := []struct {
		name    string
		pattern string
	}{
		{name: "packages/core-xyz", pattern: "packages/core-*"},
		{name: "packages/core-xyz", pattern: "*/core-*"},
		{name: "a-b", pattern: "a?b"},
		{name: "foo", pattern: "foo"},
	}
	for _, tc := range trueCases {
		if !matchGlob(tc.name, tc.pattern) {
			t.Fatalf("matchGlob(%q, %q) = false, want true", tc.name, tc.pattern)
		}
	}

	falseCases := []struct {
		name    string
		pattern string
	}{
		{name: "core-other", pattern: "packages/*"},
		{name: "abc", pattern: "[]"},
	}
	for _, tc := range falseCases {
		if matchGlob(tc.name, tc.pattern) {
			t.Fatalf("matchGlob(%q, %q) = true, want false", tc.name, tc.pattern)
		}
	}
}

func TestCmdFileSync_AddFileSyncCommand_Good(t *core.T) {
	c := core.New()
	r := AddFileSyncCommand(c, "dev")
	core.AssertTrue(t, r.OK)

	cmd := c.Command("dev/sync")
	core.AssertTrue(t, cmd.OK)
	core.AssertTrue(t, cmd.Value.(*core.Command).Flags.Has("to"))
}

func TestCmdFileSync_AddFileSyncCommand_Bad(t *core.T) {
	var c *core.Core
	core.AssertPanics(t, func() {
		AddFileSyncCommand(c, "dev")
	})
	core.AssertNil(t, c)
}

func TestCmdFileSync_AddFileSyncCommand_Ugly(t *core.T) {
	c := core.New()
	core.AssertTrue(t, AddFileSyncCommand(c, "dev").OK)

	r := AddFileSyncCommand(c, "git")
	core.AssertTrue(t, r.OK)
	core.AssertTrue(t, c.Command("git/sync").OK)
}
