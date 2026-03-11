---
title: Build System
description: Project detection, cross-compilation, code signing, and .core/build.yaml configuration.
---

# Build System

The `build/` package provides automatic project type detection, cross-compilation for multiple OS/architecture targets, artefact archiving with checksums, and code signing. Configuration lives in `.core/build.yaml` at the project root.

## How it works

1. **Detect** — Probes the project directory for marker files to determine the project type.
2. **Build** — Runs the appropriate builder plugin for each configured target (OS/arch pair).
3. **Archive** — Packages build outputs into compressed archives with SHA-256 checksums.
4. **Sign** — Optionally signs binaries before archiving (macOS codesign, GPG, or Windows signtool).

Run with:

```bash
core build                # Build for configured targets
core build --ci           # All targets, JSON output for CI pipelines
```

## Project detection

`discovery.go` probes marker files in priority order. The first match wins:

| Marker file | Detected type | Builder |
|-------------|--------------|---------|
| `wails.json` | `wails` | Wails v3 desktop app |
| `go.mod` | `go` | Standard Go binary |
| `package.json` | `node` | Node.js (stub) |
| `composer.json` | `php` | PHP (stub) |
| `CMakeLists.txt` | `cpp` | CMake C++ |
| `Dockerfile` | `docker` | Docker buildx |
| `*.yml` (LinuxKit pattern) | `linuxkit` | LinuxKit VM image |
| `Taskfile.yml` | `taskfile` | Task CLI delegation |

## Builder interface

Each builder implements three methods:

```go
type Builder interface {
    Name() string
    Detect(fs io.Medium, dir string) (bool, error)
    Build(ctx context.Context, cfg *Config, targets []Target) ([]Artifact, error)
}
```

- `Detect` checks whether the builder should handle a given directory.
- `Build` produces `[]Artifact`, where each artefact has a file `Path`, `OS`, `Arch`, and a SHA-256 `Checksum`.

## Builders

### Go builder (`go.go`)

Runs `go build` with ldflags injection and cross-compilation via `GOOS`/`GOARCH` environment variables. Reads the entry point from `build.yaml`'s `project.main` field.

### Wails builder (`wails.go`)

Builds Wails v3 desktop applications with platform-specific packaging. Uses `wails.json` for application metadata.

### Docker builder (`docker.go`)

Runs `docker buildx build` with optional multi-platform support and registry push.

### C++ builder (`cpp.go`)

Runs CMake configure and build in a temporary directory. Requires `CMakeLists.txt`.

### LinuxKit builder (`linuxkit.go`)

Produces multi-format VM images from LinuxKit YAML configurations. Output formats: ISO, qcow2, raw, VMDK.

### Taskfile builder (`taskfile.go`)

Delegates to the `task` CLI, mapping build targets to Taskfile tasks. Used for projects that have not yet migrated to native builders.

### Node and PHP stubs

Stub builders for Node.js and PHP projects. Node runs `npm run build`; PHP wraps Docker builds. Both are minimal and intended for extension.

## Artefacts and checksums

Each builder produces `[]Artifact`:

```go
type Artifact struct {
    Path     string // File path of the built artefact
    OS       string // Target operating system
    Arch     string // Target architecture
    Checksum string // SHA-256 hash
}
```

Checksums are computed automatically and written to `dist/*.sha256` files.

## Archive formats

`archive.go` packages build outputs into compressed archives. Supported formats:

- `tar.gz` — default for Linux/macOS
- `tar.xz` — smaller archives (uses Borg's xz compression)
- `zip` — default for Windows

## Code signing

The `signing/` sub-package implements a `Signer` interface with three backends:

| Signer | Platform | Tool | Key source |
|--------|----------|------|-----------|
| macOS | darwin | `codesign` | Keychain identity |
| GPG | any | `gpg --detach-sign` | GPG key ID |
| Windows | windows | `signtool` | Certificate store |

Each signer checks `Available()` at runtime to verify the signing tool exists. Signing runs after compilation, before archiving.

## .core/build.yaml reference

```yaml
version: 1

project:
  name: my-app              # Project name (used in archive filenames)
  description: My application
  main: ./cmd/my-app        # Go entry point (go/wails only)
  binary: my-app            # Output binary name

build:
  cgo: false                # Enable CGO (default: false)
  flags:
    - -trimpath             # Go build flags
  ldflags:
    - -s                    # Strip debug info
    - -w                    # Strip DWARF

targets:                    # Cross-compilation targets
  - os: linux
    arch: amd64
  - os: linux
    arch: arm64
  - os: darwin
    arch: arm64
  - os: windows
    arch: amd64
```

### Field reference

| Field | Type | Description |
|-------|------|-------------|
| `version` | `int` | Config schema version (always `1`) |
| `project.name` | `string` | Project name, used in archive filenames |
| `project.description` | `string` | Human-readable description |
| `project.main` | `string` | Go entry point path (e.g. `./cmd/myapp`) |
| `project.binary` | `string` | Output binary filename |
| `build.cgo` | `bool` | Whether to enable CGO (default `false`) |
| `build.flags` | `[]string` | Go build flags (e.g. `-trimpath`) |
| `build.ldflags` | `[]string` | Linker flags (e.g. `-s -w`) |
| `targets` | `[]Target` | List of OS/arch pairs to compile for |
| `targets[].os` | `string` | Target OS (`linux`, `darwin`, `windows`) |
| `targets[].arch` | `string` | Target architecture (`amd64`, `arm64`) |

### Generated configuration

Running `core setup repo` in a git repository auto-generates `.core/build.yaml` based on detected project type:

```bash
core setup repo              # Generates .core/build.yaml, release.yaml, test.yaml
core setup repo --dry-run    # Preview without writing files
```

## Adding a new builder

1. Create `build/builders/mylang.go`.
2. Implement the `Builder` interface:
   - `Name()` returns a human-readable name.
   - `Detect(fs, dir)` checks for a marker file (e.g. `myconfig.toml`).
   - `Build(ctx, cfg, targets)` produces artefacts.
3. Register the builder in `build/buildcmd/`.
4. Write tests verifying `Detect` (marker present/absent) and `Build` (at minimum with a mock `io.Medium`).
