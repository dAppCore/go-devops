// LEK-1 | lthn.ai | EUPL-1.2
package devkit

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleCoverProfile = `mode: set
example.com/pkg1/file1.go:10.1,20.2 5 1
example.com/pkg1/file1.go:22.1,30.2 3 0
example.com/pkg1/file2.go:5.1,15.2 4 1
example.com/pkg2/main.go:1.1,10.2 10 1
example.com/pkg2/main.go:12.1,20.2 10 0
`

func TestParseCoverProfile_Good(t *testing.T) {
	snap, err := ParseCoverProfile(sampleCoverProfile)
	require.NoError(t, err)

	assert.Len(t, snap.Packages, 2)

	// pkg1: 5+4 covered out of 5+3+4=12 => 9/12 = 75%
	assert.Equal(t, 75.0, snap.Packages["example.com/pkg1"])

	// pkg2: 10 covered out of 10+10=20 => 10/20 = 50%
	assert.Equal(t, 50.0, snap.Packages["example.com/pkg2"])

	// Total: 19/32 = 59.4%
	assert.Equal(t, 59.4, snap.Total)
	assert.False(t, snap.Timestamp.IsZero())
}

func TestParseCoverProfile_Empty_Good(t *testing.T) {
	snap, err := ParseCoverProfile("")
	require.NoError(t, err)
	assert.Empty(t, snap.Packages)
	assert.Equal(t, 0.0, snap.Total)
}

func TestParseCoverProfile_ModeOnly_Good(t *testing.T) {
	snap, err := ParseCoverProfile("mode: set\n")
	require.NoError(t, err)
	assert.Empty(t, snap.Packages)
}

func TestParseCoverProfile_FullCoverage_Good(t *testing.T) {
	data := `mode: set
example.com/perfect/main.go:1.1,10.2 10 1
example.com/perfect/main.go:12.1,20.2 5 1
`
	snap, err := ParseCoverProfile(data)
	require.NoError(t, err)
	assert.Equal(t, 100.0, snap.Packages["example.com/perfect"])
	assert.Equal(t, 100.0, snap.Total)
}

func TestParseCoverProfile_ZeroCoverage_Good(t *testing.T) {
	data := `mode: set
example.com/zero/main.go:1.1,10.2 10 0
example.com/zero/main.go:12.1,20.2 5 0
`
	snap, err := ParseCoverProfile(data)
	require.NoError(t, err)
	assert.Equal(t, 0.0, snap.Packages["example.com/zero"])
	assert.Equal(t, 0.0, snap.Total)
}

func TestParseCoverProfile_MalformedLines_Bad(t *testing.T) {
	data := `mode: set
not a valid line
example.com/pkg/file.go:1.1,10.2 5 1
another bad line with spaces
example.com/pkg/file.go:12.1,20.2 5 0
`
	snap, err := ParseCoverProfile(data)
	require.NoError(t, err)
	assert.Len(t, snap.Packages, 1)
	assert.Equal(t, 50.0, snap.Packages["example.com/pkg"])
}

func TestParseCoverOutput_Good(t *testing.T) {
	output := `?   	example.com/skipped	[no test files]
ok  	example.com/pkg1	0.5s	coverage: 85.0% of statements
ok  	example.com/pkg2	0.2s	coverage: 42.5% of statements
ok  	example.com/pkg3	1.1s	coverage: 100.0% of statements
`
	snap, err := ParseCoverOutput(output)
	require.NoError(t, err)

	assert.Len(t, snap.Packages, 3)
	assert.Equal(t, 85.0, snap.Packages["example.com/pkg1"])
	assert.Equal(t, 42.5, snap.Packages["example.com/pkg2"])
	assert.Equal(t, 100.0, snap.Packages["example.com/pkg3"])

	// Total = avg of (85 + 42.5 + 100) / 3 = 75.8333... rounded to 75.8
	assert.Equal(t, 75.8, snap.Total)
}

func TestParseCoverOutput_Empty_Good(t *testing.T) {
	snap, err := ParseCoverOutput("")
	require.NoError(t, err)
	assert.Empty(t, snap.Packages)
	assert.Equal(t, 0.0, snap.Total)
}

