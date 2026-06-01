package snapshot

import (
	. "dappco.re/go"
	"dappco.re/go/scm/manifest"
)

// TestSnapshot_GenerateAt_PermissionsWrite covers the Write arm of the
// permissions guard (snapshot.go: m.Permissions.Write != nil).
func TestSnapshot_GenerateAt_PermissionsWrite(t *T) {
	m := testManifest()
	m.Permissions.Write = []string{"/var/lib/agent"}
	data, r := GenerateAt(m, "c", "v1", UnixTime(0).UTC())

	AssertTrue(t, r.OK)
	AssertContains(t, string(data), `"permissions"`)
	AssertContains(t, string(data), `"write"`)
}

// TestSnapshot_GenerateAt_PermissionsNet covers the Net arm of the
// permissions guard (snapshot.go: m.Permissions.Net != nil).
func TestSnapshot_GenerateAt_PermissionsNet(t *T) {
	m := testManifest()
	m.Permissions.Net = []string{"api.example.test:443"}
	data, r := GenerateAt(m, "c", "v1", UnixTime(0).UTC())

	AssertTrue(t, r.OK)
	AssertContains(t, string(data), `"net"`)
}

// TestSnapshot_GenerateAt_PermissionsRun covers the Run arm of the
// permissions guard (snapshot.go: m.Permissions.Run != nil).
func TestSnapshot_GenerateAt_PermissionsRun(t *T) {
	m := testManifest()
	m.Permissions.Run = []string{"git"}
	data, r := GenerateAt(m, "c", "v1", UnixTime(0).UTC())

	AssertTrue(t, r.OK)
	AssertContains(t, string(data), `"run"`)
}

// TestSnapshot_GenerateAt_NoPermissions verifies the permissions block is
// omitted from the snapshot when every permission slice is nil — the
// false branch of the four-way guard.
func TestSnapshot_GenerateAt_NoPermissions(t *T) {
	m := &manifest.Manifest{Code: "x", Name: "X", Version: "0.0.1"}
	data, r := GenerateAt(m, "c", "v1", UnixTime(0).UTC())

	AssertTrue(t, r.OK)
	AssertNotContains(t, string(data), `"permissions"`)
}

// TestSnapshot_GenerateAt_ErrorOpPrefix asserts the failed Result carries
// the operation prefix produced by log.E, not just the bare message —
// guards against regression of the double-wrap empty-render bug where
// core.Fail(log.E(...)) yielded an empty Error() string.
func TestSnapshot_GenerateAt_ErrorOpPrefix(t *T) {
	_, r := GenerateAt(nil, "c", "v1", UnixTime(0).UTC())

	AssertFalse(t, r.OK)
	AssertContains(t, r.Error(), "snapshot")
	AssertContains(t, r.Error(), "manifest is nil")
}
