package devkit

import . "dappco.re/go"

func TestAX7_NewCoverageStore_Good(t *T) {
	path := Path(t.TempDir(), "coverage.json")
	store := NewCoverageStore(path)

	AssertNotNil(t, store)
	AssertEqual(t, path, store.path)
}

func TestAX7_NewCoverageStore_Bad(t *T) {
	store := NewCoverageStore("")
	AssertNotNil(t, store)

	AssertEqual(t, "", store.path)
	AssertError(t, store.Append(CoverageSnapshot{}))
}

func TestAX7_NewCoverageStore_Ugly(t *T) {
	path := Path(t.TempDir(), "nested", "coverage.json")
	store := NewCoverageStore(path)

	AssertNotNil(t, store)
	AssertContains(t, store.path, "nested")
}

func TestAX7_CoverageStore_Append_Good(t *T) {
	store := NewCoverageStore(Path(t.TempDir(), "coverage.json"))
	snapshot := CoverageSnapshot{CapturedAt: UnixTime(1770000000), Total: CoveragePackage{Name: "total", Coverage: 80}}

	err := store.Append(snapshot)
	AssertNoError(t, err)
	AssertTrue(t, Stat(store.path).OK)
}

func TestAX7_CoverageStore_Append_Bad(t *T) {
	dir := t.TempDir()
	store := NewCoverageStore(dir)
	err := store.Append(CoverageSnapshot{})

	AssertError(t, err)
	AssertContains(t, err.Error(), "is a directory")
}

func TestAX7_CoverageStore_Append_Ugly(t *T) {
	store := NewCoverageStore(Path(t.TempDir(), "coverage.json"))
	err := store.Append(CoverageSnapshot{})

	AssertNoError(t, err)
	AssertTrue(t, Stat(store.path).OK)
}

func TestAX7_CoverageStore_Load_Good(t *T) {
	store := NewCoverageStore(Path(t.TempDir(), "coverage.json"))
	RequireNoError(t, store.Append(CoverageSnapshot{CapturedAt: UnixTime(1)}))

	snapshots, err := store.Load()
	AssertNoError(t, err)
	AssertLen(t, snapshots, 1)
}

func TestAX7_CoverageStore_Load_Bad(t *T) {
	path := Path(t.TempDir(), "coverage.json")
	RequireTrue(t, WriteFile(path, []byte("{"), 0o600).OK)
	store := NewCoverageStore(path)

	snapshots, err := store.Load()
	AssertError(t, err)
	AssertNil(t, snapshots)
}

func TestAX7_CoverageStore_Load_Ugly(t *T) {
	path := Path(t.TempDir(), "coverage.json")
	RequireTrue(t, WriteFile(path, []byte(" \n "), 0o600).OK)
	store := NewCoverageStore(path)

	snapshots, err := store.Load()
	AssertNoError(t, err)
	AssertNil(t, snapshots)
}

func TestAX7_CoverageStore_Latest_Good(t *T) {
	store := NewCoverageStore(Path(t.TempDir(), "coverage.json"))
	RequireNoError(t, store.Append(CoverageSnapshot{CapturedAt: UnixTime(1)}))
	RequireNoError(t, store.Append(CoverageSnapshot{CapturedAt: UnixTime(2)}))

	latest, err := store.Latest()
	AssertNoError(t, err)
	AssertTrue(t, latest.CapturedAt.Equal(UnixTime(2)))
}

func TestAX7_CoverageStore_Latest_Bad(t *T) {
	store := NewCoverageStore(Path(t.TempDir(), "coverage.json"))
	latest, err := store.Latest()

	AssertError(t, err)
	AssertEqual(t, CoverageSnapshot{}, latest)
}

func TestAX7_CoverageStore_Latest_Ugly(t *T) {
	store := NewCoverageStore(Path(t.TempDir(), "coverage.json"))
	RequireNoError(t, store.Append(CoverageSnapshot{}))

	latest, err := store.Latest()
	AssertNoError(t, err)
	AssertEqual(t, CoverageSnapshot{}, latest)
}

func TestAX7_ParseCoverProfile_Good(t *T) {
	snapshot, err := ParseCoverProfile("mode: set\npkg/a.go:1.1,2.1 2 1\n")
	AssertNoError(t, err)

	AssertLen(t, snapshot.Packages, 1)
	AssertEqual(t, 100.0, snapshot.Total.Coverage)
}

func TestAX7_ParseCoverProfile_Bad(t *T) {
	snapshot, err := ParseCoverProfile("mode: set\nbroken line\n")
	AssertError(t, err)

	AssertEqual(t, CoverageSnapshot{}, snapshot)
	AssertContains(t, err.Error(), "invalid cover profile line")
}

func TestAX7_ParseCoverProfile_Ugly(t *T) {
	snapshot, err := ParseCoverProfile(" \n ")
	AssertNoError(t, err)

	AssertEmpty(t, snapshot.Packages)
	AssertEqual(t, 0.0, snapshot.Total.Coverage)
}

