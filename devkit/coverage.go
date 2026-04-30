package devkit

import (
	"regexp"
	"sort"
	"strconv"
	"time"

	core "dappco.re/go"
)

// CoveragePackage describes coverage for a single package or directory.
type CoveragePackage struct {
	Name              string  `json:"name"`
	CoveredStatements int     `json:"covered_statements"`
	TotalStatements   int     `json:"total_statements"`
	Coverage          float64 `json:"coverage"`
}

// CoverageSnapshot captures a point-in-time view of coverage across packages.
type CoverageSnapshot struct {
	CapturedAt time.Time         `json:"captured_at"`
	Packages   []CoveragePackage `json:"packages"`
	Total      CoveragePackage   `json:"total"`
}

// CoverageDelta describes how a single package changed between snapshots.
type CoverageDelta struct {
	Name     string  `json:"name"`
	Previous float64 `json:"previous"`
	Current  float64 `json:"current"`
	Delta    float64 `json:"delta"`
}

// CoverageComparison summarises the differences between two coverage snapshots.
type CoverageComparison struct {
	Regressions  []CoverageDelta   `json:"regressions"`
	Improvements []CoverageDelta   `json:"improvements"`
	NewPackages  []CoveragePackage `json:"new_packages"`
	Removed      []CoveragePackage `json:"removed"`
	TotalDelta   float64           `json:"total_delta"`
}

// CoverageStore persists coverage snapshots to disk.
type CoverageStore struct {
	path string
}

type coverageBucket struct {
	covered int
	total   int
}

var coverProfileLineRE = regexp.MustCompile(`^(.+?):\d+\.\d+,\d+\.\d+\s+(\d+)\s+(\d+)$`)
var coverOutputLineRE = regexp.MustCompile(`^(?:ok|\?)?\s*(\S+)\s+.*coverage:\s+([0-9]+(?:\.[0-9]+)?)% of statements$`)

// NewCoverageStore creates a store backed by the given file path.
func NewCoverageStore(path string) *CoverageStore {
	return &CoverageStore{path: path}
}

// Append stores a new snapshot, creating the parent directory if needed.
func (s *CoverageStore) Append(snapshot CoverageSnapshot) (_ core.Result) {
	if r := core.MkdirAll(core.PathDir(s.path), 0o755); !r.OK {
		return r
	}

	snapshots, r := s.Load()
	if !r.OK {
		return r
	}

	snapshot.CapturedAt = snapshot.CapturedAt.UTC()
	snapshots = append(snapshots, snapshot)

	data := core.JSONMarshalIndent(snapshots, "", "  ")
	if !data.OK {
		return data
	}

	if r := core.WriteFile(s.path, data.Value.([]byte), 0o600); !r.OK {
		return r
	}
	return core.Ok(nil)
}

// Load reads all snapshots from disk.
func (s *CoverageStore) Load() ([]CoverageSnapshot, core.Result) {
	read := core.ReadFile(s.path)
	if !read.OK {
		err := read.Value.(error)
		if core.IsNotExist(err) {
			return nil, core.Ok(nil)
		}
		return nil, read
	}
	data := read.Value.([]byte)
	if len(core.Trim(string(data))) == 0 {
		return nil, core.Ok(nil)
	}

	var snapshots []CoverageSnapshot
	if r := core.JSONUnmarshal(data, &snapshots); !r.OK {
		return nil, r
	}
	return snapshots, core.Ok(nil)
}

// Latest returns the newest snapshot in the store.
func (s *CoverageStore) Latest() (CoverageSnapshot, core.Result) {
	snapshots, r := s.Load()
	if !r.OK {
		return CoverageSnapshot{}, r
	}
	if len(snapshots) == 0 {
		return CoverageSnapshot{}, core.Fail(core.Errorf("coverage store is empty"))
	}
	return snapshots[len(snapshots)-1], core.Ok(nil)
}

// ParseCoverProfile parses go test -coverprofile output into a coverage snapshot.
func ParseCoverProfile(data string) (CoverageSnapshot, core.Result) {
	if core.Trim(data) == "" {
		return CoverageSnapshot{}, core.Ok(nil)
	}

	packages := make(map[string]*coverageBucket)
	total := coverageBucket{}

	for _, rawLine := range core.Split(core.Trim(data), "\n") {
		line := core.Trim(rawLine)
		if line == "" || core.HasPrefix(line, "mode:") {
			continue
		}

		match := coverProfileLineRE.FindStringSubmatch(line)
		if match == nil {
			return CoverageSnapshot{}, core.Fail(core.Errorf("invalid cover profile line: %s", line))
		}

		file := core.PathToSlash(match[1])
		stmts, err := strconv.Atoi(match[2])
		if err != nil {
			return CoverageSnapshot{}, core.Fail(err)
		}
		count, err := strconv.Atoi(match[3])
		if err != nil {
			return CoverageSnapshot{}, core.Fail(err)
		}

		dir := core.PathDir(file)
		if dir == "" {
			dir = "."
		}

		b := packages[dir]
		if b == nil {
			b = &coverageBucket{}
			packages[dir] = b
		}
		b.total += stmts
		total.total += stmts
		if count > 0 {
			b.covered += stmts
			total.covered += stmts
		}
	}

	return snapshotFromBuckets(packages, total), core.Ok(nil)
}

