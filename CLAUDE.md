# CLAUDE.md — go-devops Domain Expert Guide

You are a dedicated domain expert for `forge.lthn.ai/core/go-devops`. Virgil (in core/go) orchestrates your work via TODO.md. Pick up tasks in phase order, mark `[x]` when done, commit and push.

## What This Package Does

Infrastructure management, build automation, and release pipelines. ~29K LOC across 118 Go files. Provides:

- **Ansible engine** — Native Go playbook executor (not shelling out to `ansible-playbook`). SSH, modules, facts, handlers.
- **Build system** — Plugin-based builders (Go, Wails, Docker, C++, LinuxKit, Taskfile). Cross-compilation, code signing (macOS/GPG/Windows).
- **Release automation** — Version detection, changelog from git history, multi-target publishing (GitHub, Docker, Homebrew, AUR, Scoop, Chocolatey, npm).
- **Infrastructure APIs** — Hetzner Cloud, Hetzner Robot (bare metal), CloudNS DNS.
- **Container/VM management** — LinuxKit images on QEMU (Linux) or Hyperkit (macOS).
- **SDK generation** — OpenAPI spec parsing, TypeScript/Python/Go/PHP client generation, breaking change detection.
- **Developer toolkit** — Code quality metrics, TODO detection, coverage reports, dependency graphs.

## Commands

```bash
go test ./...                    # Run all tests
go test -v -run TestName ./...   # Single test
go test -race ./...              # Race detector
go vet ./...                     # Static analysis
```

## Local Dependencies

| Module | Local Path | Notes |
|--------|-----------|-------|
| `forge.lthn.ai/core/go` | `../core` | Framework (core.E, io.Medium, config, i18n, log) |

**Do NOT change the replace directive path.** Use go.work for local resolution if needed.

## Architecture

### ansible/ — Playbook Execution Engine (~3,162 LOC)

| File | LOC | Purpose |
|------|-----|---------|
| `executor.go` | 1,021 | Playbook runner: task/handler/fact tracking, become/sudo |
| `modules.go` | 1,434 | Module implementations: service, file, template, command, copy, apt, yum |
| `parser.go` | 438 | YAML playbook + inventory parser |
| `ssh.go` | 451 | SSH client connection management |
| `types.go` | 258 | Core types: Play, Task, Handler, Inventory, Facts |

Executes Ansible playbooks natively in Go. Supports: `when` conditionals, `register` variables, `notify` handlers, `become` privilege escalation, `loop` iteration, fact gathering.

### build/ — Project Building & Cross-Compilation (~3,637 LOC)

**Root** (797 LOC): Project type detection, archive creation (tar.gz/xz/zip via Borg compression), config from `.core/build.yaml`, SHA checksums.

