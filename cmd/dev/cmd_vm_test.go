package dev

import (
	"testing"

	"github.com/stretchr/testify/require"

	"forge.lthn.ai/core/cli/pkg/cli"
)

func TestAddVMStatusCommand_Good(t *testing.T) {
	root := &cli.Command{Use: "core"}

	AddDevCommands(root)

	statusCmd, _, err := root.Find([]string{"dev", "status"})
	require.NoError(t, err)
	require.NotNil(t, statusCmd)
	require.Equal(t, "status", statusCmd.Use)
	require.Contains(t, statusCmd.Aliases, "vm-status")

	aliasCmd, _, err := root.Find([]string{"dev", "vm-status"})
	require.NoError(t, err)
	require.NotNil(t, aliasCmd)
	require.Equal(t, statusCmd, aliasCmd)
}
