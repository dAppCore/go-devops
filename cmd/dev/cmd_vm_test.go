package dev

import (
	"slices"
	"testing"

	"dappco.re/go/cli/pkg/cli"
)

func TestAddVMStatusCommand_Good(t *testing.T) {
	root := &cli.Command{Use: "core"}

	AddDevCommands(root)

	statusCmd, _, err := root.Find([]string{"dev", "status"})
	if err != nil {
		t.Fatalf("find status command: %v", err)
	}
	if statusCmd == nil {
		t.Fatal("expected status command")
	}
	if statusCmd.Use != "status" {
		t.Fatalf("status command use = %q, want %q", statusCmd.Use, "status")
	}
	if !slices.Contains(statusCmd.Aliases, "vm-status") {
		t.Fatalf("status aliases = %v, want vm-status", statusCmd.Aliases)
	}

	aliasCmd, _, err := root.Find([]string{"dev", "vm-status"})
	if err != nil {
		t.Fatalf("find vm-status alias: %v", err)
	}
	if aliasCmd == nil {
		t.Fatal("expected vm-status alias command")
	}
	if statusCmd != aliasCmd {
		t.Fatal("expected vm-status alias to resolve to status command")
	}
}