**builders/** (1,390 LOC): Plugin interface `Builder.Build()`.
| Builder | LOC | Notes |
|---------|-----|-------|
| `go.go` | — | Go cross-compilation |
| `wails.go` | 247 | Wails desktop app |
| `docker.go` | 215 | Docker image build |
| `cpp.go` | 253 | CMake C++ |
| `linuxkit.go` | 270 | LinuxKit VM image |
| `taskfile.go` | 275 | Taskfile automation |

**signing/** (377 LOC): Signer interface. macOS `codesign`, GPG, Windows `signtool`.

**buildcmd/** (1,053 LOC): CLI handlers for `core build`, `core build pwa`, `core build sdk`, `core release`.

### container/ — LinuxKit VM Management (~1,208 LOC)

| File | LOC | Purpose |
|------|-----|---------|
| `container.go` | 106 | Manager interface + Container model |
| `linuxkit.go` | 462 | LinuxKitManager: Run, Stop, List |
| `hypervisor.go` | 273 | Abstraction: QEMU (Linux) / Hyperkit (macOS) |
| `state.go` | 172 | Container state persistence (`~/.core/state.json`) |
| `templates.go` | 301 | Packer/LinuxKit template rendering |

### devops/ — Portable Dev Environment (~1,216 LOC)

| File | LOC | Purpose |
|------|-----|---------|
| `devops.go` | 243 | Manager: install, boot, stop, status |
| `config.go` | 90 | Config from `~/.core/config.yaml` |
| `images.go` | 198 | ImageManager: download from GitHub/CDN/registry |
| `shell.go` | 74 | Shell execution wrapper |
| `test.go` | 188 | Test execution helpers |
| `serve.go` | 109 | Dev environment HTTP server |
| `claude.go` | 143 | Claude/AI integration |
| `ssh_utils.go` | 68 | SSH key scanning |

**sources/** (218 LOC): `ImageSource` interface. GitHub Releases + S3/CDN download sources.

### infra/ — Infrastructure APIs (~953 LOC)

| File | LOC | Purpose |
|------|-----|---------|
| `config.go` | 300 | `infra.yaml` types: Host, LoadBalancer, Network, DNS, Database, Cache |
| `hetzner.go` | 381 | Hetzner Cloud API (VPS) + Hetzner Robot API (bare metal) |
| `cloudns.go` | 272 | CloudNS DNS: zones, records, ACME DNS-01 challenges |

### release/ — Release Automation (~4,008 LOC)

**Root** (1,398 LOC): Release orchestrator (version → build → changelog → publish), config from `.core/release.yaml`, git-based changelog, semver detection.

**publishers/** (2,610 LOC): `Publisher` interface.
| Publisher | LOC | Notes |
|-----------|-----|-------|
| `github.go` | 233 | GitHub Releases |
| `docker.go` | 278 | Docker image build + push |
| `homebrew.go` | 371 | Homebrew formula |
| `npm.go` | 265 | npm registry |
| `aur.go` | 313 | Arch Linux AUR |
| `scoop.go` | 284 | Windows Scoop |
| `chocolatey.go` | 294 | Windows Chocolatey |
| `linuxkit.go` | 300 | LinuxKit image |

### sdk/ — OpenAPI SDK Generation (~931 LOC)

Auto-detect OpenAPI spec, generate typed clients in 4 languages, detect breaking changes via oasdiff.

**generators/** (437 LOC): TypeScript, Python, Go, PHP generators.

### devkit/ — Developer Toolkit (~560 LOC)

Code quality analysis: TODOs/FIXMEs, coverage reports, race conditions, vulnerability detection, secret leak scanning, cyclomatic complexity, dependency graphs.

### deploy/ — Deployment Integrations (~366 LOC)

- **python/** — Embedded Python 3.13 runtime (kluctl/go-embed-python)
- **coolify/** — Coolify PaaS API client via Python Swagger

## Key Interfaces

```go
// build/builders/
type Builder interface {
    Name() string
    Detect(fs io.Medium, dir string) (bool, error)
    Build(ctx context.Context, cfg *Config, targets []Target) ([]Artifact, error)
}

// release/publishers/
type Publisher interface {
    Name() string
    Publish(ctx context.Context, release *Release, pubCfg PublisherConfig, relCfg ReleaseConfig, dryRun bool) error
}

// container/
type Hypervisor interface {
    Name() string
    Available() bool
    Run(ctx context.Context, opts RunOptions) (*process.Handle, error)
}

// devops/sources/
type ImageSource interface {
    Name() string
    Available() bool
    Download(ctx context.Context, name, version string, progress func(downloaded, total int64)) (string, error)
}

// build/signing/
type Signer interface {
    Name() string
    Available() bool
    Sign(filePath, keyID string) ([]byte, error)
}

// sdk/generators/
type Generator interface {
    Language() string
    Generate(ctx context.Context, spec, outputDir string, config *Config) error
}
```

## External Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/Snider/Borg` | Compression (xz) for archives. **Not** Secure/Blob/Pointer. |
| `github.com/getkin/kin-openapi` | OpenAPI 3.x spec parsing |
| `github.com/oasdiff/oasdiff` | API breaking change detection |
| `github.com/kluctl/go-embed-python` | Embedded Python 3.13 runtime |
| `github.com/spf13/cobra` | CLI framework for build/release commands |
| `golang.org/x/crypto` | SSH connections (ansible/) |

## Configuration Files

- `.core/build.yaml` — Build targets, ldflags, signing, archive format
- `.core/release.yaml` — Version source, changelog style, SDK langs, publisher configs
- `infra.yaml` — Host inventory, DNS zones, cloud provider settings
- `~/.core/config.yaml` — Local dev environment config

## Coding Standards

- **UK English**: colour, organisation, centre
- **Tests**: testify assert/require, `_Good`/`_Bad`/`_Ugly` naming convention
- **Conventional commits**: `feat(ansible):`, `fix(infra):`, `refactor(build):`
- **Co-Author**: `Co-Authored-By: Virgil <virgil@lethean.io>`
- **Licence**: EUPL-1.2
- **Imports**: stdlib → forge.lthn.ai → third-party, each group separated by blank line

## Forge

- **Repo**: `forge.lthn.ai/core/go-devops`
- **Push via SSH**: `git push forge main` (remote: `ssh://git@forge.lthn.ai:2223/core/go-devops.git`)

## Task Queue

See `TODO.md` for prioritised work. See `FINDINGS.md` for research notes.
