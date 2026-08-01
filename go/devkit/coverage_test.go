package devkit

import (
	core "dappco.re/go"
)

func TestCoverage_NewCoverageStore_Good(t *core.T) {
	path := core.Path(t.TempDir(), "coverage.json")
	store := NewCoverageStore(path)

	core.AssertNotNil(t, store)
	core.AssertEqual(t, path, store.path)
}

func TestCoverage_NewCoverageStore_Bad(t *core.T) {
	store := NewCoverageStore("")
	core.AssertNotNil(t, store)

	core.AssertEqual(t, "", store.path)
	core.AssertFalse(t, store.Append(CoverageSnapshot{}).OK)
}

func TestCoverage_NewCoverageStore_Ugly(t *core.T) {
	path := core.Path(t.TempDir(), "nested", "coverage.json")
	store := NewCoverageStore(path)

	core.AssertNotNil(t, store)
	core.AssertContains(t, store.path, "nested")
}

func TestCoverage_CoverageStore_Append_Good(t *core.T) {
	store := NewCoverageStore(core.Path(t.TempDir(), "coverage.json"))
	snapshot := CoverageSnapshot{CapturedAt: core.UnixTime(1770000000), Total: CoveragePackage{Name: "total", Coverage: 80}}

	r := store.Append(snapshot)
	core.AssertTrue(t, r.OK)
	core.AssertTrue(t, core.Stat(store.path).OK)
}

func TestCoverage_CoverageStore_Append_Bad(t *core.T) {
	dir := t.TempDir()
	store := NewCoverageStore(dir)
	r := store.Append(CoverageSnapshot{})

	core.AssertFalse(t, r.OK)
	core.AssertContains(t, r.Error(), "is a directory")
}

func TestCoverage_CoverageStore_Append_Ugly(t *core.T) {
	store := NewCoverageStore(core.Path(t.TempDir(), "coverage.json"))
	r := store.Append(CoverageSnapshot{})

	core.AssertTrue(t, r.OK)
	core.AssertTrue(t, core.Stat(store.path).OK)
}

func TestCoverage_CoverageStore_Load_Good(t *core.T) {
	store := NewCoverageStore(core.Path(t.TempDir(), "coverage.json"))
	core.RequireTrue(t, store.Append(CoverageSnapshot{CapturedAt: core.UnixTime(1)}).OK)

	snapshots, r := store.Load()
	core.AssertTrue(t, r.OK)
	core.AssertLen(t, snapshots, 1)
}

func TestCoverage_CoverageStore_Load_Bad(t *core.T) {
	path := core.Path(t.TempDir(), "coverage.json")
	core.RequireTrue(t, core.WriteFile(path, []byte("{"), 0o600).OK)
	store := NewCoverageStore(path)

	snapshots, r := store.Load()
	core.AssertFalse(t, r.OK)
	core.AssertNil(t, snapshots)
}

func TestCoverage_CoverageStore_Load_Ugly(t *core.T) {
	path := core.Path(t.TempDir(), "coverage.json")
	core.RequireTrue(t, core.WriteFile(path, []byte(" \n "), 0o600).OK)
	store := NewCoverageStore(path)

	snapshots, r := store.Load()
	core.AssertTrue(t, r.OK)
	core.AssertNil(t, snapshots)
}

func TestCoverage_CoverageStore_Latest_Good(t *core.T) {
	store := NewCoverageStore(core.Path(t.TempDir(), "coverage.json"))
	core.RequireTrue(t, store.Append(CoverageSnapshot{CapturedAt: core.UnixTime(1)}).OK)
	core.RequireTrue(t, store.Append(CoverageSnapshot{CapturedAt: core.UnixTime(2)}).OK)

	latest, r := store.Latest()
	core.AssertTrue(t, r.OK)
	core.AssertTrue(t, latest.CapturedAt.Equal(core.UnixTime(2)))
}

func TestCoverage_CoverageStore_Latest_Bad(t *core.T) {
	store := NewCoverageStore(core.Path(t.TempDir(), "coverage.json"))
	latest, r := store.Latest()

	core.AssertFalse(t, r.OK)
	core.AssertEqual(t, CoverageSnapshot{}, latest)
}

