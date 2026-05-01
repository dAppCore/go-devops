package devkit

import . "dappco.re/go"

func ExampleNewCoverageStore() {
	store := NewCoverageStore("coverage.json")
	Println(store != nil)
	// Output: true
}

func ExampleCoverageStore_Append() {
	dir := MustCast[string](MkdirTemp("", "coverage-store-*"))
	defer RemoveAll(dir)
	store := NewCoverageStore(PathJoin(dir, "coverage.json"))
	r := store.Append(CoverageSnapshot{Total: CoveragePackage{Name: "total", Coverage: 80}})
	Println(r.OK)
	// Output: true
}

func ExampleCoverageStore_Load() {
	dir := MustCast[string](MkdirTemp("", "coverage-load-*"))
	defer RemoveAll(dir)
	store := NewCoverageStore(PathJoin(dir, "coverage.json"))
	store.Append(CoverageSnapshot{Total: CoveragePackage{Name: "total", Coverage: 80}})
	snapshots, r := store.Load()
	Println(r.OK, len(snapshots))
	// Output: true 1
}

func ExampleCoverageStore_Latest() {
	dir := MustCast[string](MkdirTemp("", "coverage-latest-*"))
	defer RemoveAll(dir)
	store := NewCoverageStore(PathJoin(dir, "coverage.json"))
	store.Append(CoverageSnapshot{Total: CoveragePackage{Name: "total", Coverage: 80}})
	latest, r := store.Latest()
	Println(r.OK, latest.Total.Coverage)
	// Output: true 80
}

func ExampleParseCoverProfile() {
	snapshot, r := ParseCoverProfile("mode: set\npkg/a.go:1.1,2.1 2 1\n")
	Println(r.OK, snapshot.Total.Coverage)
	// Output: true 100
}

func ExampleParseCoverOutput() {
	snapshot, r := ParseCoverOutput("ok  \tpkg/a\t0.1s\tcoverage: 75.0% of statements\n")
	Println(r.OK, snapshot.Total.Coverage)
	// Output: true 75
}

func ExampleCompareCoverage() {
	previous := CoverageSnapshot{Total: CoveragePackage{Name: "total", Coverage: 90}, Packages: []CoveragePackage{{Name: "pkg/a", Coverage: 90}}}
	current := CoverageSnapshot{Total: CoveragePackage{Name: "total", Coverage: 80}, Packages: []CoveragePackage{{Name: "pkg/a", Coverage: 80}}}
	comparison := CompareCoverage(previous, current)
	Println(len(comparison.Regressions), comparison.TotalDelta)
	// Output: 1 -10
}
