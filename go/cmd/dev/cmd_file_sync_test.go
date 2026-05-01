package dev

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
	"slices"
	"testing"
)

func TestAddFileSyncCommand_Good(t *testing.T) {
	root := &cli.Command{Use: "core"}

	AddDevCommands(root)

	syncCmd, _, err := root.Find([]string{"dev", "sync"})
	if err != nil {
		t.Fatalf("find sync command: %v", err)
	}
	if syncCmd == nil {
		t.Fatal("expected sync command")
	}

	yesFlag := syncCmd.Flags().Lookup("yes")
	if yesFlag == nil {
		t.Fatal("expected yes flag")
	}
	if yesFlag.Shorthand != "y" {
		t.Fatalf("yes shorthand = %q, want %q", yesFlag.Shorthand, "y")
	}

	if syncCmd.Flags().Lookup("dry-run") == nil {
		t.Fatal("expected dry-run flag")
	}
	if syncCmd.Flags().Lookup("push") == nil {
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
	root := &cli.Command{Use: "root"}
	AddFileSyncCommand(root)
	cmd := testCommand(root, "sync")

	core.AssertNotNil(t, cmd)
	core.AssertNotNil(t, cmd.Flag("to"))
}

func TestCmdFileSync_AddFileSyncCommand_Bad(t *core.T) {
	var root *cli.Command
	core.AssertPanics(t, func() {
		AddFileSyncCommand(root)
	})
	core.AssertNil(t, root)
}

func TestCmdFileSync_AddFileSyncCommand_Ugly(t *core.T) {
	root := &cli.Command{Use: "root"}
	AddFileSyncCommand(root)
	AddFileSyncCommand(root)

	core.AssertLen(t, root.Commands(), 2)
	core.AssertNotNil(t, testCommand(root, "sync"))
}
