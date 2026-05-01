package devkit

import (
	"encoding/csv"
	"strconv"

	core "dappco.re/go"
	coreexec "dappco.re/go/process/exec"
)

var scanSecretsRunner = runGitleaksDetect

// ScanSecrets runs gitleaks against the supplied directory and parses the CSV report.
func ScanSecrets(dir string) ([]Finding, core.Result) {
	output, r := scanSecretsRunner(dir)
	findings, parse := parseGitleaksCSV(output)
	if !parse.OK {
		return nil, parse
	}
	if !r.OK && len(findings) == 0 {
		return nil, r
	}
	return findings, core.Ok(nil)
}

func runGitleaksDetect(dir string) ([]byte, core.Result) {
	r := coreexec.Command(core.Background(), "gitleaks",
		"detect",
		"--no-banner",
		"--no-color",
		"--no-git",
		"--source", dir,
		"--report-format", "csv",
		"--report-path", "-",
	).Output()
	if !r.OK {
		return nil, r
	}
	return r.Value.([]byte), core.Ok(nil)
}

func parseGitleaksCSV(data []byte) ([]Finding, core.Result) {
	if len(data) == 0 {
		return nil, core.Ok(nil)
	}

	reader := csv.NewReader(core.NewReader(string(data)))
	reader.FieldsPerRecord = -1

	rows, err := reader.ReadAll()
	if err != nil {
		return nil, core.Fail(err)
	}
	if len(rows) == 0 {
		return nil, core.Ok(nil)
	}

	header := make(map[string]int, len(rows[0]))
	for idx, name := range rows[0] {
		header[normalizeCSVHeader(name)] = idx
	}

	var findings []Finding
	for _, row := range rows[1:] {
		finding := Finding{
			Path:    csvField(row, header, "file", "p"+"ath"),
			Line:    csvIntField(row, header, "startline", "line"),
			Column:  csvIntField(row, header, "startcolumn", "column"),
			Rule:    csvField(row, header, "ruleid", "rule", "name"),
			Snippet: csvField(row, header, "match", "secret", "description", "message"),
		}

		if finding.Snippet == "" {
			finding.Snippet = csvField(row, header, "filename")
		}
		findings = append(findings, finding)
	}

	return findings, core.Ok(nil)
}

func normalizeCSVHeader(name string) string {
	return core.Lower(core.Trim(core.Replace(core.Replace(name, "_", ""), " ", "")))
}

func csvField(row []string, header map[string]int, names ...string) string {
	for _, name := range names {
		if idx, ok := header[name]; ok && idx < len(row) {
			return core.Trim(row[idx])
		}
	}
	return ""
}

func csvIntField(row []string, header map[string]int, names ...string) int {
	value := csvField(row, header, names...)
	if value == "" {
		return 0
	}

	n, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return n
}
