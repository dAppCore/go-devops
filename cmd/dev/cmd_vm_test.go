package dev

import (
	"testing"

	"dappco.re/go/cli/pkg/cli"
)

func TestAddVMStatusCommand_Good(t *testing.T) {
	root := &cli.Command{Use: "core"}

	AddDevCommands(root)

	statusCmd, _, err := root.Find([]string{"dev", "status"})
	mustNoError(t, err)
	mustNotNil(t, statusCmd)
	mustEqual(t, "status", statusCmd.Use)
	mustContainsString(t, statusCmd.Aliases, "vm-status")

	aliasCmd, _, err := root.Find([]string{"dev", "vm-status"})
	mustNoError(t, err)
	mustNotNil(t, aliasCmd)
	mustTrue(t, statusCmd == aliasCmd)
}
