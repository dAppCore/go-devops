package devkit

import (
	"math"
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
	if err != nil {
		t.Fatalf("parse cover profile: %v", err)
	}
	if len(snapshot.Packages) != 2 {
		t.Fatalf("packages length = %d, want 2", len(snapshot.Packages))
	}
	if snapshot.Packages[0].Name != "github.com/acme/project/baz" {
		t.Fatalf("packages[0].Name = %q, want github.com/acme/project/baz", snapshot.Packages[0].Name)
	}
	if snapshot.Packages[1].Name != "github.com/acme/project/foo" {
		t.Fatalf("packages[1].Name = %q, want github.com/acme/project/foo", snapshot.Packages[1].Name)
	}
	for name, check := range map[string]struct {
		got  float64
		want float64
	}{
		"baz coverage":   {got: snapshot.Packages[0].Coverage, want: 100.0},
		"foo coverage":   {got: snapshot.Packages[1].Coverage, want: 40.0},
		"total coverage": {got: snapshot.Total.Coverage, want: 66.6667},
	} {
		if math.Abs(check.got-check.want) > 0.0001 {
			t.Fatalf("%s = %v, want %v", name, check.got, check.want)
		}
	}
}

func TestParseCoverProfile_Bad(t *testing.T) {
	_, err := ParseCoverProfile("mode: set\nbroken line")
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestParseCoverOutput_Good(t *testing.T) {
	snapshot, err := ParseCoverOutput(`ok  	github.com/acme/project/foo	0.123s	coverage: 75.0% of statements
ok  	github.com/acme/project/bar	0.456s	coverage: 50.0% of statements
`)
	if err != nil {
		t.Fatalf("parse cover output: %v", err)
	}
	if len(snapshot.Packages) != 2 {
		t.Fatalf("packages length = %d, want 2", len(snapshot.Packages))
	}
	if snapshot.Packages[0].Name != "github.com/acme/project/bar" {
		t.Fatalf("packages[0].Name = %q, want github.com/acme/project/bar", snapshot.Packages[0].Name)
	}
	if snapshot.Packages[1].Name != "github.com/acme/project/foo" {
		t.Fatalf("packages[1].Name = %q, want github.com/acme/project/foo", snapshot.Packages[1].Name)
	}
	if math.Abs(snapshot.Total.Coverage-62.5) > 0.0001 {
		t.Fatalf("total coverage = %v, want 62.5", snapshot.Total.Coverage)
	}
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
	if len(comparison.Regressions) != 1 {
		t.Fatalf("regressions length = %d, want 1", len(comparison.Regressions))
	}
	if len(comparison.Improvements) != 1 {
		t.Fatalf("improvements length = %d, want 1", len(comparison.Improvements))
	}
	if len(comparison.NewPackages) != 1 {
		t.Fatalf("new packages length = %d, want 1", len(comparison.NewPackages))
	}
	if len(comparison.Removed) != 0 {
		t.Fatalf("removed length = %d, want 0", len(comparison.Removed))
	}
	if comparison.Regressions[0].Name != "pkg/a" {
		t.Fatalf("regressions[0].Name = %q, want pkg/a", comparison.Regressions[0].Name)
	}
	if comparison.Improvements[0].Name != "pkg/b" {
		t.Fatalf("improvements[0].Name = %q, want pkg/b", comparison.Improvements[0].Name)
	}
	if comparison.NewPackages[0].Name != "pkg/c" {
		t.Fatalf("newPackages[0].Name = %q, want pkg/c", comparison.NewPackages[0].Name)
	}
	if math.Abs(comparison.TotalDelta-4.0) > 0.0001 {
		t.Fatalf("total delta = %v, want 4.0", comparison.TotalDelta)
	}
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

	if err := store.Append(first); err != nil {
		t.Fatalf("append first snapshot: %v", err)
	}
	if err := store.Append(second); err != nil {
		t.Fatalf("append second snapshot: %v", err)
	}

	snapshots, err := store.Load()
	if err != nil {
		t.Fatalf("load snapshots: %v", err)
	}
	if len(snapshots) != 2 {
		t.Fatalf("snapshots length = %d, want 2", len(snapshots))
	}
	if !snapshots[0].CapturedAt.Equal(first.CapturedAt) {
		t.Fatalf("snapshots[0].CapturedAt = %v, want %v", snapshots[0].CapturedAt, first.CapturedAt)
	}
	if !snapshots[1].CapturedAt.Equal(second.CapturedAt) {
		t.Fatalf("snapshots[1].CapturedAt = %v, want %v", snapshots[1].CapturedAt, second.CapturedAt)
	}

	latest, err := store.Latest()
	if err != nil {
		t.Fatalf("load latest snapshot: %v", err)
	}
	if !latest.CapturedAt.Equal(second.CapturedAt) {
		t.Fatalf("latest.CapturedAt = %v, want %v", latest.CapturedAt, second.CapturedAt)
	}
}

func TestCoverageStore_Bad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "coverage.json")
	if err := os.WriteFile(path, []byte("{"), 0o600); err != nil {
		t.Fatalf("write coverage store: %v", err)
	}

	store := NewCoverageStore(path)
	_, err := store.Load()
	if err == nil {
		t.Fatal("expected load error")
	}
}
