package dev

import (
	"testing"

	"github.com/stretchr/testify/require"

	"dappco.re/go/core/cli/pkg/cli"
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

func TestSplitPatterns_Good(t *testing.T) {
	patterns := splitPatterns("packages/core-*,  apps/* ,services/*,")
	require.Equal(t, []string{"packages/core-*", "apps/*", "services/*"}, patterns)
}

func TestMatchGlob_Good(t *testing.T) {
	require.True(t, matchGlob("packages/core-xyz", "packages/core-*"))
	require.True(t, matchGlob("packages/core-xyz", "*/core-*"))
	require.True(t, matchGlob("a-b", "a?b"))
	require.True(t, matchGlob("foo", "foo"))
	require.False(t, matchGlob("core-other", "packages/*"))
	require.False(t, matchGlob("abc", "[]"))
}
