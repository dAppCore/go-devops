package snapshot

import (
	"encoding/json"
	"testing"
	"time"

	"forge.lthn.ai/core/go-scm/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fixedTime = time.Date(2026, 3, 9, 15, 0, 0, 0, time.UTC)

func TestGenerate_Good(t *testing.T) {
	m := &manifest.Manifest{
		Code:        "test-app",
		Name:        "Test App",
		Version:     "1.0.0",
		Description: "A test application",
		Layout:      "HLCRF",
		Slots:       map[string]string{"C": "main-content"},
		Daemons: map[string]manifest.DaemonSpec{
			"serve": {Binary: "core-php", Args: []string{"php", "serve"}, Default: true},
		},
		Permissions: manifest.Permissions{
			Read: []string{"./photos/"},
		},
		Modules: []string{"core/media"},
	}

	data, err := GenerateAt(m, "abc123def456", "v1.0.0", fixedTime)
	require.NoError(t, err)

	var snap Snapshot
	require.NoError(t, json.Unmarshal(data, &snap))

	assert.Equal(t, 1, snap.Schema)
	assert.Equal(t, "test-app", snap.Code)
	assert.Equal(t, "Test App", snap.Name)
	assert.Equal(t, "1.0.0", snap.Version)
	assert.Equal(t, "A test application", snap.Description)
	assert.Equal(t, "abc123def456", snap.Commit)
	assert.Equal(t, "v1.0.0", snap.Tag)
	assert.Equal(t, "2026-03-09T15:00:00Z", snap.Built)
	assert.Equal(t, "HLCRF", snap.Layout)
	assert.Equal(t, "main-content", snap.Slots["C"])
	assert.Len(t, snap.Daemons, 1)
	assert.Equal(t, "core-php", snap.Daemons["serve"].Binary)
	require.NotNil(t, snap.Permissions)
	assert.Equal(t, []string{"./photos/"}, snap.Permissions.Read)
	assert.Equal(t, []string{"core/media"}, snap.Modules)
}

func TestGenerate_Good_NoDaemons(t *testing.T) {
	m := &manifest.Manifest{
		Code:    "simple",
		Name:    "Simple",
		Version: "0.1.0",
	}

	data, err := GenerateAt(m, "abc123", "v0.1.0", fixedTime)
	require.NoError(t, err)

	var snap Snapshot
	require.NoError(t, json.Unmarshal(data, &snap))

	assert.Equal(t, 1, snap.Schema)
	assert.Equal(t, "simple", snap.Code)
	assert.Nil(t, snap.Daemons)
	assert.Nil(t, snap.Permissions)
}

func TestGenerate_Bad_NilManifest(t *testing.T) {
	_, err := Generate(nil, "abc123", "v1.0.0")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "manifest is nil")
}
