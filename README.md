# go-devops

Infrastructure and build automation library for the Lethean ecosystem. Provides a native Go Ansible playbook executor (~30 modules over SSH without shelling out), a multi-target build pipeline with project type auto-detection (Go, Wails, Docker, C++, LinuxKit, Taskfile), code signing (macOS codesign, GPG, Windows signtool), release orchestration with changelog generation and eight publisher backends (GitHub Releases, Docker, Homebrew, npm, AUR, Scoop, Chocolatey, LinuxKit), Hetzner Cloud and Robot API clients, CloudNS DNS management, container/VM management via QEMU and Hyperkit, an OpenAPI SDK generator (TypeScript, Python, Go, PHP), and a developer toolkit with cyclomatic complexity analysis, vulnerability scanning, and coverage trending.

**Module**: `forge.lthn.ai/core/go-devops`
**Licence**: EUPL-1.2
**Language**: Go 1.26

> **Service registration**: go-devops is intentionally a multi-Service repo per Mantis #1336/#1379. Each independently-stateful subpackage exposes its own canonical `NewService(opts) + Register(c)` pair under its own package name (`dev`, `devkit`, `coolify`). Pure-utility subpackages (`snapshot`, `deploy/python`) intentionally have no Service wiring.

## Quick Start

```go
import (
    "forge.lthn.ai/core/go-devops/ansible"
    "forge.lthn.ai/core/go-devops/build"
    "forge.lthn.ai/core/go-devops/release"
)

// Run an Ansible playbook over SSH
pb, _ := ansible.ParsePlaybook("playbooks/deploy.yml")
inv, _ := ansible.ParseInventory("inventory.yml")
pb.Run(ctx, inv)

// Build and release
artifacts, _ := build.Build(ctx, ".")
release.Publish(ctx, releaseCfg, false)
```

## Documentation

- [Architecture](docs/architecture.md) — Ansible integration, build pipeline, infrastructure APIs, release workflow, devkit, SDK generation
- [Development Guide](docs/development.md) — building, testing, coding standards
- [Project History](docs/history.md) — completed phases and known limitations

## Build & Test

```bash
go test ./...
go test -race ./...
go vet ./...
go build ./...
```

## Licence

European Union Public Licence 1.2 — see [LICENCE](LICENCE) for details.
