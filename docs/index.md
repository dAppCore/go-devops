---
title: go-devops
description: Build system, release publishers, infrastructure management, and DevOps tooling for the Lethean ecosystem.
---

# go-devops

`forge.lthn.ai/core/go-devops` is the build, release, and infrastructure automation library for the Lethean ecosystem. It replaces goreleaser with a native Go pipeline that auto-detects project types, cross-compiles, signs artefacts, generates changelogs, and publishes to eight distribution targets.

**Module**: `forge.lthn.ai/core/go-devops`
**Go**: 1.26
**Licence**: EUPL-1.2

## What it does

| Area | Summary |
|------|---------|
| **Build system** | Auto-detect project type from marker files, cross-compile for multiple OS/arch targets, archive and checksum artefacts |
| **Code signing** | macOS `codesign`, GPG detached signatures, Windows `signtool` |
| **Release publishers** | GitHub Releases, Docker, Homebrew, npm, AUR, Scoop, Chocolatey, LinuxKit |
| **SDK generation** | Generate typed API clients from OpenAPI specs (TypeScript, Python, Go, PHP) with breaking change detection |
| **Ansible executor** | Native Go playbook runner with ~30 modules over SSH — no `ansible-playbook` shell-out |
| **Infrastructure** | Hetzner Cloud/Robot provisioning, CloudNS DNS management |
| **Container/VM** | LinuxKit-based VMs via QEMU (Linux) or Hyperkit (macOS) |
| **Developer toolkit** | Cyclomatic complexity analysis, vulnerability scanning, coverage trending, secret scanning |
| **Doc sync** | Collect documentation from multi-repo workspaces into a central location |

## Package layout

```
go-devops/
├── ansible/          Ansible playbook execution engine (native Go, no shell-out)
├── build/            Build system: project detection, archives, checksums
│   ├── builders/     Builders: Go, Wails, Docker, C++, LinuxKit, Taskfile
│   ├── signing/      Code signing: macOS codesign, GPG, Windows signtool
│   └── buildcmd/     CLI handlers for core build / core release
├── container/        LinuxKit VM management, hypervisor abstraction
├── deploy/           Deployment integrations (Coolify PaaS, embedded Python)
├── devkit/           Code quality, security, coverage trending
├── devops/           Portable dev environment management
│   └── sources/      Image download: GitHub Releases, S3/CDN
├── infra/            Infrastructure APIs: Hetzner Cloud, Hetzner Robot, CloudNS
├── release/          Release orchestration: version, changelog, publishing
│   └── publishers/   8 publisher backends
├── sdk/              OpenAPI SDK generation and breaking change detection
│   └── generators/   Language generators: TypeScript, Python, Go, PHP
├── snapshot/         Frozen release manifest generation (core.json)
└── cmd/              CLI command registrations
    ├── dev/          Multi-repo workflow commands (work, health, commit, push, pull)
    ├── docs/         Documentation sync and listing
    ├── deploy/       Coolify deployment commands
    ├── setup/        Repository and CI bootstrapping
    └── gitcmd/       Git helpers
```

## CLI commands

go-devops registers commands into the `core` CLI binary (built from `forge.lthn.ai/core/cli`). Key commands:

```bash
# Build
core build                     # Auto-detect project type, build for configured targets
core build --ci                # All targets, JSON output
core build sdk                 # Generate SDKs from OpenAPI spec

# Release
core build release             # Build + changelog + publish (requires --we-are-go-for-launch)

# Multi-repo development
core dev health                # Quick summary across all repos
core dev work                  # Combined status, commit, push workflow
core dev commit                # Claude-assisted commits for dirty repos
core dev push                  # Push repos with unpushed commits
core dev pull                  # Pull repos behind remote

# GitHub integration
core dev issues                # List open issues across repos
core dev reviews               # PRs needing review
core dev ci                    # GitHub Actions status

# Documentation
core docs list                 # Scan repos for docs
core docs sync                 # Copy docs to central location
core docs sync --target gohelp # Sync to go-help format

# Deployment
core deploy servers            # List Coolify servers
core deploy apps               # List Coolify applications

# Setup
core setup repo                # Generate .core/ configuration for a repo
core setup ci                  # Bootstrap CI configuration
```

## Configuration

Two YAML files in `.core/` at the project root control build and release behaviour:

| File | Purpose |
|------|---------|
| `.core/build.yaml` | Project name, binary, build flags, cross-compilation targets |
| `.core/release.yaml` | Repository, changelog rules, publisher configs, SDK settings |

See [Build System](build-system.md) and [Publishers](publishers.md) for full configuration reference.

## Core interfaces

Every extensible subsystem is defined by a small interface:

```go
// Builder — project type plugin (build/builders/)
type Builder interface {
    Name() string
    Detect(fs io.Medium, dir string) (bool, error)
    Build(ctx context.Context, cfg *Config, targets []Target) ([]Artifact, error)
}

// Publisher — distribution target plugin (release/publishers/)
type Publisher interface {
    Name() string
    Publish(ctx context.Context, release *Release, pubCfg PublisherConfig,
            relCfg ReleaseConfig, dryRun bool) error
}

// Generator — SDK language generator (sdk/generators/)
type Generator interface {
    Language() string
    Generate(ctx context.Context, spec, outputDir string, config *Config) error
}

// Signer — code signing plugin (build/signing/)
type Signer interface {
    Name() string
    Available() bool
    Sign(filePath, keyID string) ([]byte, error)
}
```

## Further reading

- [Build System](build-system.md) — Builders, project detection, `.core/build.yaml` reference
- [Publishers](publishers.md) — Release publishers, `.core/release.yaml` reference
- [SDK Generation](sdk-generation.md) — OpenAPI client generation and breaking change detection
- [Doc Sync](sync.md) — Documentation sync across multi-repo workspaces
- [Architecture](architecture.md) — Full architecture deep-dive (Ansible, infra, devkit, containers)
- [Development Guide](development.md) — Building, testing, coding standards