// ParseCoverOutput parses human-readable go test -cover output into a snapshot.
func ParseCoverOutput(output string) (CoverageSnapshot, core.Result) {
	if core.Trim(output) == "" {
		return CoverageSnapshot{}, core.Ok(nil)
	}

	packages := make(map[string]*CoveragePackage)
	var total CoveragePackage

	for _, rawLine := range core.Split(core.Trim(output), "\n") {
		line := core.Trim(rawLine)
		if line == "" {
			continue
		}

		match := coverOutputLineRE.FindStringSubmatch(line)
		if match == nil {
			continue
		}

		name := match[1]
		coverage, err := strconv.ParseFloat(match[2], 64)
		if err != nil {
			return CoverageSnapshot{}, core.Fail(err)
		}

		pkg := &CoveragePackage{
			Name:     name,
			Coverage: coverage,
		}
		packages[name] = pkg

		total.Coverage += coverage
		total.TotalStatements++
	}

	if len(packages) == 0 {
		return CoverageSnapshot{}, core.Ok(nil)
	}

	snapshot := CoverageSnapshot{
		CapturedAt: time.Now().UTC(),
		Packages:   make([]CoveragePackage, 0, len(packages)),
	}

	for _, pkg := range packages {
		snapshot.Packages = append(snapshot.Packages, *pkg)
	}
	sort.Slice(snapshot.Packages, func(i, j int) bool {
		return snapshot.Packages[i].Name < snapshot.Packages[j].Name
	})

	snapshot.Total.Name = "total"
	if total.TotalStatements > 0 {
		snapshot.Total.Coverage = total.Coverage / float64(total.TotalStatements)
	}
	return snapshot, core.Ok(nil)
}

// CompareCoverage compares two snapshots and reports regressions and improvements.
func CompareCoverage(previous, current CoverageSnapshot) CoverageComparison {
	prevPackages := coverageMap(previous.Packages)
	currPackages := coverageMap(current.Packages)

	comparison := CoverageComparison{
		NewPackages: make([]CoveragePackage, 0),
		Removed:     make([]CoveragePackage, 0),
	}

	for name, curr := range currPackages {
		prev, ok := prevPackages[name]
		if !ok {
			comparison.NewPackages = append(comparison.NewPackages, curr)
			continue
		}

		delta := curr.Coverage - prev.Coverage
		change := CoverageDelta{
			Name:     name,
			Previous: prev.Coverage,
			Current:  curr.Coverage,
			Delta:    delta,
		}
		if delta < 0 {
			comparison.Regressions = append(comparison.Regressions, change)
		} else if delta > 0 {
			comparison.Improvements = append(comparison.Improvements, change)
		}
	}

	for name, prev := range prevPackages {
		if _, ok := currPackages[name]; !ok {
			comparison.Removed = append(comparison.Removed, prev)
		}
	}

	sortCoverageComparison(&comparison)
	comparison.TotalDelta = current.Total.Coverage - previous.Total.Coverage
	return comparison
}

func snapshotFromBuckets(packages map[string]*coverageBucket, total coverageBucket) CoverageSnapshot {
	snapshot := CoverageSnapshot{
		CapturedAt: time.Now().UTC(),
		Packages:   make([]CoveragePackage, 0, len(packages)),
	}

	for name, b := range packages {
		snapshot.Packages = append(snapshot.Packages, coverageAverage(name, b.covered, b.total))
	}

	sort.Slice(snapshot.Packages, func(i, j int) bool {
		return snapshot.Packages[i].Name < snapshot.Packages[j].Name
	})

	snapshot.Total = coverageAverage("total", total.covered, total.total)
	return snapshot
}

func coverageAverage(name string, covered, total int) CoveragePackage {
	pkg := CoveragePackage{
		Name:              name,
		CoveredStatements: covered,
		TotalStatements:   total,
	}
	if total > 0 {
		pkg.Coverage = float64(covered) / float64(total) * 100
	}
	return pkg
}

func coverageMap(packages []CoveragePackage) map[string]CoveragePackage {
	result := make(map[string]CoveragePackage, len(packages))
	for _, pkg := range packages {
		result[pkg.Name] = pkg
	}
	return result
}

func sortCoverageComparison(comparison *CoverageComparison) {
	sort.Slice(comparison.Regressions, func(i, j int) bool {
		return comparison.Regressions[i].Name < comparison.Regressions[j].Name
	})
	sort.Slice(comparison.Improvements, func(i, j int) bool {
		return comparison.Improvements[i].Name < comparison.Improvements[j].Name
	})
	sort.Slice(comparison.NewPackages, func(i, j int) bool {
		return comparison.NewPackages[i].Name < comparison.NewPackages[j].Name
	})
	sort.Slice(comparison.Removed, func(i, j int) bool {
		return comparison.Removed[i].Name < comparison.Removed[j].Name
	})
}
