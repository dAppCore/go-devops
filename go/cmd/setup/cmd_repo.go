// cmd_repo.go implements repository setup with .core/ configuration.
//
// When running setup in an existing git repository, this generates
// build.yaml, release.yaml, and test.yaml configurations based on
// detected project type.

package setup

import (
	"net/url"

	core "dappco.re/go"
	"dappco.re/go/cli/pkg/cli"
	"dappco.re/go/i18n"
	coreio "dappco.re/go/io"
	log "dappco.re/go/log"
	coreexec "dappco.re/go/process/exec"
)

var repoDryRun bool

// addRepoCommand adds the 'repo' subcommand to generate .core configuration.
func addRepoCommand(parent *cli.Command) {
	repoCmd := &cli.Command{
		Use:   "repo",
		Short: i18n.T("cmd.setup.repo.short"),
		Long:  i18n.T("cmd.setup.repo.long"),
		Args:  cli.ExactArgs(0),
		RunE: func(cmd *cli.Command, args []string) error {
			cwdResult := core.Getwd()
			if !cwdResult.OK {
				return log.E("setup.repo", "failed to get working directory", cwdResult.Value.(error))
			}

			return resultError(runRepoSetup(cwdResult.Value.(string), repoDryRun))
		},
	}

	repoCmd.Flags().BoolVar(&repoDryRun, "dry-run", false, i18n.T("cmd.setup.flag.dry_run"))

	parent.AddCommand(repoCmd)
}

// runRepoSetup sets up the current repository with .core/ configuration.
func runRepoSetup(repoPath string, dryRun bool) (_ core.Result) {
	cli.Print("%s %s: %s\n", dimStyle.Render(">>"), i18n.T("cmd.setup.repo.setting_up"), repoPath)

	// Detect project type
	projectType := detectProjectType(repoPath)
	cli.Print("%s %s: %s\n", dimStyle.Render(">>"), i18n.T("cmd.setup.repo.detected_type"), projectType)

	// Create .core directory
	coreDir := core.PathJoin(repoPath, ".core")
	if !dryRun {
		if err := coreio.Local.EnsureDir(coreDir); err != nil {
			return core.Fail(log.E("setup.repo", "failed to create .core directory", err))
		}
	}

	// Generate configs based on project type
	name := core.PathBase(repoPath)
	configs := map[string]string{
		"build.yaml":   generateBuildConfig(repoPath, projectType),
		"release.yaml": generateReleaseConfig(name, projectType),
		"test.yaml":    generateTestConfig(projectType),
	}

	if dryRun {
		cli.Print("\n%s %s:\n", dimStyle.Render(">>"), i18n.T("cmd.setup.repo.would_create"))
		for filename, content := range configs {
			cli.Print("\n  %s:\n", core.PathJoin(coreDir, filename))
			// Indent content for display
			for _, line := range core.Split(content, "\n") {
				cli.Print("    %s\n", line)
			}
		}
		return core.Ok(nil)
	}

	for filename, content := range configs {
		configPath := core.PathJoin(coreDir, filename)
		if err := coreio.Local.Write(configPath, content); err != nil {
			return core.Fail(log.E("setup.repo", core.Sprintf("failed to write %s", filename), err))
		}
		cli.Print("%s %s %s\n", successStyle.Render(">>"), i18n.T("cmd.setup.repo.created"), configPath)
	}

	return core.Ok(nil)
}

// detectProjectType identifies the project type from files present.
func detectProjectType(path string) string {
	// Check in priority order
	if coreio.Local.IsFile(core.PathJoin(path, "wails.json")) {
		return "wails"
	}
	if coreio.Local.IsFile(core.PathJoin(path, "go.mod")) {
		return "go"
	}
	if coreio.Local.IsFile(core.PathJoin(path, "package.json")) {
		return "node"
	}
	if coreio.Local.IsFile(core.PathJoin(path, "composer.json")) {
		return "php"
	}
	return "unknown"
}

