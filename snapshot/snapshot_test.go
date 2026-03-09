package snapshot

import (
	"encoding/json"
	"testing"

	"forge.lthn.ai/core/go-scm/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerate_Good(t *testing.T) {
	m := &manifest.Manifest{
		Code:        "test-app",
		Name:        "Test App",
		Version:     "1.0.0",
		Description: "A test application",
		Daemons: map[string]manifest.DaemonSpec{
			"serve": {Binary: "core-php", Args: []string{"php", "serve"}, Default: true},
		},
		Modules: []string{"core/media"},
	}

	data, err := Generate(m, "abc123def456", "v1.0.0")
	require.NoError(t, err)

	var snap Snapshot
	require.NoError(t, json.Unmarshal(data, &snap))

	assert.Equal(t, 1, snap.Schema)
	assert.Equal(t, "test-app", snap.Code)
	assert.Equal(t, "1.0.0", snap.Version)
	assert.Equal(t, "abc123def456", snap.Commit)
	assert.Equal(t, "v1.0.0", snap.Tag)
	assert.NotEmpty(t, snap.Built)
	assert.Len(t, snap.Daemons, 1)
	assert.Equal(t, "core-php", snap.Daemons["serve"].Binary)
}

func TestGenerate_Good_NoDaemons(t *testing.T) {
	m := &manifest.Manifest{
		Code:    "simple",
		Name:    "Simple",
		Version: "0.1.0",
	}

	data, err := Generate(m, "abc123", "v0.1.0")
	require.NoError(t, err)

	var snap Snapshot
	require.NoError(t, json.Unmarshal(data, &snap))

	assert.Equal(t, "simple", snap.Code)
	assert.Nil(t, snap.Daemons)
}
