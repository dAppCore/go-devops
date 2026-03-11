---
title: Release Publishers
description: Release orchestration, changelog generation, eight publisher backends, and .core/release.yaml configuration.
---

# Release Publishers

The `release/` package orchestrates the full release pipeline: version detection from git tags, changelog generation from commit history, and publishing to multiple distribution targets. Configuration lives in `.core/release.yaml`.

## How it works

The build and release stages are deliberately separate, allowing CI pipelines to build once and publish to multiple targets independently.

```
core build           # 1. Compile artefacts into dist/
core build release   # 2. Version + changelog + publish (requires --we-are-go-for-launch)
```

The release pipeline runs these steps:

1. **Resolve version** — from git tags on HEAD, or increment the latest tag's patch version.
2. **Scan artefacts** — find pre-built files in `dist/`.
3. **Generate changelog** — parse conventional commit messages since the previous tag.
4. **Publish** — iterate configured publishers, calling `Publisher.Publish()` on each.

## Version detection

`DetermineVersion(dir)` checks in priority order:

1. Git tag on `HEAD` (exact match).
2. Most recent tag with patch version incremented (e.g. `v1.2.3` becomes `v1.2.4`).
3. Default `v0.0.1` if no tags exist.

Pre-release suffixes are stripped when incrementing.

## Changelog generation

`Generate()` in `changelog.go` reads the git log since the previous tag and groups entries by conventional commit prefix:

| Prefix | Section |
|--------|---------|
| `feat:` | Features |
| `fix:` | Bug Fixes |
| `perf:` | Performance |
| `refactor:` | Refactoring |

The `changelog` section in `release.yaml` controls which prefixes are included or excluded.

## Publisher interface

All publishers implement:

```go
type Publisher interface {
    Name() string
    Publish(ctx context.Context, release *Release, pubCfg PublisherConfig,
            relCfg ReleaseConfig, dryRun bool) error
}
```

When `dryRun` is `true`, publishers log what they would do without making external calls.

## Publisher backends

### GitHub Releases (`github.go`)

Creates a GitHub Release via the API and uploads artefact files as release assets.

```yaml
publishers:
  - type: github
    draft: false
    prerelease: false
```

### Docker (`docker.go`)

Runs `docker buildx build --push` to a configured container registry. Supports multi-platform images.

```yaml
publishers:
  - type: docker
    registry: ghcr.io
    image: myorg/myapp
```

### Homebrew (`homebrew.go`)

Generates a Ruby formula file and commits it to a Homebrew tap repository.

```yaml
publishers:
  - type: homebrew
    tap: myorg/homebrew-tap
    formula: myapp
```

### npm (`npm.go`)

Runs `npm publish` to the npm registry.

```yaml
publishers:
  - type: npm
    registry: https://registry.npmjs.org
```

### AUR (`aur.go`)

Generates a `PKGBUILD` and `.SRCINFO`, then pushes to the Arch User Repository git remote.

```yaml
publishers:
  - type: aur
    package: myapp
```

### Scoop (`scoop.go`)

Generates a JSON manifest and commits it to a Scoop bucket repository.

```yaml
publishers:
  - type: scoop
    bucket: myorg/scoop-bucket
```

### Chocolatey (`chocolatey.go`)

Generates a `.nuspec` file and calls `choco push`.

```yaml
publishers:
  - type: chocolatey
    package: myapp
```

### LinuxKit (`linuxkit.go`)

Builds and uploads LinuxKit multi-format VM images (ISO, qcow2, raw, VMDK).

```yaml
publishers:
  - type: linuxkit
    formats:
      - qcow2
      - iso
```

## .core/release.yaml reference

```yaml
version: 1

project:
  name: my-app
  repository: myorg/my-app     # GitHub owner/repo (auto-detected from git remote)

changelog:
  include:                      # Conventional commit types to include
    - feat
    - fix
    - perf
    - refactor
  exclude:                      # Types to exclude
    - chore
    - docs
    - style
    - test
    - ci

publishers:                     # List of publisher configurations
  - type: github
    draft: false
    prerelease: false

# Optional: SDK generation (see sdk-generation.md)
sdk:
  spec: openapi.yaml
  languages: [typescript, python, go, php]
  output: sdk
  diff:
    enabled: true
    fail_on_breaking: false
```

### Field reference

| Field | Type | Description |
|-------|------|-------------|
| `version` | `int` | Config schema version (always `1`) |
| `project.name` | `string` | Project name |
| `project.repository` | `string` | Git remote in `owner/repo` format |
| `changelog.include` | `[]string` | Conventional commit prefixes to include in the changelog |
| `changelog.exclude` | `[]string` | Prefixes to exclude |
| `publishers` | `[]PublisherConfig` | List of publisher configurations |
| `publishers[].type` | `string` | Publisher type: `github`, `docker`, `homebrew`, `npm`, `aur`, `scoop`, `chocolatey`, `linuxkit` |
| `sdk` | `SDKConfig` | Optional SDK generation settings (see [SDK Generation](sdk-generation.md)) |

### Generated configuration

`core setup repo` generates a default `release.yaml` with sensible defaults for the detected project type:

- **Go/Wails projects** get changelog rules for `feat`, `fix`, `perf`, `refactor` and a GitHub publisher.
- **PHP projects** get `feat`, `fix`, `perf` and a GitHub publisher.
- **Other projects** get `feat`, `fix` and a GitHub publisher.

## Adding a new publisher

1. Create `release/publishers/myplatform.go`.
2. Implement `Publisher`:
   - `Name()` returns the platform name (matches the `type` field in `release.yaml`).
   - `Publish(ctx, release, pubCfg, relCfg, dryRun)` performs the publication. When `dryRun` is `true`, log intent and return `nil`.
3. Register the publisher in `release/config.go`.
4. Write `release/publishers/myplatform_test.go` with dry-run tests. Verify command arguments, generated file content, and interface compliance.
