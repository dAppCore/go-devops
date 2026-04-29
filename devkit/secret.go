package devkit

import (
	"bufio"
	"io/fs"
	"regexp"

	core "dappco.re/go"
)

// Finding describes a secret-like match discovered while scanning source files.
type Finding struct {
	Path    string `json:"file"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Rule    string `json:"rule"`
	Snippet string `json:"snippet"`
}

var secretRules = []struct {
	name  string
	match *regexp.Regexp
}{
	{
		name:  "aws-access-key-id",
		match: regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`),
	},
	{
		name:  "github-token",
		match: regexp.MustCompile(`\bgh[pousr]_[A-Za-z0-9_]{20,}\b`),
	},
	{
		name:  "generic-secret-assignment",
		match: regexp.MustCompile(`(?i)\b(?:api[_-]?key|client[_-]?secret|secret|token|password)\b\s*[:=]\s*["']?([A-Za-z0-9._\-+/]{8,})["']?`),
	},
}

var skipDirs = map[string]struct{}{
	".git":         {},
	"vendor":       {},
	"node_modules": {},
}

var textExts = map[string]struct{}{
	".go":     {},
	".md":     {},
	".txt":    {},
	".json":   {},
	".yaml":   {},
	".yml":    {},
	".toml":   {},
	".env":    {},
	".ini":    {},
	".cfg":    {},
	".conf":   {},
	".sh":     {},
	".tf":     {},
	".tfvars": {},
}

// ScanDir recursively scans a directory for secret-like patterns.
func ScanDir(root string) ([]Finding, coreFailure) {
	var findings []Finding

	if err := core.PathWalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		name := d.Name()
		if d.IsDir() {
			if _, ok := skipDirs[name]; ok || core.HasPrefix(name, ".") && path != root {
				return core.PathSkipDir
			}
			return nil
		}

		if !isTextCandidate(name) {
			return nil
		}

		fileFindings, err := scanFile(path)
		if err != nil {
			return err
		}
		findings = append(findings, fileFindings...)
		return nil
	}); err != nil {
		return nil, err
	}

	return findings, nil
}

func scanFile(path string) ([]Finding, coreFailure) {
	data, err := fileRead(path)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 || hasNUL(data) {
		return nil, nil
	}

	var findings []Finding
	scanner := bufio.NewScanner(core.NewReader(string(data)))
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := scanner.Text()
		matchedSpecific := false
		for _, rule := range secretRules {
			if rule.name == "generic-secret-assignment" && matchedSpecific {
				continue
			}
			if loc := rule.match.FindStringIndex(line); loc != nil {
				findings = append(findings, Finding{
					Path:    path,
					Line:    lineNo,
					Column:  loc[0] + 1,
					Rule:    rule.name,
					Snippet: core.Trim(line),
				})
				if rule.name != "generic-secret-assignment" {
					matchedSpecific = true
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return findings, nil
}

func isTextCandidate(name string) bool {
	if ext := core.Lower(core.PathExt(name)); ext != "" {
		_, ok := textExts[ext]
		return ok
	}
	// Allow extension-less files such as Makefile, LICENSE, and .env.
	switch name {
	case "Makefile", "Dockerfile", "LICENSE", "README", "CLAUDE.md":
		return true
	}
	return core.HasPrefix(name, ".")
}

// fileRead is factored out for tests.
var fileRead = func(path string) ([]byte, error) {
	r := core.ReadFile(path)
	if !r.OK {
		return nil, r.Value.(error)
	}
	return r.Value.([]byte), nil
}

func hasNUL(data []byte) bool {
	for _, b := range data {
		if b == 0 {
			return true
		}
	}
	return false
}
