package snapshot

import (
	"encoding/json"
	"testing"
	"time"

	"dappco.re/go/scm/manifest"
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
	mustNoError(t, err)

	var snap Snapshot
	mustNoError(t, json.Unmarshal(data, &snap))

	mustEqual(t, 1, snap.Schema)
	mustEqual(t, "test-app", snap.Code)
	mustEqual(t, "Test App", snap.Name)
	mustEqual(t, "1.0.0", snap.Version)
	mustEqual(t, "A test application", snap.Description)
	mustEqual(t, "abc123def456", snap.Commit)
	mustEqual(t, "v1.0.0", snap.Tag)
	mustEqual(t, "2026-03-09T15:00:00Z", snap.Built)
	mustEqual(t, "HLCRF", snap.Layout)
	mustEqual(t, "main-content", snap.Slots["C"])
	mustLenMap(t, snap.Daemons, 1)
	mustEqual(t, "core-php", snap.Daemons["serve"].Binary)
	if snap.Permissions == nil {
		t.Fatal("expected non-nil permissions")
	}
	mustDeepEqual(t, []string{"./photos/"}, snap.Permissions.Read)
	mustDeepEqual(t, []string{"core/media"}, snap.Modules)
}

func TestGenerate_NoDaemons_Good(t *testing.T) {
	m := &manifest.Manifest{
		Code:    "simple",
		Name:    "Simple",
		Version: "0.1.0",
	}

	data, err := GenerateAt(m, "abc123", "v0.1.0", fixedTime)
	mustNoError(t, err)

	var snap Snapshot
	mustNoError(t, json.Unmarshal(data, &snap))

	mustEqual(t, 1, snap.Schema)
	mustEqual(t, "simple", snap.Code)
	if snap.Daemons != nil {
		t.Fatalf("expected nil daemons, got %v", snap.Daemons)
	}
	if snap.Permissions != nil {
		t.Fatalf("expected nil permissions, got %v", snap.Permissions)
	}
}

func TestGenerate_NilManifest_Bad(t *testing.T) {
	_, err := Generate(nil, "abc123", "v1.0.0")
	mustErrorContains(t, err, "manifest is nil")
}
