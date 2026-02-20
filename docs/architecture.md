# Architecture — go-devops

`forge.lthn.ai/core/go-devops` is an infrastructure and build automation library written in Go. It provides native Go implementations of Ansible playbook execution, multi-target build pipelines, release automation, infrastructure API clients, container/VM management, SDK generation, and a developer toolkit with static analysis capabilities. The library is approximately 29,000 lines of source across 71 source files.

## Package Map

```
go-devops/
├── ansible/        Ansible playbook execution engine (native Go, no shell-out)
├── build/          Build system: project detection, archives, checksums
│   ├── builders/   Plugin implementations: Go, Wails, Docker, C++, LinuxKit, Taskfile
│   ├── signing/    Code signing: macOS codesign, GPG, Windows signtool
│   └── buildcmd/   Cobra CLI handlers for core build / core release
├── container/      LinuxKit VM management, hypervisor abstraction, state
├── deploy/         Deployment integrations
│   ├── python/     Embedded Python 3.13 runtime
│   └── coolify/    Coolify PaaS API client (via Python Swagger)
├── devkit/         Developer toolkit: quality metrics, security, coverage trending
├── devops/         Portable dev environment management
│   └── sources/    Image download sources: GitHub Releases, S3/CDN
├── infra/          Infrastructure APIs: Hetzner Cloud, Hetzner Robot, CloudNS
├── release/        Release orchestration: version, changelog, publishing
│   └── publishers/ Platform publishers: GitHub, Docker, Homebrew, npm, AUR, Scoop, Chocolatey, LinuxKit
└── sdk/            OpenAPI SDK generation and breaking-change detection
    └── generators/ Language generators: TypeScript, Python, Go, PHP
```

## Core Interfaces

Every extensible subsystem is defined by a small interface.

```go
// build/builders — project type plugin
type Builder interface {
    Name() string
    Detect(fs io.Medium, dir string) (bool, error)
    Build(ctx context.Context, cfg *Config, targets []Target) ([]Artifact, error)
}

// release/publishers — distribution target plugin
type Publisher interface {
    Name() string
    Publish(ctx context.Context, release *Release, pubCfg PublisherConfig, relCfg ReleaseConfig, dryRun bool) error
}

// container — hypervisor abstraction
type Hypervisor interface {
    Name() string
    Available() bool
    Run(ctx context.Context, opts RunOptions) (*process.Handle, error)
}

// devops/sources — image download plugin
type ImageSource interface {
    Name() string
    Available() bool
    Download(ctx context.Context, name, version string, progress func(downloaded, total int64)) (string, error)
}

// build/signing — code signing plugin
type Signer interface {
    Name() string
    Available() bool
    Sign(filePath, keyID string) ([]byte, error)
}

// sdk/generators — language SDK generator
type Generator interface {
    Language() string
    Generate(ctx context.Context, spec, outputDir string, config *Config) error
}
```

---

## Ansible Integration

### Overview

The `ansible/` package executes Ansible playbooks natively in Go without shelling out to `ansible-playbook`. It implements the Ansible execution model — facts, handlers, `register`, `when`, `loop`, `become` — over SSH connections managed with `golang.org/x/crypto/ssh`.

### Package Files

| File | LOC | Responsibility |
|------|-----|---------------|
| `executor.go` | 1,021 | Playbook runner: task dispatch, handler firing, fact injection, become/sudo |
| `modules.go` | 1,434 | Module implementations: ~30 Ansible modules in pure Go |
| `parser.go` | 438 | YAML playbook + inventory parser |
| `ssh.go` | 451 | SSH client with persistent connection and file upload |
| `types.go` | 258 | Core types: Play, Task, Handler, Inventory, Facts |

### Data Model

A `Playbook` contains one or more `Play` structs. Each play targets a host pattern from the `Inventory` and runs a list of `Task` structs. Tasks carry a `Module` name (derived from the YAML key), `Args` map, and optional control fields (`when`, `register`, `loop`, `notify`, `become`).

