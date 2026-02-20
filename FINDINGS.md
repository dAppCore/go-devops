# FINDINGS.md — go-devops Research & Discovery

## 2026-02-20: Initial Analysis (Virgil)

### Origin

Extracted from `core/go` on 16 Feb 2026 (commit `392ad68`). Single extraction commit — fresh repo.

### Package Inventory

| Package | Files | Source LOC | Test Files | Notes |
|---------|-------|-----------|-----------|-------|
| `ansible/` | 5 | 3,162 | 1 | Playbook executor, SSH, modules, parser |
| `build/` | 6 | 797 | 4 | Project detection, archives, checksums, config |
| `build/builders/` | 6 | 1,390 | — | Go, Wails, Docker, C++, LinuxKit, Taskfile |
| `build/signing/` | 5 | 377 | — | macOS, GPG, Windows signtool |
| `build/buildcmd/` | 6 | 1,053 | — | CLI command handlers |
| `container/` | 5 | 1,208 | 4 | LinuxKit VMs, hypervisor abstraction, state |
| `deploy/python/` | 1 | 147 | — | Embedded Python 3.13 |
| `deploy/coolify/` | 1 | 219 | — | Coolify PaaS API client |
| `devkit/` | 1 | 560 | 1 | Code quality metrics |
| `devops/` | 8 | 1,216 | 8 | Dev environment manager |
| `devops/sources/` | 3 | 218 | — | GitHub/CDN image sources |
| `infra/` | 3 | 953 | 1 | Hetzner, CloudNS, config |
| `release/` | 5 | 1,398 | 5 | Release orchestrator |
| `release/publishers/` | 9 | 2,610 | 9 | 8 target platforms |
| `sdk/` | 3 | 494 | 3 | OpenAPI detection + diff |
| `sdk/generators/` | 5 | 437 | 5 | 4-language SDK gen |

**Total**: ~29K LOC across 71 source files + 47 test files

### Key Observations

1. **ansible/modules.go is the largest file** — 1,434 LOC implementing Ansible modules in pure Go. Zero tests. Highest-priority testing gap.

2. **Borg dependency is compression-only** — `github.com/Snider/Borg` used for xz archive creation in `build/archive.go`. Does NOT use the Secure/Blob/Pointer features.

3. **Python 3.13 embedded** — `deploy/python/` embeds a full Python runtime via kluctl/go-embed-python. Used exclusively for Coolify API client (Python Swagger). Consider replacing with native Go HTTP client to remove the 50MB+ Python dependency.

4. **DigitalOcean gap** — Referenced in `infra/config.go` types but no `digitalocean.go` implementation exists. Either implement or remove the dead types.

5. **Single-commit repo** — Entire codebase arrived in one `feat: extract` commit. No git history for individual components. This makes blame/bisect impossible for bugs originating before extraction.

6. **Hypervisor platform detection** — `container/hypervisor.go` auto-selects QEMU on Linux, Hyperkit on macOS. Both are platform-specific — tests may need build tags or mocking.

7. **CLI via Cobra** — `build/buildcmd/` uses Cobra directly (not core/go's CLI framework). May need alignment.

8. **8 release publishers** — GitHub, Docker, Homebrew, npm, AUR, Scoop, Chocolatey, LinuxKit. All implement the `Publisher` interface. Each is ~250-370 LOC. All have test files.

### Test Coverage Gaps

| Package | Gap Severity | Notes |
|---------|-------------|-------|
| `ansible/modules.go` | **Critical** | 1,434 LOC, zero tests |
| `ansible/executor.go` | **Critical** | 1,021 LOC, zero tests |
| `ansible/parser.go` | High | 438 LOC, zero tests |
| `infra/hetzner.go` | High | 381 LOC, zero tests — API calls untested |
| `infra/cloudns.go` | High | 272 LOC, zero tests — DNS ops untested |
| `build/builders/*` | Medium | 1,390 LOC, no individual builder tests |
| `build/signing/*` | Medium | 377 LOC, signing logic untested |
| `deploy/*` | Low | 366 LOC, Python/Coolify integration |

### Integration Points

- **core/go** → Framework (core.E, io.Medium, config, logging)
- **core/go-crypt** → SSH key management (ansible/ssh.go uses golang.org/x/crypto directly, could use go-crypt)
- **core/cli** → Build/release commands registered via Cobra
- **DevOps repo** → `infra.yaml` config used by Ansible playbooks in `/Users/snider/Code/DevOps`

### Config File Ecosystem

| File | Location | Purpose |
|------|----------|---------|
| `.core/build.yaml` | Project root | Build targets, signing, archives |
| `.core/release.yaml` | Project root | Version, changelog, publishers |
| `infra.yaml` | Project root | Host inventory, DNS, cloud providers |
| `~/.core/config.yaml` | User home | Local dev environment config |
| `~/.core/state.json` | User home | Container/VM state persistence |
