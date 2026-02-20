// Package devkit provides a developer toolkit for common automation commands.
// LEK-1 | lthn.ai | EUPL-1.2
package devkit

import (
	"encoding/json"
	"fmt"
	"strings"
)

// VulnFinding represents a single vulnerability found by govulncheck.
type VulnFinding struct {
	ID             string   // e.g. GO-2024-1234
	Aliases        []string // CVE/GHSA aliases
	Package        string   // Affected package path
	CalledFunction string   // Function in call stack (empty if not called)
	Description    string   // Human-readable summary
	Severity       string   // "HIGH", "MEDIUM", "LOW", or empty
	FixedVersion   string   // Version that contains the fix
	ModulePath     string   // Go module path
}

// VulnResult holds the complete output of a vulnerability scan.
type VulnResult struct {
	Findings []VulnFinding
	Module   string // Module path that was scanned
}

// --- govulncheck JSON wire types ---

// govulncheckMessage represents a single JSON line from govulncheck -json output.
type govulncheckMessage struct {
	Config   *govulncheckConfig `json:"config,omitempty"`
	OSV      *govulncheckOSV    `json:"osv,omitempty"`
	Finding  *govulncheckFind   `json:"finding,omitempty"`
	Progress *json.RawMessage   `json:"progress,omitempty"`
}

type govulncheckConfig struct {
	GoVersion  string `json:"go_version"`
	ModulePath string `json:"module_path"`
}

type govulncheckOSV struct {
	ID       string              `json:"id"`
	Aliases  []string            `json:"aliases"`
	Summary  string              `json:"summary"`
	Affected []govulncheckAffect `json:"affected"`
}

type govulncheckAffect struct {
	Package  *govulncheckPkg      `json:"package,omitempty"`
	Ranges   []govulncheckRange   `json:"ranges,omitempty"`
	Severity []govulncheckSeverity `json:"database_specific,omitempty"`
}

type govulncheckPkg struct {
	Name      string `json:"name"`
	Ecosystem string `json:"ecosystem"`
}

type govulncheckRange struct {
	Events []govulncheckEvent `json:"events"`
}

type govulncheckEvent struct {
	Fixed string `json:"fixed,omitempty"`
}

type govulncheckSeverity struct {
	Severity string `json:"severity,omitempty"`
}

type govulncheckFind struct {
	OSV   string               `json:"osv"`
	Trace []govulncheckTrace   `json:"trace"`
}

type govulncheckTrace struct {
	Module   string `json:"module,omitempty"`
	Package  string `json:"package,omitempty"`
	Function string `json:"function,omitempty"`
	Version  string `json:"version,omitempty"`
}

// VulnCheck runs govulncheck -json on the given module path and parses
// the output into structured VulnFindings.
func (t *Toolkit) VulnCheck(modulePath string) (*VulnResult, error) {
	if modulePath == "" {
		modulePath = "./..."
	}

	stdout, stderr, exitCode, err := t.Run("govulncheck", "-json", modulePath)
	if err != nil && exitCode == -1 {
		return nil, fmt.Errorf("govulncheck not installed or not available: %w", err)
	}

	return ParseVulnCheckJSON(stdout, stderr)
}

// ParseVulnCheckJSON parses govulncheck -json output (newline-delimited JSON messages).
func ParseVulnCheckJSON(stdout, stderr string) (*VulnResult, error) {
	result := &VulnResult{}

	// Collect OSV entries and findings separately, then correlate.
	osvMap := make(map[string]*govulncheckOSV)
	var findings []govulncheckFind

	// Parse line-by-line to gracefully skip malformed entries.
	// json.Decoder.More() hangs on non-JSON input, so we split first.
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var msg govulncheckMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			// Skip malformed lines — govulncheck sometimes emits progress text
			continue
		}

		if msg.Config != nil {
			result.Module = msg.Config.ModulePath
		}
		if msg.OSV != nil {
			osvMap[msg.OSV.ID] = msg.OSV
		}
		if msg.Finding != nil {
			findings = append(findings, *msg.Finding)
		}
	}

	// Build VulnFindings by correlating findings with OSV metadata.
	for _, f := range findings {
		finding := VulnFinding{
			ID: f.OSV,
		}

		// Extract package, function, and module from trace.
		if len(f.Trace) > 0 {
			// The first trace entry is the called function in user code;
			// the last is the vulnerable symbol.
			last := f.Trace[len(f.Trace)-1]
			finding.Package = last.Package
			finding.CalledFunction = last.Function
			finding.ModulePath = last.Module

			// If the trace has a version, capture it.
			for _, tr := range f.Trace {
				if tr.Version != "" {
					finding.FixedVersion = tr.Version
					break
				}
			}
		}

		// Enrich from OSV entry.
		if osv, ok := osvMap[f.OSV]; ok {
			finding.Description = osv.Summary
			finding.Aliases = osv.Aliases

			// Extract fixed version and severity from affected entries.
			for _, aff := range osv.Affected {
				for _, r := range aff.Ranges {
					for _, ev := range r.Events {
						if ev.Fixed != "" && finding.FixedVersion == "" {
							finding.FixedVersion = ev.Fixed
						}
					}
				}
			}
		}

		result.Findings = append(result.Findings, finding)
	}

	return result, nil
}

// LEK-1 | lthn.ai | EUPL-1.2