```
Playbook
└── []Play
    ├── Hosts      string          (host pattern: "webservers", "all")
    ├── Become     bool
    ├── Vars       map[string]any
    ├── PreTasks   []Task
    ├── Tasks      []Task
    ├── PostTasks  []Task
    ├── Handlers   []Task
    └── Roles      []RoleRef
```

`TaskResult` carries the Ansible result contract: `Changed`, `Failed`, `Skipped`, `Stdout`, `Stderr`, `RC`, and a `Data` map for module-specific output.

### Execution Model

1. Parser reads a YAML playbook file and builds the `Playbook` struct.
2. The inventory is parsed from a separate YAML file (or from the play's `vars`).
3. For each play, the executor resolves target hosts from the `Inventory`.
4. If `gather_facts` is enabled, the executor SSHs to each host, reads `/etc/os-release`, and populates a `Facts` struct.
5. Tasks are executed in order. For each task:
   - `when` conditionals are evaluated using Go template logic and registered variables.
   - `loop` items are resolved and the module is called once per item.
   - The module function is dispatched via a string-keyed registry that normalises both long (`ansible.builtin.shell`) and short (`shell`) module names.
   - `register` stores the `TaskResult` in a variable map for subsequent `when` and template evaluations.
   - `notify` queues handler names; handlers fire once at the end of the play if any task triggered them.
6. `become: true` prefixes commands with `sudo -u <become_user>`.

### Implemented Modules

Modules are grouped by category:

- **Command execution**: `command`, `shell`, `raw`, `script`
- **File operations**: `copy`, `template`, `file`, `lineinfile`, `blockinfile`, `stat`
- **Package management**: `apt`, `apt_key`, `apt_repository`, `yum`, `dnf`, `package`, `pip`
- **Service management**: `service`, `systemd`
- **User and group**: `user`, `group`, `authorized_key`, `cron`
- **Source control**: `git`, `unarchive`
- **Network**: `uri`, `get_url`
- **Firewall**: `ufw`
- **Container**: `docker_compose`
- **Control flow**: `debug`, `fail`, `assert`, `set_fact`, `include_vars`, `wait_for`, `pause`, `meta`

### SSH Layer

`ssh.go` manages a persistent `*ssh.Client` per host. It exposes three operations:

- `Run(cmd string) (stdout, stderr string, rc int, err error)` — executes a command
- `RunScript(script string) (...)` — uploads a temporary script and runs it
- `Upload(localPath, remotePath string) error` — SCP-style file upload via SFTP subsystem

Connection parameters (host, port, user, private key file) are drawn from the `Host` struct in the inventory.

---

## Build Pipeline

### Overview

The `build/` package provides project-type detection and cross-compilation. Configuration is read from `.core/build.yaml`. The `buildcmd/` sub-package registers Cobra commands (`core build`, `core build pwa`, `core build sdk`, `core release`) into the main CLI.

### Project Detection

`discovery.go` probes marker files in priority order:

| Marker file | Project type |
|-------------|-------------|
| `wails.json` | `wails` |
| `go.mod` | `go` |
| `package.json` | `node` |
| `composer.json` | `php` |
| `CMakeLists.txt` | `cpp` |
| `Dockerfile` | `docker` |
| `*.yml` (linuxkit pattern) | `linuxkit` |
| `Taskfile.yml` | `taskfile` |

### Build Targets and Artifacts

A `Target` carries `OS` and `Arch` (matching `GOOS`/`GOARCH`). Each builder produces `[]Artifact`, where each artifact has a file `Path`, `OS`, `Arch`, and a SHA-256 `Checksum`. Checksums are computed and stored in `dist/*.sha256` files.

### Archive Creation

`archive.go` packages build outputs using `github.com/Snider/Borg` for xz compression. Supported formats: `tar.gz`, `tar.xz`, `zip`. The Borg dependency is used only for xz support; it does not use the Secure/Blob or Secure/Pointer features.

### Builder Plugins

Each builder implements `Builder.Detect()` to self-identify for a directory and `Builder.Build()` to produce artifacts.

| Builder | Notes |
|---------|-------|
| `go.go` | `go build` with ldflags injection, cross-compilation via `GOOS`/`GOARCH` env |
| `wails.go` | Wails v3 desktop app build, platform-specific packaging |
| `docker.go` | `docker buildx` with multi-platform support, optional push |
| `cpp.go` | CMake configure + build in a temp directory |
| `linuxkit.go` | LinuxKit YAML config → multi-format VM images (iso, qcow2, raw, vmdk) |
| `taskfile.go` | Delegates to `task` CLI with target mapping |

### Code Signing

`signing/` implements a `Signer` interface with three backends:

| Signer | Platform | Tool |
|--------|----------|------|
| macOS | darwin | `codesign` |
| GPG | any | `gpg --detach-sign` |
| Windows | windows | `signtool` |

`Available()` checks whether the required tool exists at runtime. Signing is applied to binary artifacts after build, before archiving.

---

## Infrastructure Management

### Overview

The `infra/` package provides typed API clients for Hetzner Cloud (VPS), Hetzner Robot (bare metal), and CloudNS DNS. All three share a common `APIClient` with exponential backoff, rate-limit handling, and configurable authentication.

### Shared API Client

`client.go` defines `APIClient`:

```go
type APIClient struct {
    client       *http.Client
    retry        RetryConfig
    authFn       func(req *http.Request)
    prefix       string
    mu           sync.Mutex
    blockedUntil time.Time
}
```

- `Do(req, result)` — executes a request with JSON decoding.
- `DoRaw(req)` — executes a request and returns raw bytes.
- Both methods apply: auth injection, rate-limit window respect, exponential backoff with jitter on 5xx and transport errors. 4xx errors (except 429) are not retried.

**Retry configuration** (`RetryConfig`):

| Field | Default |
|-------|---------|
| `MaxRetries` | 3 |
| `InitialBackoff` | 100 ms |
| `MaxBackoff` | 5 s |

**Rate limiting**: on HTTP 429, the `Retry-After` header (seconds format) is parsed and a `blockedUntil` timestamp is set. All subsequent requests on the same `APIClient` instance wait until that timestamp before proceeding. Context cancellation is honoured during the wait.

### Hetzner Cloud Client

`hetzner.go` wraps the Hetzner Cloud v1 API (Bearer token auth). Supported resources: servers, load balancers, networks, volumes, SSH keys, firewalls. The Hetzner Robot client (Basic Auth) supports bare metal servers.

### CloudNS Client

`cloudns.go` wraps the CloudNS API v1 (auth-id + auth-password query parameters). Supports: zone listing, record CRUD, ACME DNS-01 challenge records. Used by the `infra` provisioning pipeline for automatic TLS certificate issuance via Let's Encrypt.

### Infrastructure Configuration

`infra.yaml` (project root) defines the full host inventory and cloud resources:

```yaml
hosts:
  - name: de1
    ip: 1.2.3.4
    provider: hetzner-robot
    roles: [web, db]

dns:
  provider: cloudns
  zones: [lthn.ai, leth.in]

loadbalancers:
  - name: lb-de1
    provider: hetzner-cloud
```

The `config.go` file in `infra/` parses this into typed structs: `Host`, `LoadBalancer`, `Network`, `DNSZone`, `Database`, `Cache`.

---

## Release Workflow

### Overview

The `release/` package orchestrates the full release pipeline: version detection, changelog generation from git history, and publishing to multiple distribution targets. Configuration is read from `.core/release.yaml`.

### Version Detection

`DetermineVersion(dir string)` checks in priority order:

1. Git tag on `HEAD` (exact match).
2. Most recent tag with patch increment (`IncrementVersion`).
3. Default `v0.0.1` if no tags exist.

`IncrementVersion` parses semver and increments the patch component, stripping any pre-release suffix.

### Changelog Generation

`changelog.go` (`Generate` function) reads git log since the previous tag and formats commits into a markdown changelog. Conventional commit prefixes (`feat:`, `fix:`, `refactor:`, etc.) are parsed to group entries.

### Release Orchestration

`Publish(ctx, cfg, dryRun)`:

1. Resolves version.
2. Scans `dist/` for pre-built artifacts (built by `core build`).
3. Generates changelog.
4. Iterates configured publishers, calling `Publisher.Publish()` on each.
5. Returns a `*Release` struct with version, artifacts, and changelog.

The separation of `core build` and `core release` allows CI pipelines to build once and publish to multiple targets independently.

### Publishers

All publishers implement `Publisher.Publish(ctx, release, pubCfg, relCfg, dryRun)`. When `dryRun` is true, publishers log what they would do without making external calls.

| Publisher | Distribution method |
|-----------|-------------------|
| `github.go` | GitHub Releases API — creates release, uploads artifact files |
| `docker.go` | `docker buildx build --push` to configured registry |
| `homebrew.go` | Generates a Ruby formula file, commits to a tap repository |
| `npm.go` | `npm publish` to the npm registry |
| `aur.go` | Generates `PKGBUILD` + `.SRCINFO`, pushes to AUR git remote |
| `scoop.go` | Generates a JSON manifest, commits to a Scoop bucket |
| `chocolatey.go` | Generates `.nuspec`, calls `choco push` |
| `linuxkit.go` | Builds and uploads LinuxKit multi-format VM images |

---

## Container and VM Management

### Overview

The `container/` package manages LinuxKit-based VM images. It abstracts hypervisor differences (QEMU on Linux, Hyperkit on macOS) and persists container state to `~/.core/state.json`.

### Hypervisor Abstraction

`hypervisor.go` auto-selects the backend at runtime:

- `Available()` checks for the hypervisor binary in `PATH`.
- Linux: QEMU (`qemu-system-x86_64` / `qemu-system-aarch64`).
- macOS: Hyperkit (`hyperkit`).

### State Persistence

`state.go` serialises container records to `~/.core/state.json`. Each record includes the container name, LinuxKit image path, hypervisor PID, and network configuration.

### Template Rendering

`templates.go` renders Packer and LinuxKit YAML templates using Go `text/template`, substituting image name, kernel version, architecture, and port mappings.

---

## DevKit — Developer Toolkit

### Overview

The `devkit/` package provides code quality, security, and metrics functions exposed as a `Toolkit` struct. All methods operate on a working directory passed to `New(dir)`.

### Code Quality

| Function | Description |
|----------|-------------|
| `FindTODOs(dir)` | Uses `git grep` to locate `TODO`, `FIXME`, `HACK` comments |
| `Lint(pkg)` | Runs `go vet` and parses findings |
| `TestCount(pkg)` | Lists test functions via `go test -list` |
| `Coverage(pkg)` | Runs `go test -cover` and parses per-package percentages |
| `RaceDetect(pkg)` | Runs `go test -race` and extracts `DATA RACE` reports |
| `Build(targets...)` | Compiles targets, returns `BuildResult` with any errors |
| `ModTidy()` | Runs `go mod tidy` |

### Security

| Function | Description |
|----------|-------------|
| `AuditDeps()` | Runs `govulncheck ./...` (human-readable), parses `Vulnerability #` blocks |
| `VulnCheck(modulePath)` | Runs `govulncheck -json`, parses newline-delimited JSON into `VulnFinding` structs |
| `ScanSecrets(dir)` | Runs `gitleaks detect --report-format csv`, parses CSV output |
| `CheckPerms(dir)` | Walks directory tree, flags world-writable files |

### Vulnerability Scanning Detail

`VulnCheck` produces structured `VulnFinding` values:

```go
type VulnFinding struct {
    ID             string   // GO-2024-xxxx
    Aliases        []string // CVE/GHSA identifiers
    Package        string   // Affected package path
    CalledFunction string   // Function in call stack
    Description    string   // OSV summary
    FixedVersion   string   // Minimum fixed version
    ModulePath     string   // Go module path
}
```

`ParseVulnCheckJSON` correlates `finding` messages with `osv` metadata messages from govulncheck's JSON stream. It skips malformed lines gracefully (govulncheck occasionally emits non-JSON progress lines).

### Cyclomatic Complexity Analysis

`AnalyseComplexity(cfg ComplexityConfig)` walks Go source files using `go/ast` without external tools:

- Default threshold: 15.
- Skips `vendor/`, hidden directories, and `_test.go` files.
- Counts branching constructs: `if`, `for`, `range`, `case` (non-default), `select` comm clause, `&&`, `||`, type switch, select statement.
- Returns `[]ComplexityResult` with function name, package, file, line, and score.

`AnalyseComplexitySource(src, filename, threshold)` accepts source as a string for in-memory analysis.

### Coverage Trending

Three complementary functions handle coverage over time:

- `ParseCoverProfile(data)` — parses `go test -coverprofile` format; computes per-package statement ratios.
- `ParseCoverOutput(output)` — parses human-readable `go test -cover ./...` output.
- `CompareCoverage(previous, current)` — diffs two `CoverageSnapshot` values, returning regressions, improvements, new packages, and removed packages.

`CoverageStore` persists snapshots as a JSON array at a configurable file path, with `Append`, `Load`, and `Latest` methods.

### Git and Metrics

| Function | Description |
|----------|-------------|
| `DiffStat()` | Parses `git diff --stat` summary |
| `UncommittedFiles()` | Lists files with uncommitted changes via `git status --porcelain` |
| `GitLog(n)` | Returns last `n` commits as structured `Commit` values |
| `DepGraph(pkg)` | Parses `go mod graph` into a `Graph{Nodes, Edges}` |
| `Complexity(threshold)` | Wraps external `gocyclo` tool (distinct from `AnalyseComplexity` which uses `go/ast`) |

---

## SDK Generation

### Overview

The `sdk/` package auto-detects an OpenAPI specification, generates typed client libraries in up to four languages, and detects breaking changes between spec versions using oasdiff.

### Spec Detection

`DetectSpec(dir)` checks locations in priority order:

1. Path configured in `.core/release.yaml`.
2. Common paths: `openapi.yaml`, `openapi.json`, `api/openapi.yaml`, `docs/openapi.yaml`, and four others.

### Language Generators

Each generator implements the `Generator` interface. Supported languages: TypeScript, Python, Go, PHP. Generator registration uses a string-keyed registry allowing overrides.

### Breaking Change Detection

`DetectBreakingChanges(baseSpec, revisionSpec)` uses `github.com/oasdiff/oasdiff` to compare two spec files. Returns a `DiffResult` with a human-readable summary and a list of individual breaking changes. Exit codes: 0 = no changes, 1 = non-breaking changes, 2 = breaking changes.

---

## Configuration Files

| File | Location | Purpose |
|------|----------|---------|
| `.core/build.yaml` | Project root | Build targets, ldflags, signing, archive format |
| `.core/release.yaml` | Project root | Version source, changelog style, SDK languages, publisher configs |
| `infra.yaml` | Project root | Host inventory, DNS zones, cloud provider credentials |
| `~/.core/config.yaml` | User home | Local dev environment configuration |
| `~/.core/state.json` | User home | Container/VM runtime state |

---

## External Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/Snider/Borg` | xz compression for build archives |
| `github.com/getkin/kin-openapi` | OpenAPI 3.x spec parsing |
| `github.com/oasdiff/oasdiff` | API breaking change detection |
| `github.com/kluctl/go-embed-python` | Embedded Python 3.13 runtime for Coolify client |
| `github.com/spf13/cobra` | CLI framework for `build/buildcmd/` |
| `golang.org/x/crypto` | SSH connections in `ansible/ssh.go` |
| `gopkg.in/yaml.v3` | Playbook and config YAML parsing |
| `github.com/stretchr/testify` | Test assertions |

## Dependency on `forge.lthn.ai/core/go`

The parent framework is referenced via a `replace` directive in `go.mod`:

```
replace forge.lthn.ai/core/go => ../core
```

Provides: `core.E` (contextual errors), `io.Medium` (file system abstraction), config, logging, and i18n utilities.
