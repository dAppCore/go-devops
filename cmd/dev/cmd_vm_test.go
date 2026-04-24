package dev

import (
	"testing"

	"dappco.re/go/core/cli/pkg/cli"
)

func TestAddVMStatusCommand_Good(t *testing.T) {
	root := &cli.Command{Use: "core"}

	AddDevCommands(root)

	statusCmd, _, err := root.Find([]string{"dev", "status"})
	mustNoError(t, err)
	if statusCmd == nil {
		t.Fatal("expected non-nil status command")
	}
	mustEqual(t, "status", statusCmd.Use)
	mustContainsAlias(t, statusCmd.Aliases, "vm-status")

	aliasCmd, _, err := root.Find([]string{"dev", "vm-status"})
	mustNoError(t, err)
	if aliasCmd == nil {
		t.Fatal("expected non-nil alias command")
	}
	if statusCmd != aliasCmd {
		t.Fatalf("want alias to be same command, got %v vs %v", statusCmd, aliasCmd)
	}
}

func mustContainsAlias(t *testing.T, haystack []string, needle string) {
	t.Helper()
	for _, s := range haystack {
		if s == needle {
			return
		}
	}
	t.Fatalf("expected %v to contain %q", haystack, needle)
}
