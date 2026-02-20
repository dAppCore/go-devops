// LEK-1 | lthn.ai | EUPL-1.2
package devkit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleVulnJSON = `{"config":{"module_path":"example.com/mymod","go_version":"go1.22.0"}}
{"progress":{"message":"Scanning your code..."}}
{"osv":{"id":"GO-2024-0001","aliases":["CVE-2024-1234","GHSA-abcd-1234"],"summary":"Buffer overflow in net/http","affected":[{"package":{"name":"stdlib","ecosystem":"Go"},"ranges":[{"events":[{"fixed":"1.22.1"}]}]}]}}
{"osv":{"id":"GO-2024-0002","aliases":["CVE-2024-5678"],"summary":"Path traversal in archive/zip","affected":[{"package":{"name":"stdlib","ecosystem":"Go"},"ranges":[{"events":[{"fixed":"1.21.9"}]}]}]}}
{"finding":{"osv":"GO-2024-0001","trace":[{"module":"example.com/mymod","package":"example.com/mymod/server","function":"HandleRequest"},{"module":"stdlib","package":"net/http","function":"ReadRequest","version":"go1.22.0"}]}}
{"finding":{"osv":"GO-2024-0002","trace":[{"module":"stdlib","package":"archive/zip","function":"OpenReader","version":"go1.22.0"}]}}
`

func TestParseVulnCheckJSON_Good(t *testing.T) {
	result, err := ParseVulnCheckJSON(sampleVulnJSON, "")
	require.NoError(t, err)

	assert.Equal(t, "example.com/mymod", result.Module)
	assert.Len(t, result.Findings, 2)

	// First finding: GO-2024-0001
	f0 := result.Findings[0]
	assert.Equal(t, "GO-2024-0001", f0.ID)
	assert.Equal(t, "net/http", f0.Package)
	assert.Equal(t, "ReadRequest", f0.CalledFunction)
	assert.Equal(t, "Buffer overflow in net/http", f0.Description)
	assert.Contains(t, f0.Aliases, "CVE-2024-1234")
	assert.Contains(t, f0.Aliases, "GHSA-abcd-1234")
	assert.Equal(t, "go1.22.0", f0.FixedVersion) // from trace version

	// Second finding: GO-2024-0002
	f1 := result.Findings[1]
	assert.Equal(t, "GO-2024-0002", f1.ID)
	assert.Equal(t, "archive/zip", f1.Package)
	assert.Equal(t, "OpenReader", f1.CalledFunction)
	assert.Equal(t, "Path traversal in archive/zip", f1.Description)
	assert.Contains(t, f1.Aliases, "CVE-2024-5678")
}

func TestParseVulnCheckJSON_EmptyOutput_Good(t *testing.T) {
	result, err := ParseVulnCheckJSON("", "")
	require.NoError(t, err)
	assert.Empty(t, result.Findings)
	assert.Empty(t, result.Module)
}

func TestParseVulnCheckJSON_ConfigOnly_Good(t *testing.T) {
	input := `{"config":{"module_path":"example.com/clean","go_version":"go1.23.0"}}
`
	result, err := ParseVulnCheckJSON(input, "")
	require.NoError(t, err)
	assert.Equal(t, "example.com/clean", result.Module)
	assert.Empty(t, result.Findings)
}

func TestParseVulnCheckJSON_MalformedLines_Bad(t *testing.T) {
	input := `not valid json
{"config":{"module_path":"example.com/mod"}}
also broken {{{
{"osv":{"id":"GO-2024-0099","summary":"Test vuln","aliases":[],"affected":[]}}
{"finding":{"osv":"GO-2024-0099","trace":[{"module":"stdlib","package":"crypto/tls","function":"Dial"}]}}
`
	result, err := ParseVulnCheckJSON(input, "")
	require.NoError(t, err)
	assert.Equal(t, "example.com/mod", result.Module)
	assert.Len(t, result.Findings, 1)
	assert.Equal(t, "GO-2024-0099", result.Findings[0].ID)
	assert.Equal(t, "Dial", result.Findings[0].CalledFunction)
}

func TestParseVulnCheckJSON_FindingWithoutOSV_Bad(t *testing.T) {
	// Finding references an OSV ID that was never emitted — should still parse.
	input := `{"finding":{"osv":"GO-2024-UNKNOWN","trace":[{"module":"example.com/mod","package":"example.com/mod/pkg","function":"DoStuff"}]}}
`
	result, err := ParseVulnCheckJSON(input, "")
	require.NoError(t, err)
	assert.Len(t, result.Findings, 1)

	f := result.Findings[0]
	assert.Equal(t, "GO-2024-UNKNOWN", f.ID)
	assert.Equal(t, "example.com/mod/pkg", f.Package)
	assert.Equal(t, "DoStuff", f.CalledFunction)
	assert.Empty(t, f.Description) // No OSV entry to enrich from
	assert.Empty(t, f.Aliases)
}

