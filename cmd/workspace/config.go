// Package workspace provides workspace configuration for the devops CLI.
//
// It reads and writes workspace configuration from .core/workspace.yaml:
//
//	cfg, err := workspace.LoadConfig("/path/to/workspace")
//	if err == nil && cfg != nil {
//	    fmt.Println(cfg.PackagesDir)
//	}
package workspace

import (
	core "dappco.re/go"
	coreio "dappco.re/go/io"
	log "dappco.re/go/log"
	"gopkg.in/yaml.v3"
)

// WorkspaceConfig holds workspace-level configuration from .core/workspace.yaml.
type WorkspaceConfig struct {
	Version     int      `yaml:"version"`
	Active      string   `yaml:"active"`       // Active package name
	DefaultOnly []string `yaml:"default_only"` // Default types for setup
	PackagesDir string   `yaml:"packages_dir"` // Where packages are cloned
}

// DefaultConfig returns a config with sensible defaults.
func DefaultConfig() *WorkspaceConfig {
	return &WorkspaceConfig{
		Version:     1,
		PackagesDir: "./packages",
	}
}

// LoadConfig reads .core/workspace.yaml from the given directory, walking up to parent dirs.
// Returns nil (no error) if no config file is found.
func LoadConfig(dir string) (*WorkspaceConfig, core.Result) {
	absDirResult := core.PathAbs(dir)
	if !absDirResult.OK {
		return nil, core.Fail(core.Errorf("workspace.LoadConfig: resolve %q: %w", dir, absDirResult.Value.(error)))
	}

	return loadConfig(core.PathJoin(absDirResult.Value.(string)))
}

func loadConfig(dir string) (*WorkspaceConfig, core.Result) {
	path := core.PathJoin(dir, ".core", "workspace.yaml")

	if !coreio.Local.IsFile(path) {
		parent := core.PathDir(dir)
		if parent != dir {
			return loadConfig(parent)
		}
		return nil, core.Ok(nil)
	}

	data, err := coreio.Local.Read(path)
	if err != nil {
		return nil, core.Fail(core.Errorf("workspace.LoadConfig: failed to read workspace config: %w", err))
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal([]byte(data), cfg); err != nil {
		return nil, core.Fail(core.Errorf("workspace.LoadConfig: failed to parse workspace config: %w", err))
	}

	return cfg, core.Ok(nil)
}

// FindRoot searches upward for the root directory containing .core/workspace.yaml.
func FindRoot() (string, core.Result) {
	dirResult := core.Getwd()
	if !dirResult.OK {
		return "", dirResult
	}
	dir := dirResult.Value.(string)

	for {
		if coreio.Local.IsFile(core.PathJoin(dir, ".core", "workspace.yaml")) {
			return dir, core.Ok(nil)
		}
		parent := core.PathDir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", core.Fail(log.E("workspace.FindRoot", "not inside a workspace", nil))
}
