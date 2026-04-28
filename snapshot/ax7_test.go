package snapshot

import (
	. "dappco.re/go"
	"dappco.re/go/scm/manifest"
)

func ax7Manifest() *manifest.Manifest {
	return &manifest.Manifest{
		Code:        "agent",
		Name:        "Agent",
		Version:     "1.2.3",
		Description: "test manifest",
		Modules:     []string{"core/devops"},
	}
}

func TestAX7_Generate_Good(t *T) {
	data, err := Generate(ax7Manifest(), "abc123", "v1.2.3")
	AssertNoError(t, err)
	AssertContains(t, string(data), `"code": "agent"`)
	AssertContains(t, string(data), `"tag": "v1.2.3"`)
}

func TestAX7_Generate_Bad(t *T) {
	data, err := Generate(nil, "abc123", "v1.2.3")
	AssertError(t, err)

	AssertNil(t, data)
	AssertContains(t, err.Error(), "manifest is nil")
}

func TestAX7_Generate_Ugly(t *T) {
	data, err := Generate(&manifest.Manifest{}, "", "")
	AssertNoError(t, err)

	AssertContains(t, string(data), `"schema": 1`)
	AssertContains(t, string(data), `"built":`)
}

func TestAX7_GenerateAt_Good(t *T) {
	built := UnixTime(1770000000).UTC()
	data, err := GenerateAt(ax7Manifest(), "deadbeef", "v1.2.3", built)
	AssertNoError(t, err)

	AssertContains(t, string(data), `"commit": "deadbeef"`)
	AssertContains(t, string(data), TimeFormat(built, TimeRFC3339))
}

func TestAX7_GenerateAt_Bad(t *T) {
	data, err := GenerateAt(nil, "deadbeef", "v1.2.3", UnixTime(0))
	AssertError(t, err)

	AssertNil(t, data)
	AssertContains(t, err.Error(), "manifest is nil")
}

func TestAX7_GenerateAt_Ugly(t *T) {
	m := ax7Manifest()
	m.Permissions.Read = []string{"."}
	data, err := GenerateAt(m, "", "", UnixTime(-1).UTC())

	AssertNoError(t, err)
	AssertContains(t, string(data), `"permissions"`)
	AssertContains(t, string(data), `1969-12-31T23:59:59Z`)
}
