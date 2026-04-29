package docs

import (
	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
	"dappco.re/go/devops/cmd/workspace"
	"dappco.re/go/i18n"
	"dappco.re/go/io"
	"dappco.re/go/scm/repos"
)

// RepoDocInfo holds documentation info for a repo
type RepoDocInfo struct {
	Name      string
	Path      string
	HasDocs   bool
	Readme    string
	ClaudeMd  string
	Changelog string
	DocsFiles []string // All files in docs/ directory (recursive)
	KBFiles   []string // All files in KB/ directory (recursive)
}

func loadRegistry(registryPath string) (*repos.Registry, string, coreFailure) {
	var reg *repos.Registry
	var err error
	var registryDir string

	if registryPath != "" {
		reg, err = repos.LoadRegistry(io.Local, registryPath)
		if err != nil {
			return nil, "", cli.Wrap(err, i18n.T("i18n.fail.load", "registry"))
		}
		registryDir = core.PathDir(registryPath)
	} else {
		registryPath, err = repos.FindRegistry(io.Local)
		if err == nil {
			reg, err = repos.LoadRegistry(io.Local, registryPath)
			if err != nil {
				return nil, "", cli.Wrap(err, i18n.T("i18n.fail.load", "registry"))
			}
			registryDir = core.PathDir(registryPath)
		} else {
			cwd := "."
			if cwdResult := core.Getwd(); cwdResult.OK {
				cwd = cwdResult.Value.(string)
			}
			reg, err = repos.ScanDirectory(io.Local, cwd)
			if err != nil {
				return nil, "", cli.Wrap(err, i18n.T("i18n.fail.scan", "directory"))
			}
			registryDir = cwd
		}
	}

	// Load workspace config to respect packages_dir
	wsConfig, err := workspace.LoadConfig(registryDir)
	if err != nil {
		return nil, "", cli.Wrap(err, i18n.T("i18n.fail.load", "workspace config"))
	}

	basePath := registryDir

	if wsConfig != nil && wsConfig.PackagesDir != "" && wsConfig.PackagesDir != "./packages" {
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
		basePath = pkgDir

		// Update repo paths if they were relative to registry
		// This ensures consistency when packages_dir overrides the default
		reg.BasePath = basePath
		for _, repo := range reg.Repos {
			repo.Path = core.PathJoin(basePath, repo.Name)
		}
	}

	return reg, basePath, nil
}

func scanRepoDocs(repo *repos.Repo) RepoDocInfo {
	info := RepoDocInfo{
		Name: repo.Name,
		Path: repo.Path,
	}

	// Check for README.md
	readme := core.PathJoin(repo.Path, "README.md")
	if io.Local.IsFile(readme) {
		info.Readme = readme
		info.HasDocs = true
	}

	// Check for CLAUDE.md
	claudeMd := core.PathJoin(repo.Path, "CLAUDE.md")
	if io.Local.IsFile(claudeMd) {
		info.ClaudeMd = claudeMd
		info.HasDocs = true
	}

	// Check for CHANGELOG.md
	changelog := core.PathJoin(repo.Path, "CHANGELOG.md")
	if io.Local.IsFile(changelog) {
		info.Changelog = changelog
		info.HasDocs = true
	}

	// Recursively scan docs/ directory for .md files
	docsDir := core.PathJoin(repo.Path, "docs")
	// Check if directory exists by listing it
	if _, err := io.Local.List(docsDir); err == nil {
		if err := core.PathWalkDir(docsDir, func(path string, d core.FsDirEntry, err error) error {
			if err != nil {
				return nil
			}
			// Skip plans/ directory
			if d.IsDir() && d.Name() == "plans" {
				return core.PathSkipDir
			}
			// Skip non-markdown files
			if d.IsDir() || !core.HasSuffix(d.Name(), ".md") {
				return nil
			}
			// Get relative path from docs/
			relResult := core.PathRel(docsDir, path)
			if !relResult.OK {
				return relResult.Value.(error)
			}
			info.DocsFiles = append(info.DocsFiles, relResult.Value.(string))
			info.HasDocs = true
			return nil
		}); err != nil {
			return info
		}
	}

	// Recursively scan KB/ directory for .md files
	kbDir := core.PathJoin(repo.Path, "KB")
	if _, err := io.Local.List(kbDir); err == nil {
		if err := core.PathWalkDir(kbDir, func(path string, d core.FsDirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() || !core.HasSuffix(d.Name(), ".md") {
				return nil
			}
			relResult := core.PathRel(kbDir, path)
			if !relResult.OK {
				return relResult.Value.(error)
			}
			info.KBFiles = append(info.KBFiles, relResult.Value.(string))
			info.HasDocs = true
			return nil
		}); err != nil {
			return info
		}
	}

	return info
}
