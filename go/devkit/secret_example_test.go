package devkit

import core "dappco.re/go"

func ExampleScanDir() {
	dir := core.MustCast[string](core.MkdirTemp("", "secret-scan-*"))
	defer core.RemoveAll(dir)
	core.WriteFile(core.PathJoin(dir, "config.env"), []byte("API_KEY=abcdefghijk\n"), 0o600)
	findings, r := ScanDir(dir)
	core.Println(r.OK, findings[0].Rule)
	// Output: true generic-secret-assignment
}
