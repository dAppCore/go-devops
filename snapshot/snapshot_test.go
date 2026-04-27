package snapshot

import (
	"encoding/json"
	"slices"
	"strings"
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
	if err != nil {
		t.Fatalf("generate snapshot: %v", err)
	}

	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		t.Fatalf("unmarshal snapshot: %v", err)
	}

	if snap.Schema != 1 {
		t.Fatalf("schema = %d, want 1", snap.Schema)
	}
	for name, check := range map[string]struct {
		got  string
		want string
	}{
		"code":        {got: snap.Code, want: "test-app"},
		"name":        {got: snap.Name, want: "Test App"},
		"version":     {got: snap.Version, want: "1.0.0"},
		"description": {got: snap.Description, want: "A test application"},
		"commit":      {got: snap.Commit, want: "abc123def456"},
		"tag":         {got: snap.Tag, want: "v1.0.0"},
		"built":       {got: snap.Built, want: "2026-03-09T15:00:00Z"},
		"layout":      {got: snap.Layout, want: "HLCRF"},
		"slot C":      {got: snap.Slots["C"], want: "main-content"},
	} {
		if check.got != check.want {
			t.Fatalf("%s = %q, want %q", name, check.got, check.want)
		}
	}
	if len(snap.Daemons) != 1 {
		t.Fatalf("daemons length = %d, want 1", len(snap.Daemons))
	}
	if snap.Daemons["serve"].Binary != "core-php" {
		t.Fatalf("serve binary = %q, want core-php", snap.Daemons["serve"].Binary)
	}
	if snap.Permissions == nil {
		t.Fatal("expected non-nil permissions")
	}
	if !slices.Equal(snap.Permissions.Read, []string{"./photos/"}) {
		t.Fatalf("permission reads = %v, want [./photos/]", snap.Permissions.Read)
	}
	if !slices.Equal(snap.Modules, []string{"core/media"}) {
		t.Fatalf("modules = %v, want [core/media]", snap.Modules)
	}
}

func TestGenerate_NoDaemons_Good(t *testing.T) {
	m := &manifest.Manifest{
		Code:    "simple",
		Name:    "Simple",
		Version: "0.1.0",
	}

	data, err := GenerateAt(m, "abc123", "v0.1.0", fixedTime)
	if err != nil {
		t.Fatalf("generate snapshot: %v", err)
	}

	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		t.Fatalf("unmarshal snapshot: %v", err)
	}

	if snap.Schema != 1 {
		t.Fatalf("schema = %d, want 1", snap.Schema)
	}
	if snap.Code != "simple" {
		t.Fatalf("code = %q, want simple", snap.Code)
	}
	if snap.Daemons != nil {
		t.Fatalf("expected nil daemons, got %v", snap.Daemons)
	}
	if snap.Permissions != nil {
		t.Fatalf("expected nil permissions, got %v", snap.Permissions)
	}
}

func TestGenerate_NilManifest_Bad(t *testing.T) {
	_, err := Generate(nil, "abc123", "v1.0.0")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "manifest is nil") {
		t.Fatalf("error = %q, want substring %q", err.Error(), "manifest is nil")
	}
}
