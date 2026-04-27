package dev

import (
	"testing"

	"dappco.re/go/cli/pkg/cli"
)

func TestAddFileSyncCommand_Good(t *testing.T) {
	root := &cli.Command{Use: "core"}

	AddDevCommands(root)

	syncCmd, _, err := root.Find([]string{"dev", "sync"})
	mustNoError(t, err)
	mustNotNil(t, syncCmd)

	yesFlag := syncCmd.Flags().Lookup("yes")
	mustNotNil(t, yesFlag)
	mustEqual(t, "y", yesFlag.Shorthand)

	mustNotNil(t, syncCmd.Flags().Lookup("dry-run"))
	mustNotNil(t, syncCmd.Flags().Lookup("push"))
}

func TestSplitPatterns_Good(t *testing.T) {
	patterns := splitPatterns("packages/core-*,  apps/* ,services/*,")
	want := []string{"packages/core-*", "apps/*", "services/*"}
	mustDeepEqual(t, want, patterns)
}

func TestMatchGlob_Good(t *testing.T) {
	mustTrue(t, matchGlob("packages/core-xyz", "packages/core-*"))
	mustTrue(t, matchGlob("packages/core-xyz", "*/core-*"))
	mustTrue(t, matchGlob("a-b", "a?b"))
	mustTrue(t, matchGlob("foo", "foo"))
	mustFalse(t, matchGlob("core-other", "packages/*"))
	mustFalse(t, matchGlob("abc", "[]"))
}
