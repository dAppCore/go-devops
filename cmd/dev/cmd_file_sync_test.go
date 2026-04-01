package dev

import (
	"testing"

	"github.com/stretchr/testify/require"

	"forge.lthn.ai/core/cli/pkg/cli"
)

func TestAddFileSyncCommand_Good(t *testing.T) {
	root := &cli.Command{Use: "core"}

	AddDevCommands(root)

	syncCmd, _, err := root.Find([]string{"dev", "sync"})
	require.NoError(t, err)
	require.NotNil(t, syncCmd)

	yesFlag := syncCmd.Flags().Lookup("yes")
	require.NotNil(t, yesFlag)
	require.Equal(t, "y", yesFlag.Shorthand)

	require.NotNil(t, syncCmd.Flags().Lookup("dry-run"))
	require.NotNil(t, syncCmd.Flags().Lookup("push"))
}
