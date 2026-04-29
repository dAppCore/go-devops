package snapshot

import (
	. "dappco.re/go"
	"dappco.re/go/scm/manifest"
)

func testManifest() *manifest.Manifest {
	return &manifest.Manifest{
		Code:        "agent",
		Name:        "Agent",
		Version:     "1.2.3",
		Description: "release automation",
		Daemons: map[string]manifest.DaemonSpec{
			"worker": {Binary: "agentd", Args: []string{"serve"}},
		},
		Layout:  "service",
		Slots:   map[string]string{"config": "/etc/agent"},
		Modules: []string{"deploy"},
	}
}

func TestSnapshot_Generate_Good(t *T) {
	data, err := Generate(testManifest(), "abc123", "v1.2.3")
	AssertNoError(t, err)
	AssertContains(t, string(data), `"code": "agent"`)
	AssertContains(t, string(data), `"tag": "v1.2.3"`)
}

func TestSnapshot_Generate_Bad(t *T) {
	data, err := Generate(nil, "abc123", "v1.2.3")
	AssertError(t, err)

	AssertNil(t, data)
	AssertContains(t, err.Error(), "manifest is nil")
}

func TestSnapshot_Generate_Ugly(t *T) {
	data, err := Generate(&manifest.Manifest{}, "", "")
	AssertNoError(t, err)

	AssertContains(t, string(data), `"schema": 1`)
	AssertContains(t, string(data), `"built":`)
}

func TestSnapshot_GenerateAt_Good(t *T) {
	built := UnixTime(1770000000).UTC()
	data, err := GenerateAt(testManifest(), "deadbeef", "v1.2.3", built)
	AssertNoError(t, err)

	AssertContains(t, string(data), `"commit": "deadbeef"`)
	AssertContains(t, string(data), TimeFormat(built, TimeRFC3339))
}

func TestSnapshot_GenerateAt_Bad(t *T) {
	data, err := GenerateAt(nil, "deadbeef", "v1.2.3", UnixTime(0))
	AssertError(t, err)

	AssertNil(t, data)
	AssertContains(t, err.Error(), "manifest is nil")
}

func TestSnapshot_GenerateAt_Ugly(t *T) {
	m := testManifest()
	m.Permissions.Read = []string{"."}
	data, err := GenerateAt(m, "", "", UnixTime(-1).UTC())

	AssertNoError(t, err)
	AssertContains(t, string(data), `"permissions"`)
	AssertContains(t, string(data), `1969-12-31T23:59:59Z`)
}
