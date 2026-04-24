package devkit

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseCoverProfile_Good(t *testing.T) {
	snapshot, err := ParseCoverProfile(`mode: set
github.com/acme/project/foo/foo.go:1.1,3.1 2 1
github.com/acme/project/foo/bar.go:1.1,4.1 3 0
github.com/acme/project/baz/baz.go:1.1,2.1 4 4
`)
	mustNoError(t, err)
	mustLen(t, snapshot.Packages, 2)
	mustEqual(t, "github.com/acme/project/baz", snapshot.Packages[0].Name)
	mustEqual(t, "github.com/acme/project/foo", snapshot.Packages[1].Name)
	mustInDelta(t, 100.0, snapshot.Packages[0].Coverage, 0.0001)
	mustInDelta(t, 40.0, snapshot.Packages[1].Coverage, 0.0001)
	mustInDelta(t, 66.6667, snapshot.Total.Coverage, 0.0001)
}

func TestParseCoverProfile_Bad(t *testing.T) {
	_, err := ParseCoverProfile("mode: set\nbroken line")
	mustError(t, err)
}

func TestParseCoverOutput_Good(t *testing.T) {
	snapshot, err := ParseCoverOutput(`ok  	github.com/acme/project/foo	0.123s	coverage: 75.0% of statements
ok  	github.com/acme/project/bar	0.456s	coverage: 50.0% of statements
`)
	mustNoError(t, err)
	mustLen(t, snapshot.Packages, 2)
	mustEqual(t, "github.com/acme/project/bar", snapshot.Packages[0].Name)
	mustEqual(t, "github.com/acme/project/foo", snapshot.Packages[1].Name)
	mustInDelta(t, 62.5, snapshot.Total.Coverage, 0.0001)
}

func TestCompareCoverage_Good(t *testing.T) {
	previous := CoverageSnapshot{
		Packages: []CoveragePackage{
			{Name: "pkg/a", Coverage: 90.0},
			{Name: "pkg/b", Coverage: 80.0},
		},
		Total: CoveragePackage{Name: "total", Coverage: 85.0},
	}
	current := CoverageSnapshot{
		Packages: []CoveragePackage{
			{Name: "pkg/a", Coverage: 87.5},
			{Name: "pkg/b", Coverage: 82.0},
			{Name: "pkg/c", Coverage: 100.0},
		},
		Total: CoveragePackage{Name: "total", Coverage: 89.0},
	}

	comparison := CompareCoverage(previous, current)
	mustLen(t, comparison.Regressions, 1)
	mustLen(t, comparison.Improvements, 1)
	mustLen(t, comparison.NewPackages, 1)
	mustEmpty(t, comparison.Removed)
	mustEqual(t, "pkg/a", comparison.Regressions[0].Name)
	mustEqual(t, "pkg/b", comparison.Improvements[0].Name)
	mustEqual(t, "pkg/c", comparison.NewPackages[0].Name)
	mustInDelta(t, 4.0, comparison.TotalDelta, 0.0001)
}

func TestCoverageStore_Good(t *testing.T) {
	dir := t.TempDir()
	store := NewCoverageStore(filepath.Join(dir, "coverage.json"))

	first := CoverageSnapshot{
		CapturedAt: time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC),
		Packages:   []CoveragePackage{{Name: "pkg/a", Coverage: 80.0}},
		Total:      CoveragePackage{Name: "total", Coverage: 80.0},
	}
	second := CoverageSnapshot{
		CapturedAt: time.Date(2026, 4, 1, 11, 0, 0, 0, time.UTC),
		Packages:   []CoveragePackage{{Name: "pkg/a", Coverage: 82.5}},
		Total:      CoveragePackage{Name: "total", Coverage: 82.5},
	}

	mustNoError(t, store.Append(first))
	mustNoError(t, store.Append(second))

	snapshots, err := store.Load()
	mustNoError(t, err)
	mustLen(t, snapshots, 2)
	mustEqual(t, first.CapturedAt, snapshots[0].CapturedAt)
	mustEqual(t, second.CapturedAt, snapshots[1].CapturedAt)

	latest, err := store.Latest()
	mustNoError(t, err)
	mustEqual(t, second.CapturedAt, latest.CapturedAt)
}

func TestCoverageStore_Bad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "coverage.json")
	mustNoError(t, os.WriteFile(path, []byte("{"), 0o600))

	store := NewCoverageStore(path)
	_, err := store.Load()
	mustError(t, err)
}
