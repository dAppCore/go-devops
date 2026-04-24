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
	"os"
	"path/filepath"

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
func LoadConfig(dir string) (*WorkspaceConfig, error) {
	path := filepath.Join(dir, ".core", "workspace.yaml")

	if !coreio.Local.IsFile(path) {
		parent := filepath.Dir(dir)
		if parent != dir {
			return LoadConfig(parent)
		}
		return nil, nil
	}

	data, err := coreio.Local.Read(path)
	if err != nil {
		return nil, log.E("workspace.LoadConfig", "failed to read workspace config", err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal([]byte(data), cfg); err != nil {
		return nil, log.E("workspace.LoadConfig", "failed to parse workspace config", err)
	}

	return cfg, nil
}

// FindRoot searches upward for the root directory containing .core/workspace.yaml.
func FindRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if coreio.Local.IsFile(filepath.Join(dir, ".core", "workspace.yaml")) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", log.E("workspace.FindRoot", "not inside a workspace", nil)
}
