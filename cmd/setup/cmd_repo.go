// cmd_repo.go implements repository setup with .core/ configuration.
//
// When running setup in an existing git repository, this generates
// build.yaml, release.yaml, and test.yaml configurations based on
// detected project type.

package setup

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"dappco.re/go/core/i18n"
	coreio "dappco.re/go/core/io"
	log "dappco.re/go/core/log"
	"dappco.re/go/core/cli/pkg/cli"
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
			cwd, err := os.Getwd()
			if err != nil {
				return log.E("setup.repo", "failed to get working directory", err)
			}

			return runRepoSetup(cwd, repoDryRun)
		},
	}

	repoCmd.Flags().BoolVar(&repoDryRun, "dry-run", false, i18n.T("cmd.setup.flag.dry_run"))

	parent.AddCommand(repoCmd)
}

// runRepoSetup sets up the current repository with .core/ configuration.
func runRepoSetup(repoPath string, dryRun bool) error {
	fmt.Printf("%s %s: %s\n", dimStyle.Render(">>"), i18n.T("cmd.setup.repo.setting_up"), repoPath)

	// Detect project type
	projectType := detectProjectType(repoPath)
	fmt.Printf("%s %s: %s\n", dimStyle.Render(">>"), i18n.T("cmd.setup.repo.detected_type"), projectType)

	// Create .core directory
	coreDir := filepath.Join(repoPath, ".core")
	if !dryRun {
		if err := coreio.Local.EnsureDir(coreDir); err != nil {
			return log.E("setup.repo", "failed to create .core directory", err)
		}
	}

	// Generate configs based on project type
	name := filepath.Base(repoPath)
	configs := map[string]string{
		"build.yaml":   generateBuildConfig(repoPath, projectType),
		"release.yaml": generateReleaseConfig(name, projectType),
		"test.yaml":    generateTestConfig(projectType),
	}

	if dryRun {
		fmt.Printf("\n%s %s:\n", dimStyle.Render(">>"), i18n.T("cmd.setup.repo.would_create"))
		for filename, content := range configs {
			fmt.Printf("\n  %s:\n", filepath.Join(coreDir, filename))
			// Indent content for display
			for _, line := range strings.Split(content, "\n") {
				fmt.Printf("    %s\n", line)
			}
		}
		return nil
	}

	for filename, content := range configs {
		configPath := filepath.Join(coreDir, filename)
		if err := coreio.Local.Write(configPath, content); err != nil {
			return log.E("setup.repo", fmt.Sprintf("failed to write %s", filename), err)
		}
		fmt.Printf("%s %s %s\n", successStyle.Render(">>"), i18n.T("cmd.setup.repo.created"), configPath)
	}

	return nil
}

// detectProjectType identifies the project type from files present.
func detectProjectType(path string) string {
	// Check in priority order
	if coreio.Local.IsFile(filepath.Join(path, "wails.json")) {
		return "wails"
	}
	if coreio.Local.IsFile(filepath.Join(path, "go.mod")) {
		return "go"
	}
	if coreio.Local.IsFile(filepath.Join(path, "package.json")) {
		return "node"
	}
	if coreio.Local.IsFile(filepath.Join(path, "composer.json")) {
		return "php"
	}
	return "unknown"
}

// generateBuildConfig creates a build.yaml configuration based on project type.
func generateBuildConfig(path, projectType string) string {
	name := filepath.Base(path)

	switch projectType {
	case "go", "wails":
		return fmt.Sprintf(`version: 1
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
		return fmt.Sprintf(`version: 1
project:
  name: %s
  description: PHP application
  type: php
build:
  dockerfile: Dockerfile
  image: %s
`, name, name)

	case "node":
		return fmt.Sprintf(`version: 1
project:
  name: %s
  description: Node.js application
  type: node
build:
  script: npm run build
  output: dist
`, name)

	default:
		return fmt.Sprintf(`version: 1
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

	base := fmt.Sprintf(`version: 1
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
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return parseGitHubRepoURL(strings.TrimSpace(string(output)))
}

// parseGitHubRepoURL extracts owner/repo from a GitHub remote URL.
//
// Supports the common remote formats used by git:
// - git@github.com:owner/repo.git
// - ssh://git@github.com/owner/repo.git
// - https://github.com/owner/repo.git
// - git://github.com/owner/repo.git
func parseGitHubRepoURL(remote string) string {
	remote = strings.TrimSpace(remote)
	if remote == "" {
		return ""
	}

	// Handle SSH-style scp syntax first.
	if strings.HasPrefix(remote, "git@github.com:") {
		repo := strings.TrimPrefix(remote, "git@github.com:")
		return strings.TrimSuffix(repo, ".git")
	}

	if parsed, err := url.Parse(remote); err == nil && parsed.Host != "" {
		host := strings.TrimPrefix(parsed.Hostname(), "www.")
		if host == "github.com" {
			repo := strings.TrimPrefix(parsed.Path, "/")
			repo = strings.TrimSuffix(repo, ".git")
			repo = strings.TrimSuffix(repo, "/")
			return repo
		}
	}

	if strings.Contains(remote, "github.com/") {
		parts := strings.SplitN(remote, "github.com/", 2)
		if len(parts) == 2 {
			repo := strings.TrimPrefix(parts[1], "/")
			repo = strings.TrimSuffix(repo, ".git")
			return strings.TrimSuffix(repo, "/")
		}
	}

	return ""
}