func TestParseCoverOutput_NoTestFiles_Good(t *testing.T) {
	output := `?   	example.com/empty	[no test files]
`
	snap, err := ParseCoverOutput(output)
	require.NoError(t, err)
	assert.Empty(t, snap.Packages)
}

// --- CompareCoverage tests ---

func TestCompareCoverage_Regression_Good(t *testing.T) {
	prev := CoverageSnapshot{
		Packages: map[string]float64{
			"pkg1": 90.0,
			"pkg2": 85.0,
			"pkg3": 70.0,
		},
		Total: 81.7,
	}
	curr := CoverageSnapshot{
		Packages: map[string]float64{
			"pkg1": 90.0, // unchanged
			"pkg2": 75.0, // regression: -10
			"pkg3": 80.0, // improvement: +10
		},
		Total: 81.7,
	}

	comp := CompareCoverage(prev, curr)

	assert.Len(t, comp.Regressions, 1)
	assert.Equal(t, "pkg2", comp.Regressions[0].Package)
	assert.Equal(t, -10.0, comp.Regressions[0].Delta)
	assert.Equal(t, 85.0, comp.Regressions[0].Previous)
	assert.Equal(t, 75.0, comp.Regressions[0].Current)

	assert.Len(t, comp.Improvements, 1)
	assert.Equal(t, "pkg3", comp.Improvements[0].Package)
	assert.Equal(t, 10.0, comp.Improvements[0].Delta)
}

func TestCompareCoverage_NewAndRemoved_Good(t *testing.T) {
	prev := CoverageSnapshot{
		Packages: map[string]float64{
			"old-pkg": 50.0,
			"stable":  80.0,
		},
		Total: 65.0,
	}
	curr := CoverageSnapshot{
		Packages: map[string]float64{
			"stable":  80.0,
			"new-pkg": 60.0,
		},
		Total: 70.0,
	}

	comp := CompareCoverage(prev, curr)

	assert.Contains(t, comp.NewPackages, "new-pkg")
	assert.Contains(t, comp.Removed, "old-pkg")
	assert.Equal(t, 5.0, comp.TotalDelta)
	assert.Empty(t, comp.Regressions)
}

func TestCompareCoverage_Identical_Good(t *testing.T) {
	snap := CoverageSnapshot{
		Packages: map[string]float64{
			"pkg1": 80.0,
			"pkg2": 90.0,
		},
		Total: 85.0,
	}

	comp := CompareCoverage(snap, snap)

	assert.Empty(t, comp.Regressions)
	assert.Empty(t, comp.Improvements)
	assert.Empty(t, comp.NewPackages)
	assert.Empty(t, comp.Removed)
	assert.Equal(t, 0.0, comp.TotalDelta)
}

func TestCompareCoverage_EmptySnapshots_Good(t *testing.T) {
	prev := CoverageSnapshot{Packages: map[string]float64{}}
	curr := CoverageSnapshot{Packages: map[string]float64{}}

	comp := CompareCoverage(prev, curr)
	assert.Empty(t, comp.Regressions)
	assert.Empty(t, comp.Improvements)
	assert.Empty(t, comp.NewPackages)
	assert.Empty(t, comp.Removed)
}

func TestCompareCoverage_AllNew_Good(t *testing.T) {
	prev := CoverageSnapshot{Packages: map[string]float64{}}
	curr := CoverageSnapshot{
		Packages: map[string]float64{
			"new1": 50.0,
			"new2": 75.0,
		},
		Total: 62.5,
	}

	comp := CompareCoverage(prev, curr)
	assert.Len(t, comp.NewPackages, 2)
	assert.Empty(t, comp.Regressions)
	assert.Equal(t, 62.5, comp.TotalDelta)
}

// --- CoverageStore tests ---