func TestParseVulnCheckJSON_NoTrace_Bad(t *testing.T) {
	input := `{"osv":{"id":"GO-2024-0050","summary":"Empty trace test","aliases":["CVE-2024-0050"],"affected":[]}}
{"finding":{"osv":"GO-2024-0050","trace":[]}}
`
	result, err := ParseVulnCheckJSON(input, "")
	require.NoError(t, err)
	assert.Len(t, result.Findings, 1)

	f := result.Findings[0]
	assert.Equal(t, "GO-2024-0050", f.ID)
	assert.Equal(t, "Empty trace test", f.Description)
	assert.Empty(t, f.Package)
	assert.Empty(t, f.CalledFunction)
}

func TestParseVulnCheckJSON_MultipleFindings_Good(t *testing.T) {
	input := `{"osv":{"id":"GO-2024-0010","summary":"Vuln A","aliases":["CVE-A"],"affected":[{"ranges":[{"events":[{"fixed":"1.20.5"}]}]}]}}
{"osv":{"id":"GO-2024-0011","summary":"Vuln B","aliases":["CVE-B"],"affected":[]}}
{"osv":{"id":"GO-2024-0012","summary":"Vuln C","aliases":["CVE-C"],"affected":[{"ranges":[{"events":[{"fixed":"1.21.0"}]}]}]}}
{"finding":{"osv":"GO-2024-0010","trace":[{"package":"net/http","function":"Serve"}]}}
{"finding":{"osv":"GO-2024-0011","trace":[{"package":"encoding/xml","function":"Unmarshal"}]}}
{"finding":{"osv":"GO-2024-0012","trace":[{"package":"os/exec","function":"Command"}]}}
`
	result, err := ParseVulnCheckJSON(input, "")
	require.NoError(t, err)
	assert.Len(t, result.Findings, 3)

	assert.Equal(t, "Vuln A", result.Findings[0].Description)
	assert.Equal(t, "1.20.5", result.Findings[0].FixedVersion)
	assert.Equal(t, "Vuln B", result.Findings[1].Description)
	assert.Equal(t, "Vuln C", result.Findings[2].Description)
	assert.Equal(t, "1.21.0", result.Findings[2].FixedVersion)
}

func TestParseVulnCheckJSON_FixedVersionFromOSV_Good(t *testing.T) {
	// When trace has no version, fixed version should come from OSV affected ranges.
	input := `{"osv":{"id":"GO-2024-0077","summary":"Test","aliases":[],"affected":[{"ranges":[{"events":[{"fixed":"0.9.1"}]}]}]}}
{"finding":{"osv":"GO-2024-0077","trace":[{"package":"example.com/lib","function":"Process"}]}}
`
	result, err := ParseVulnCheckJSON(input, "")
	require.NoError(t, err)
	require.Len(t, result.Findings, 1)
	assert.Equal(t, "0.9.1", result.Findings[0].FixedVersion)
}

func TestVulnCheck_NotInstalled_Ugly(t *testing.T) {
	setupMockCmdExit(t, "govulncheck-nonexistent", "", "", 1)
	// Don't mock govulncheck — ensure it handles missing binary gracefully
	// We'll rely on the binary not being in the test temp PATH.

	tk := New(t.TempDir())
	// Remove PATH to simulate govulncheck not found
	t.Setenv("PATH", t.TempDir())
	_, err := tk.VulnCheck("./...")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not installed or not available")
}

func TestVulnCheck_WithMock_Good(t *testing.T) {
	// Mock govulncheck to return our sample JSON
	setupMockCmd(t, "govulncheck", sampleVulnJSON)

	tk := New(t.TempDir())
	result, err := tk.VulnCheck("./...")
	require.NoError(t, err)
	assert.Equal(t, "example.com/mymod", result.Module)
	assert.Len(t, result.Findings, 2)
}

func TestVulnCheck_DefaultModulePath_Good(t *testing.T) {
	setupMockCmd(t, "govulncheck", `{"config":{"module_path":"default/mod"}}`)

	tk := New(t.TempDir())
	result, err := tk.VulnCheck("")
	require.NoError(t, err)
	assert.Equal(t, "default/mod", result.Module)
}

func TestParseVulnCheckJSON_ProgressOnly_Good(t *testing.T) {
	input := `{"progress":{"message":"Scanning..."}}
{"progress":{"message":"Done"}}
`
	result, err := ParseVulnCheckJSON(input, "")
	require.NoError(t, err)
	assert.Empty(t, result.Findings)
}

func TestParseVulnCheckJSON_ModulePathFromTrace_Good(t *testing.T) {
	input := `{"finding":{"osv":"GO-2024-0099","trace":[{"module":"example.com/vulnerable","package":"example.com/vulnerable/pkg","function":"Bad","version":"v1.2.3"}]}}
`
	result, err := ParseVulnCheckJSON(input, "")
	require.NoError(t, err)
	require.Len(t, result.Findings, 1)
	assert.Equal(t, "example.com/vulnerable", result.Findings[0].ModulePath)
	assert.Equal(t, "v1.2.3", result.Findings[0].FixedVersion)
}

// LEK-1 | lthn.ai | EUPL-1.2