// generateBuildConfig creates a build.yaml configuration based on project type.
func generateBuildConfig(path, projectType string) string {
	name := core.PathBase(path)

	switch projectType {
	case "go", "wails":
		return core.Sprintf(`version: 1
project:
  name: %s
  description: Go application
  main: ./cmd/%s
  binary: %s
build:
  cgo: false
  flags:
    - -trimpath
  ldflags:
    - -s
    - -w
targets:
  - os: linux
    arch: amd64
  - os: linux
    arch: arm64
  - os: darwin
    arch: amd64
  - os: darwin
    arch: arm64
  - os: windows
    arch: amd64
`, name, name, name)

	case "php":
		return core.Sprintf(`version: 1
project:
  name: %s
  description: PHP application
  type: php
build:
  dockerfile: Dockerfile
  image: %s
`, name, name)

	case "node":
		return core.Sprintf(`version: 1
project:
  name: %s
  description: Node.js application
  type: node
build:
  script: npm run build
  output: dist
`, name)

	default:
		return core.Sprintf(`version: 1
project:
  name: %s
  description: Application
`, name)
	}
}

// generateReleaseConfig creates a release.yaml configuration.
func generateReleaseConfig(name, projectType string) string {
	// Try to detect GitHub repo from git remote
	repo := detectGitHubRepo()
	if repo == "" {
		repo = "owner/" + name
	}

	base := core.Sprintf(`version: 1
project:
  name: %s
  repository: %s
`, name, repo)

	switch projectType {
	case "go", "wails":
		return base + `
changelog:
  include:
    - feat
    - fix
    - perf
    - refactor
  exclude:
    - chore
    - docs
    - style
    - test

publishers:
  - type: github
    draft: false
    prerelease: false
`
	case "php":
		return base + `
changelog:
  include:
    - feat
    - fix
    - perf

publishers:
  - type: github
    draft: false
`
	default:
		return base + `
changelog:
  include:
    - feat
    - fix

publishers:
  - type: github
`
	}
}

// generateTestConfig creates a test.yaml configuration.
func generateTestConfig(projectType string) string {
	switch projectType {
	case "go", "wails":
		return `version: 1

commands:
  - name: unit
    run: go test ./...
  - name: coverage
    run: go test -coverprofile=coverage.out ./...
  - name: race
    run: go test -race ./...

env:
  CGO_ENABLED: "0"
`
	case "php":
		return `version: 1

commands:
  - name: unit
    run: vendor/bin/pest --parallel
  - name: types
    run: vendor/bin/phpstan analyse
  - name: lint
    run: vendor/bin/pint --test

env:
  APP_ENV: testing
  DB_CONNECTION: sqlite
`
	case "node":
		return `version: 1

commands:
  - name: unit
    run: npm test
  - name: lint
    run: npm run lint
  - name: typecheck
    run: npm run typecheck

env:
  NODE_ENV: test
`
	default:
		return `version: 1

commands:
  - name: test
    run: echo "No tests configured"
`
	}
}

// detectGitHubRepo tries to extract owner/repo from git remote.
func detectGitHubRepo() string {
	cmd := coreexec.Command(core.Background(), "git", "remote", "get-url", "origin")
	r := cmd.Output()
	if !r.OK {
		return ""
	}

	return parseGitHubRepoURL(core.Trim(string(r.Value.([]byte))))
}

// parseGitHubRepoURL extracts owner/repo from a GitHub remote URL.
//
// Supports the common remote formats used by git:
// - git@github.com:owner/repo.git
// - ssh://git@github.com/owner/repo.git
// - https://github.com/owner/repo.git
// - git://github.com/owner/repo.git
func parseGitHubRepoURL(remote string) string {
	remote = core.Trim(remote)
	if remote == "" {
		return ""
	}

	// Handle SSH-style scp syntax first.
	if core.HasPrefix(remote, "git@github.com:") {
		repo := core.TrimPrefix(remote, "git@github.com:")
		return core.TrimSuffix(repo, ".git")
	}

	if parsed, err := url.Parse(remote); err == nil && parsed.Host != "" {
		host := core.TrimPrefix(parsed.Hostname(), "www.")
		if host == "github.com" {
			repo := core.TrimPrefix(parsed.Path, "/")
			repo = core.TrimSuffix(repo, ".git")
			repo = core.TrimSuffix(repo, "/")
			return repo
		}
	}

	if core.Contains(remote, "github.com/") {
		parts := core.SplitN(remote, "github.com/", 2)
		if len(parts) == 2 {
			repo := core.TrimPrefix(parts[1], "/")
			repo = core.TrimSuffix(repo, ".git")
			return core.TrimSuffix(repo, "/")
		}
	}

	return ""
}