func TestCoverage_CoverageStore_Latest_Ugly(t *core.T) {
	store := NewCoverageStore(core.Path(t.TempDir(), "coverage.json"))
	core.RequireTrue(t, store.Append(CoverageSnapshot{}).OK)

	latest, r := store.Latest()
	core.AssertTrue(t, r.OK)
	core.AssertEqual(t, CoverageSnapshot{}, latest)
}

func TestCoverage_ParseCoverProfile_Good(t *core.T) {
	snapshot, r := ParseCoverProfile("mode: set\npkg/a.go:1.1,2.1 2 1\n")
	core.AssertTrue(t, r.OK)

	core.AssertLen(t, snapshot.Packages, 1)
	core.AssertEqual(t, 100.0, snapshot.Total.Coverage)
}

func TestCoverage_ParseCoverProfile_Bad(t *core.T) {
	snapshot, r := ParseCoverProfile("mode: set\nbroken line\n")
	core.AssertFalse(t, r.OK)

	core.AssertEqual(t, CoverageSnapshot{}, snapshot)
	core.AssertContains(t, r.Error(), "invalid cover profile line")
}

func TestCoverage_ParseCoverProfile_Ugly(t *core.T) {
	snapshot, r := ParseCoverProfile(" \n ")
	core.AssertTrue(t, r.OK)

	core.AssertEmpty(t, snapshot.Packages)
	core.AssertEqual(t, 0.0, snapshot.Total.Coverage)
}

func TestCoverage_ParseCoverOutput_Good(t *core.T) {
	snapshot, r := ParseCoverOutput("ok  \tpkg/a\t0.1s\tcoverage: 75.0% of statements\n")
	core.AssertTrue(t, r.OK)

	core.AssertLen(t, snapshot.Packages, 1)
	core.AssertEqual(t, 75.0, snapshot.Total.Coverage)
}

func TestCoverage_ParseCoverOutput_Bad(t *core.T) {
	snapshot, r := ParseCoverOutput("no coverage here\n")
	core.AssertTrue(t, r.OK)

	core.AssertEmpty(t, snapshot.Packages)
	core.AssertEqual(t, 0.0, snapshot.Total.Coverage)
}

func TestCoverage_ParseCoverOutput_Ugly(t *core.T) {
	snapshot, r := ParseCoverOutput("?   \tpkg/a\t0.1s\tcoverage: 0.0% of statements\n")
	core.AssertTrue(t, r.OK)

	core.AssertLen(t, snapshot.Packages, 1)
	core.AssertEqual(t, 0.0, snapshot.Total.Coverage)
}

func TestCoverage_CompareCoverage_Good(t *core.T) {
	previous := CoverageSnapshot{Packages: []CoveragePackage{{Name: "pkg/a", Coverage: 90}}}
	current := CoverageSnapshot{Packages: []CoveragePackage{{Name: "pkg/a", Coverage: 95}}}
	comparison := CompareCoverage(previous, current)

	core.AssertLen(t, comparison.Improvements, 1)
	core.AssertEqual(t, 5.0, comparison.Improvements[0].Delta)
}

func TestCoverage_CompareCoverage_Bad(t *core.T) {
	previous := CoverageSnapshot{Packages: []CoveragePackage{{Name: "pkg/a", Coverage: 90}}}
	current := CoverageSnapshot{Packages: []CoveragePackage{{Name: "pkg/a", Coverage: 80}}}
	comparison := CompareCoverage(previous, current)

	core.AssertLen(t, comparison.Regressions, 1)
	core.AssertEqual(t, -10.0, comparison.Regressions[0].Delta)
}

func TestCoverage_CompareCoverage_Ugly(t *core.T) {
	comparison := CompareCoverage(CoverageSnapshot{}, CoverageSnapshot{})
	core.AssertEmpty(t, comparison.Regressions)

	core.AssertEmpty(t, comparison.Improvements)
	core.AssertEqual(t, 0.0, comparison.TotalDelta)
}
