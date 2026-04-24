package dev

import (
	"reflect"
	"testing"

	"dappco.re/go/cli/pkg/cli"
)

func TestAddFileSyncCommand_Good(t *testing.T) {
	root := &cli.Command{Use: "core"}

	AddDevCommands(root)

	syncCmd, _, err := root.Find([]string{"dev", "sync"})
	mustNoError(t, err)
	if syncCmd == nil {
		t.Fatal("expected non-nil sync command")
	}

	yesFlag := syncCmd.Flags().Lookup("yes")
	if yesFlag == nil {
		t.Fatal("expected yes flag")
	}
	mustEqual(t, "y", yesFlag.Shorthand)

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
	if !reflect.DeepEqual(want, patterns) {
		t.Fatalf("want %v, got %v", want, patterns)
	}
}

func TestMatchGlob_Good(t *testing.T) {
	mustTrue(t, matchGlob("packages/core-xyz", "packages/core-*"))
	mustTrue(t, matchGlob("packages/core-xyz", "*/core-*"))
	mustTrue(t, matchGlob("a-b", "a?b"))
	mustTrue(t, matchGlob("foo", "foo"))
	mustFalse(t, matchGlob("core-other", "packages/*"))
	mustFalse(t, matchGlob("abc", "[]"))
}
