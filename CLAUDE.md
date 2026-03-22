# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
go test ./...                    # Run all tests
go test -v -run TestName ./...   # Single test
go test -race ./...              # Race detector
go vet ./...                     # Static analysis
```

## Workspace Context

This module (`dappco.re/go/core/devops`) is part of a 57-module Go workspace rooted at `/Users/snider/Code/go.work`. The parent framework module `forge.lthn.ai/core/go` (at `../go`) provides core libraries: `core.E` errors, `io.Medium` filesystem abstraction, config, i18n, and logging.

Most implementation code (ansible engine, build system, infra clients, release pipeline, devkit, SDK generators) lives in the parent framework. This repo contains CLI commands that wire those packages together, plus deployment integrations and infrastructure playbooks.

## Architecture

### Package Layout

- **`cmd/dev/`** — Multi-repo developer commands registered under `core dev`. The main CLI surface (~4,700 LOC across 21 files).
- **`cmd/deploy/`** — `core deploy servers` — Coolify PaaS server/app listing.
- **`cmd/docs/`** — `core docs sync` — Documentation sync across the multi-repo workspace.
- **`cmd/setup/`** — `core setup repo` — Generate `.core` configuration for a project.
- **`cmd/gitcmd/`** — Git helper commands (mirrors dev commands under `core git`).
- **`cmd/vanity-import/`** — Vanity import path server (the default build target in `.core/build.yaml`).
- **`cmd/community/`** — Community landing page assets.
- **`deploy/coolify/`** — Coolify PaaS API HTTP client.
- **`deploy/python/`** — Embedded Python 3.13 runtime wrapper (adds ~50 MB to binary).
- **`locales/`** — Embedded i18n translation files (en.json).
- **`snapshot/`** — `core.json` release manifest generation.
- **`playbooks/`** — Ansible YAML playbooks for production infrastructure (Galera, Redis). Executed by the native Go Ansible engine, not `ansible-playbook`.

### Key CLI Commands (`cmd/dev/`)

| Command | Purpose |
|---------|---------|
| `core dev work` | Combined git status/commit/push workflow |
| `core dev commit` | Claude-assisted commit generation |
| `core dev push/pull` | Push/pull repos with pending changes |
| `core dev issues` | List open Forgejo issues |
| `core dev reviews` | List PRs needing review |
| `core dev ci` | Check CI/workflow status |
| `core dev impact` | Analyse dependency impact across workspace |
| `core dev vm` | Boot, stop, shell, serve dev environments |
| `core dev workflow` | List/sync CI workflows across repos |
| `core dev file-sync` | Safe file sync for AI agents |
| `core dev apply` | Apply safe changes (AI-friendly) |

### Extension Interfaces (in parent framework)

All extensible subsystems follow a plugin/provider pattern with small interfaces:

| Interface | Package | Implementations |
|-----------|---------|-----------------|
| `Builder` | `build/builders/` | Go, Wails, Docker, C++, LinuxKit, Taskfile |
| `Publisher` | `release/publishers/` | GitHub, Docker, Homebrew, npm, AUR, Scoop, Chocolatey, LinuxKit |
| `Signer` | `build/signing/` | macOS codesign, GPG, Windows signtool |
| `Hypervisor` | `container/` | QEMU (Linux), Hyperkit (macOS) |
| `ImageSource` | `devops/sources/` | GitHub Releases, S3/CDN |
| `Generator` | `sdk/generators/` | TypeScript, Python, Go, PHP |

### Shared API Client Pattern

`infra/client.go` (parent module) provides HTTP client abstraction with exponential backoff retry (3 retries, 100ms–5s), HTTP 429 rate-limit handling with Retry-After parsing, and configurable auth (Bearer, Basic, query params). Used by Hetzner Cloud/Robot, CloudNS, and Forgejo clients.

### Build & Release Flow

`core build` → auto-detects project type → produces artifacts in `dist/`
`core build release` → version detection → changelog generation → publish via configured publishers

Configuration lives in `.core/build.yaml` (targets, ldflags) and `.core/release.yaml` (publishers, changelog filters).

## Coding Standards

- **UK English**: colour, organisation, centre
- **Tests**: testify assert/require, `_Good`/`_Bad`/`_Ugly` naming convention
- **Conventional commits**: `feat(ansible):`, `fix(infra):`, `refactor(build):`
- **Co-Author**: `Co-Authored-By: Virgil <virgil@lethean.io>`
- **Licence**: EUPL-1.2
- **Imports**: stdlib → forge.lthn.ai → third-party, each group separated by blank line
- **Errors**: `log.E(op, msg, err)` from `go-log` for all contextual errors (never `fmt.Errorf` or `errors.New`)

## Forge

- **Repo**: `dappco.re/go/core/devops` (hosted at `forge.lthn.ai/core/go-devops`)
- **Push via SSH**: `git push forge main` (remote: `ssh://git@forge.lthn.ai:2223/core/go-devops.git`)
- **Issues/PRs**: Managed via Forgejo SDK (`code.gitea.io/sdk/gitea`), not GitHub

## Documentation

- Architecture deep-dive: `docs/architecture.md`
- Development guide & testing patterns: `docs/development.md`
- Project history & known limitations: `docs/history.md`
