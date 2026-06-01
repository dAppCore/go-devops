package workspace

import (
	. "dappco.re/go"
)

// TestConfig_LoadConfig_Malformed covers the YAML parse-error path of
// loadConfig — a workspace.yaml that exists but is not valid YAML.
func TestConfig_LoadConfig_Malformed(t *T) {
	dir := t.TempDir()
	RequireTrue(t, MkdirAll(Path(dir, ".core"), 0o755).OK)
	RequireTrue(t, WriteFile(Path(dir, ".core", "workspace.yaml"), []byte("version: [unterminated\n"), 0o644).OK)

	cfg, r := LoadConfig(dir)
	AssertFalse(t, r.OK)
	AssertNil(t, cfg)
	AssertContains(t, r.Error(), "failed to parse workspace config")
}

// TestConfig_LoadConfig_MalformedNested verifies the parse error is
// surfaced even when the malformed file is discovered via the upward
// directory walk rather than directly in the requested directory.
func TestConfig_LoadConfig_MalformedNested(t *T) {
	dir := t.TempDir()
	child := Path(dir, "x", "y")
	RequireTrue(t, MkdirAll(Path(dir, ".core"), 0o755).OK)
	RequireTrue(t, MkdirAll(child, 0o755).OK)
	RequireTrue(t, WriteFile(Path(dir, ".core", "workspace.yaml"), []byte("active: {bad: : :}\n"), 0o644).OK)

	cfg, r := LoadConfig(child)
	AssertFalse(t, r.OK)
	AssertNil(t, cfg)
	AssertContains(t, r.Error(), "failed to parse workspace config")
}
