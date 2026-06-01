package locales

import (
	core "dappco.re/go"
)

// Named import rather than the house `. "dappco.re/go"` dot-import: this
// package declares a package-level `FS` symbol that collides with the
// core.FS type re-exported by the dot-import. A named import keeps the
// embedded FS reachable while still using the core test DSL.

// TestEmbed_FS_Good verifies the embedded locale filesystem carries the
// canonical en.json catalogue that init() registers with the i18n hub.
func TestEmbed_FS_Good(t *core.T) {
	data, err := FS.ReadFile("en.json")
	core.RequireTrue(t, err == nil)
	core.AssertTrue(t, len(data) > 0)
}

// TestEmbed_FS_Bad confirms a non-existent locale is reported as missing
// rather than silently returning empty bytes.
func TestEmbed_FS_Bad(t *core.T) {
	_, err := FS.ReadFile("nonexistent-locale.json")
	core.AssertTrue(t, err != nil)
}

// TestEmbed_FS_Ugly parses every embedded *.json catalogue to guard
// against a malformed translation file shipping in the binary.
func TestEmbed_FS_Ugly(t *core.T) {
	entries, err := FS.ReadDir(".")
	core.RequireTrue(t, err == nil)

	seen := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, readErr := FS.ReadFile(e.Name())
		core.RequireTrue(t, readErr == nil)

		var catalogue map[string]any
		r := core.JSONUnmarshal(data, &catalogue)
		core.AssertTrue(t, r.OK, e.Name())
		core.AssertTrue(t, len(catalogue) > 0, e.Name())
		seen++
	}
	core.AssertTrue(t, seen > 0)
}
