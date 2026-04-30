package dev

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
	"dappco.re/go/devops/cmd/workspace"
	"dappco.re/go/i18n"
	"dappco.re/go/io"
	"dappco.re/go/scm/repos"
)

// loadRegistryWithConfig loads the registry and applies workspace configuration.
func loadRegistryWithConfig(registryPath string) (*repos.Registry, string, core.Result) {
	var reg *repos.Registry
	var err error
	var registryDir string

	if registryPath != "" {
		reg, err = repos.LoadRegistry(io.Local, registryPath)
		if err != nil {
			return nil, "", core.Fail(cli.Wrap(err, "failed to load registry"))
		}
		cli.Print("%s %s\n\n", dimStyle.Render(i18n.Label("registry")), registryPath)
		registryDir = core.PathDir(registryPath)
	} else {
		registryPath, err = repos.FindRegistry(io.Local)
		if err == nil {
			reg, err = repos.LoadRegistry(io.Local, registryPath)
			if err != nil {
				return nil, "", core.Fail(cli.Wrap(err, "failed to load registry"))
			}
			cli.Print("%s %s\n\n", dimStyle.Render(i18n.Label("registry")), registryPath)
			registryDir = core.PathDir(registryPath)
		} else {
			// Fallback: scan current directory
			cwd := "."
			if cwdResult := core.Getwd(); cwdResult.OK {
				cwd = cwdResult.Value.(string)
			}
			reg, err = repos.ScanDirectory(io.Local, cwd)
			if err != nil {
				return nil, "", core.Fail(cli.Wrap(err, "failed to scan directory"))
			}
			cli.Print("%s %s\n\n", dimStyle.Render(i18n.T("cmd.dev.scanning_label")), cwd)
			registryDir = cwd
		}
	}
	// Load workspace config to respect packages_dir (only if config exists)
	if wsConfig, r := workspace.LoadConfig(registryDir); r.OK && wsConfig != nil {
		if wsConfig.PackagesDir != "" {
			pkgDir := wsConfig.PackagesDir
			// Expand ~
			if core.HasPrefix(pkgDir, "~/") {
				if homeResult := core.UserHomeDir(); homeResult.OK {
					pkgDir = core.PathJoin(homeResult.Value.(string), pkgDir[2:])
				}
			}
			if !core.PathIsAbs(pkgDir) {
				pkgDir = core.PathJoin(registryDir, pkgDir)
			}

			// Update repo paths
			for _, repo := range reg.Repos {
				repo.Path = core.PathJoin(pkgDir, repo.Name)
			}
		}
	}

	return reg, registryDir, core.Ok(nil)
}