func TestCoverageStore_AppendAndLoad_Good(t *testing.T) {
	path := filepath.Join(t.TempDir(), "coverage.json")
	store := NewCoverageStore(path)

	snap1 := CoverageSnapshot{
		Timestamp: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		Packages:  map[string]float64{"pkg1": 80.0},
		Total:     80.0,
	}
	snap2 := CoverageSnapshot{
		Timestamp: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
		Packages:  map[string]float64{"pkg1": 85.0},
		Total:     85.0,
	}

	require.NoError(t, store.Append(snap1))
	require.NoError(t, store.Append(snap2))

	loaded, err := store.Load()
	require.NoError(t, err)
	assert.Len(t, loaded, 2)
	assert.Equal(t, 80.0, loaded[0].Total)
	assert.Equal(t, 85.0, loaded[1].Total)
}

func TestCoverageStore_Latest_Good(t *testing.T) {
	path := filepath.Join(t.TempDir(), "coverage.json")
	store := NewCoverageStore(path)

	// Add snapshots out of chronological order
	snap1 := CoverageSnapshot{
		Timestamp: time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC),
		Packages:  map[string]float64{"pkg1": 90.0},
		Total:     90.0,
	}
	snap2 := CoverageSnapshot{
		Timestamp: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		Packages:  map[string]float64{"pkg1": 70.0},
		Total:     70.0,
	}

	require.NoError(t, store.Append(snap2)) // older first
	require.NoError(t, store.Append(snap1)) // newer second

	latest, err := store.Latest()
	require.NoError(t, err)
	require.NotNil(t, latest)
	assert.Equal(t, 90.0, latest.Total)
}

func TestCoverageStore_LatestEmpty_Good(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.json")
	store := NewCoverageStore(path)

	latest, err := store.Latest()
	require.NoError(t, err)
	assert.Nil(t, latest)
}

func TestCoverageStore_LoadNonexistent_Bad(t *testing.T) {
	store := NewCoverageStore("/nonexistent/path/coverage.json")
	_, err := store.Load()
	assert.Error(t, err)
}

func TestCoverageStore_LoadCorrupt_Bad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "corrupt.json")
	require.NoError(t, os.WriteFile(path, []byte("not json!!!"), 0644))

	store := NewCoverageStore(path)
	_, err := store.Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse")
}

func TestCoverageStore_WithMeta_Good(t *testing.T) {
	path := filepath.Join(t.TempDir(), "coverage.json")
	store := NewCoverageStore(path)

	snap := CoverageSnapshot{
		Timestamp: time.Now(),
		Packages:  map[string]float64{"pkg1": 80.0},
		Total:     80.0,
		Meta: map[string]string{
			"commit": "abc123",
			"branch": "main",
		},
	}

	require.NoError(t, store.Append(snap))

	loaded, err := store.Load()
	require.NoError(t, err)
	require.Len(t, loaded, 1)
	assert.Equal(t, "abc123", loaded[0].Meta["commit"])
	assert.Equal(t, "main", loaded[0].Meta["branch"])
}

func TestCoverageStore_Persistence_Good(t *testing.T) {
	path := filepath.Join(t.TempDir(), "persist.json")

	// Write with one store instance
	store1 := NewCoverageStore(path)
	snap := CoverageSnapshot{
		Timestamp: time.Now(),
		Packages:  map[string]float64{"pkg1": 55.5},
		Total:     55.5,
	}
	require.NoError(t, store1.Append(snap))

	// Read with a different store instance
	store2 := NewCoverageStore(path)
	loaded, err := store2.Load()
	require.NoError(t, err)
	assert.Len(t, loaded, 1)
	assert.Equal(t, 55.5, loaded[0].Total)
}

func TestCompareCoverage_SmallDelta_Good(t *testing.T) {
	// Test that very small deltas (<0.05) round to 0 and are not flagged.
	prev := CoverageSnapshot{
		Packages: map[string]float64{"pkg1": 80.01},
		Total:    80.01,
	}
	curr := CoverageSnapshot{
		Packages: map[string]float64{"pkg1": 80.04},
		Total:    80.04,
	}

	comp := CompareCoverage(prev, curr)
	assert.Empty(t, comp.Regressions)
	assert.Empty(t, comp.Improvements) // 0.03 rounds to 0.0
}

// LEK-1 | lthn.ai | EUPL-1.2