func TestAX7_ParseCoverOutput_Good(t *T) {
	snapshot, err := ParseCoverOutput("ok  \tpkg/a\t0.1s\tcoverage: 75.0% of statements\n")
	AssertNoError(t, err)

	AssertLen(t, snapshot.Packages, 1)
	AssertEqual(t, 75.0, snapshot.Total.Coverage)
}

func TestAX7_ParseCoverOutput_Bad(t *T) {
	snapshot, err := ParseCoverOutput("no coverage here\n")
	AssertNoError(t, err)

	AssertEmpty(t, snapshot.Packages)
	AssertEqual(t, 0.0, snapshot.Total.Coverage)
}

func TestAX7_ParseCoverOutput_Ugly(t *T) {
	snapshot, err := ParseCoverOutput("?   \tpkg/a\t0.1s\tcoverage: 0.0% of statements\n")
	AssertNoError(t, err)

	AssertLen(t, snapshot.Packages, 1)
	AssertEqual(t, 0.0, snapshot.Total.Coverage)
}

func TestAX7_CompareCoverage_Good(t *T) {
	previous := CoverageSnapshot{Packages: []CoveragePackage{{Name: "pkg/a", Coverage: 90}}}
	current := CoverageSnapshot{Packages: []CoveragePackage{{Name: "pkg/a", Coverage: 95}}}
	comparison := CompareCoverage(previous, current)

	AssertLen(t, comparison.Improvements, 1)
	AssertEqual(t, 5.0, comparison.Improvements[0].Delta)
}

func TestAX7_CompareCoverage_Bad(t *T) {
	previous := CoverageSnapshot{Packages: []CoveragePackage{{Name: "pkg/a", Coverage: 90}}}
	current := CoverageSnapshot{Packages: []CoveragePackage{{Name: "pkg/a", Coverage: 80}}}
	comparison := CompareCoverage(previous, current)

	AssertLen(t, comparison.Regressions, 1)
	AssertEqual(t, -10.0, comparison.Regressions[0].Delta)
}

func TestAX7_CompareCoverage_Ugly(t *T) {
	comparison := CompareCoverage(CoverageSnapshot{}, CoverageSnapshot{})
	AssertEmpty(t, comparison.Regressions)

	AssertEmpty(t, comparison.Improvements)
	AssertEqual(t, 0.0, comparison.TotalDelta)
}

func TestAX7_ScanSecrets_Good(t *T) {
	original := scanSecretsRunner
	t.Cleanup(func() { scanSecretsRunner = original })
	scanSecretsRunner = func(string) ([]byte, error) {
		return []byte("RuleID,File,StartLine,StartColumn,Match\ngithub-token,config.yml,2,3,ghp_exampletoken1234567890\n"), nil
	}

	findings, err := ScanSecrets("/tmp/project")
	AssertNoError(t, err)
	AssertEqual(t, "github-token", findings[0].Rule)
}

func TestAX7_ScanSecrets_Bad(t *T) {
	original := scanSecretsRunner
	t.Cleanup(func() { scanSecretsRunner = original })
	scanSecretsRunner = func(string) ([]byte, error) { return nil, AnError }

	findings, err := ScanSecrets("/tmp/project")
	AssertError(t, err)
	AssertNil(t, findings)
}

func TestAX7_ScanSecrets_Ugly(t *T) {
	original := scanSecretsRunner
	t.Cleanup(func() { scanSecretsRunner = original })
	scanSecretsRunner = func(string) ([]byte, error) { return nil, nil }

	findings, err := ScanSecrets("/tmp/project")
	AssertNoError(t, err)
	AssertNil(t, findings)
}

func TestAX7_ScanDir_Good(t *T) {
	dir := t.TempDir()
	RequireTrue(t, WriteFile(Path(dir, "config.env"), []byte("API_KEY=abcdefghijk\n"), 0o600).OK)
	findings, err := ScanDir(dir)

	AssertNoError(t, err)
	AssertEqual(t, "generic-secret-assignment", findings[0].Rule)
}

func TestAX7_ScanDir_Bad(t *T) {
	findings, err := ScanDir(Path(t.TempDir(), "missing"))
	AssertError(t, err)

	AssertNil(t, findings)
	AssertContains(t, err.Error(), "no such file")
}

func TestAX7_ScanDir_Ugly(t *T) {
	dir := t.TempDir()
	RequireTrue(t, MkdirAll(Path(dir, ".git"), 0o755).OK)
	RequireTrue(t, WriteFile(Path(dir, ".git", "secret.env"), []byte("API_KEY=abcdefghijk\n"), 0o600).OK)

	findings, err := ScanDir(dir)
	AssertNoError(t, err)
	AssertEmpty(t, findings)
}
