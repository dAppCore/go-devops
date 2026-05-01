package devkit

import . "dappco.re/go"

func ExampleScanDir() {
	dir := MustCast[string](MkdirTemp("", "secret-scan-*"))
	defer RemoveAll(dir)
	WriteFile(PathJoin(dir, "config.env"), []byte("API_KEY=abcdefghijk\n"), 0o600)
	findings, r := ScanDir(dir)
	Println(r.OK, findings[0].Rule)
	// Output: true generic-secret-assignment
}
