# Project History — go-devops

## Origin

`go-devops` was extracted from the `forge.lthn.ai/core/go` monorepo on 16 February 2026. The entire codebase arrived in a single extraction commit and was pushed to its own Forge repository (`core/go-devops`). This means git blame and bisect cannot distinguish the history of individual components prior to the extraction date; all pre-extraction bugs are outside the revision graph.

**Extraction commit**: the repository's first commit (`feat: extract`) contains the full initial codebase — approximately 29,000 lines across 71 source files and 47 test files spanning 16 packages.

---

## Phase 0: Test Coverage and Hardening

**Commit**: `6e346cb`, `5d22ed9`

**Scope**: Established a baseline test suite across the packages with the most critical coverage gaps at extraction.

### Completed work

- **ansible/ tests** — Added `parser_test.go` (17 tests covering `ParsePlaybook`, `ParseInventory`, `ParseTasks`, `GetHosts`, `GetHostVars`, `isModule`, `NormalizeModule`), `types_test.go` (covering `RoleRef`/`Task` `UnmarshalYAML`, `Inventory`, `Facts`, `TaskResult`, `KnownModules`), and `executor_test.go` (covering `getHosts`, `matchesTags`, `evaluateWhen`, `templateString`, `applyFilter`, `resolveLoop`, `templateArgs`, `handleNotify`, `normalizeConditions`, and helper functions).

- **infra/ tests** — Added `hetzner_test.go` (covering `HCloudClient`/`HRobotClient` construction, `do()` round-trip via `httptest`, API error handling, and JSON serialisation for `HCloudServer`, `HCloudLoadBalancer`, `HRobotServer`) and `cloudns_test.go` (covering `doRaw()` round-trip, zone/record JSON, CRUD responses, ACME challenge, auth parameters, and errors).

- **build/ tests** — Added `archive_test.go` (249 LOC, archive round-trip for tar.gz and zip, multi-file archives) and extended `signing_test.go` (+181 LOC with mock signer tests, path verification, and error handling).

- **release/ nil guard** — Fixed a nil pointer crash in `release/publishers/linuxkit.go` line 50. Added a `release.FS == nil` guard. Added a corresponding nil FS test case to `linuxkit_test.go` (+23 LOC). Total test count across build/ and release/ reached 862.

- **Race detector** — `go test -race ./...` confirmed clean across `ansible/`, `infra/`, `container/`, `devops/`, and `build/` packages.

- **`go vet ./...`** — Fixed stale API calls in `container/linuxkit_test.go`, `state_test.go`, `templates_test.go`, and `devops/devops_test.go`. Fixed the `go.mod` `replace` directive.

---

## Phase 1: Ansible Engine Hardening

**Commits**: `3330e55`, `c7da9ad`, `9638e77`, `427929f`, `8ab8643`

**Scope**: Brought the Ansible engine from zero test coverage to comprehensive coverage across all module categories, SSH infrastructure, and executor logic.

### Step 1.0: SSH mock infrastructure (`3330e55`)

Created `ansible/mock_ssh_test.go` providing:
- `MockSSHClient` with a command registry (`expectCommand`), in-memory filesystem, become-state tracking, and execution/upload logs.
- Assertion helpers: `hasExecuted`, `hasExecutedMethod`, `findExecuted`.
- Module shims via the `sshRunner` interface to decouple module functions from real SSH connections.
- 12 mock infrastructure tests confirming the mock behaves correctly in isolation.

### Step 1.1: Command execution modules (`3330e55`)

36 module tests covering `command`, `shell`, `raw`, and `script`. Verified: `command` uses `Run()`, `shell` uses `RunScript()`, `raw` passes through unmodified, `script` reads a local file before uploading. Cross-module differentiation and dispatch routing tests included. Total ansible tests at this point: 48.

### Step 1.2: File operation modules (`c7da9ad`)

54 new tests across `copy` (8), `file` (12), `lineinfile` (8), `blockinfile` (7), `stat` (5), `template` (6), dispatch (6), and integration (2). Extended the mock with an `sshFileRunner` interface and 6 module shims. Fixed an unsupported-module test (copy to hostname). Total ansible tests: 208.

### Step 1.3: Service and package modules (`9638e77`)

56 new tests across `service` (12), `systemd` (4), `apt` (9), `apt_key` (6), `apt_repository` (8), `package` (3), `pip` (8), and dispatch (7). 7 new module shims added to `mock_ssh_test.go`.

### Step 1.4: User, group, and advanced modules (`427929f`)

69 new tests across `user` (7), `group` (7), `cron` (5), `authorized_key` (7), `git` (8), `unarchive` (8), `uri` (6), `ufw` (8), `docker_compose` (7), and dispatch (6). 9 module shims. Total ansible tests: 334.

### Step 1.5: Error propagation, become, facts, idempotency (`8ab8643`)

- **Error propagation** — 68 tests across `getHosts`, `matchesTags`, `evaluateWhen`/`evalCondition`, `templateString`, `applyFilter`, `resolveLoop`, `handleNotify`, `normalizeConditions`, and cross-cutting scenarios.
- **Become/sudo** — 8 tests: enable/disable cycle, default user, passwordless sudo, play-level become.
- **Fact gathering** — 9 tests: Ubuntu, CentOS, Alpine, and Debian `os-release` parsing, hostname, and localhost behaviour.
- **Idempotency checks** — 8 tests: group exists, authorised key present, Docker Compose up-to-date, stat always reports unchanged.
- Total ansible tests at phase completion: 438.

---

## Phase 2: Infrastructure API Robustness

**Commit**: included in Phase 2 work

**Scope**: Consolidated three separate API clients behind a shared `APIClient` abstraction and added retry and rate-limit handling.

### Completed work

- **API client abstraction** — Extracted shared `APIClient` struct in `infra/client.go`. `HCloudClient`, `HRobotClient`, and `CloudNSClient` now delegate to `APIClient` via configurable auth functions and error prefixes. Options pattern: `WithHTTPClient`, `WithRetry`, `WithAuth`, `WithPrefix`. Added 30 `client_test.go` tests.

- **Retry logic** — `APIClient` implements exponential backoff with jitter. Retries on 5xx responses and transport errors. Does not retry 4xx errors (except 429). `RetryConfig` carries `MaxRetries` (default 3), `InitialBackoff` (100 ms), and `MaxBackoff` (5 s). Context cancellation is respected during backoff sleeps.

- **Rate limiting** — Detects HTTP 429 responses, parses `Retry-After` header (seconds format; falls back to 1 s). Sets a per-`APIClient` `blockedUntil` timestamp guarded by a mutex. All subsequent requests on the instance wait until the window expires. Tests include real 1 s `Retry-After` delays.

- **DigitalOcean references removed** — Investigation confirmed no DigitalOcean types or implementation existed in the codebase. Only stale documentation references were present. Removed from CLAUDE.md and FINDINGS.md. No code changes were required.

---

## Phase 3: Release Pipeline Testing

**Commit**: `032d862`

**Scope**: Complete test coverage for the release pipeline: all eight publishers, SDK orchestration, and breaking change detection.

### Completed work

- **Publisher integration tests** (`integration_test.go`, 48 tests):
  - GitHub: dry-run, command-argument building, repository detection, artifact upload.
  - Docker: dry-run, buildx argument construction, config parsing.
  - Homebrew: dry-run, formula generation, Ruby class naming.
  - Scoop: dry-run, manifest JSON generation.
  - AUR: dry-run, `PKGBUILD` and `.SRCINFO` generation.
  - Chocolatey: dry-run, `.nuspec` generation.
  - npm: dry-run, `package.json` generation.
  - LinuxKit: dry-run, multi-format and multi-platform.
  - Cross-publisher: name uniqueness, nil `relCfg` safety, checksum field mapping, interface compliance.

- **SDK generation tests** (`generation_test.go`, 38 tests):
  - SDK orchestration: `Generate` iterates languages, output directory creation, missing-spec error, unknown-language error.
  - Generator registry: register, get, overwrite, language listing.
  - Interface compliance: language identifier correctness, `Available()` safety, install instruction presence.
  - SDK config: defaults, `SetVersion`, nil config safety.
  - Spec detection priority: configured path takes precedence over common paths; all 8 common paths checked.

- **Breaking change detection** (`breaking_test.go`, 30 tests):
  - oasdiff integration: add-endpoint (non-breaking), remove-endpoint (breaking), add-required-parameter (breaking), add-optional-parameter (non-breaking), change-response-type (breaking), remove-HTTP-method (breaking), identical specs.
  - Multiple breaking changes simultaneously.
  - JSON spec format support.
  - Error handling: non-existent base spec, non-existent revision spec, invalid YAML.
  - `DiffExitCode` values: 0 (no diff), 1 (non-breaking), 2 (breaking).
  - `DiffResult` summary and human-readable changes.

---

## Phase 4: DevKit Expansion

**Commit**: `e20083d`

**Scope**: Added three new capabilities to `devkit/`: structured vulnerability scanning, native cyclomatic complexity analysis, and coverage trending with persistence.

### Completed work

- **Vulnerability scanning** (`vulncheck.go` + `vulncheck_test.go`, 13 tests):
  - `VulnCheck(modulePath)` runs `govulncheck -json` and delegates to `ParseVulnCheckJSON`.
  - `ParseVulnCheckJSON` processes newline-delimited JSON, correlating `finding` messages with `osv` metadata. Handles malformed lines, missing OSV entries, empty call traces.
  - `VulnFinding` carries: `ID` (GO-2024-xxxx), `Aliases` (CVE/GHSA), `Package`, `CalledFunction`, `Description`, `FixedVersion`, `ModulePath`.

- **Cyclomatic complexity analysis** (`complexity.go` + `complexity_test.go`, 21 tests):
  - `AnalyseComplexity(cfg)` walks Go source via `go/ast`. No external tools required.
  - `AnalyseComplexitySource(src, filename, threshold)` for in-memory parsing (used in tests).
  - Counts: `if`, `for`, `range`, non-default `case`, `select` comm clause, `&&`, `||`, type switch, `select` statement.
  - Skips `vendor/`, hidden directories, and `_test.go` files.
  - Default threshold: 15.

- **Coverage trending** (`coverage.go` + `coverage_test.go`, 19 tests):
  - `ParseCoverProfile(data)` parses `go test -coverprofile` format, computing per-package statement ratios.
  - `ParseCoverOutput(output)` parses human-readable `go test -cover` output.
  - `CoverageStore` with JSON persistence: `Append`, `Load`, `Latest`.
  - `CompareCoverage(previous, current)` diffs snapshots, returning `CoverageComparison` with `Regressions`, `Improvements`, `NewPackages`, `Removed`, and `TotalDelta`.

---

## Known Limitations

### Embedded Python runtime

`deploy/coolify/` uses an embedded Python 3.13 runtime (`github.com/kluctl/go-embed-python`) to run a Python Swagger client against the Coolify PaaS API. This adds approximately 50 MB to binary size. The design trades binary size for zero native Coolify Go client maintenance. A native Go HTTP client would eliminate this dependency but requires writing and maintaining Coolify API type mappings.

### Single-commit extraction history

All code predating 16 February 2026 arrived in a single commit. `git blame` and `git bisect` cannot identify which changes introduced bugs that existed before extraction. When investigating pre-extraction defects, examine the corresponding history in the `core/go` repository.

### Hypervisor platform specificity

`container/hypervisor.go` selects QEMU (Linux) or Hyperkit (macOS) at runtime. Neither hypervisor is available in standard CI environments. Container package tests use mock hypervisors. Real integration testing requires a machine with the hypervisor binary present.

### Ansible: no role resolution

The Ansible engine supports `include_role` and `import_role` task directives syntactically but does not implement file system role discovery (searching `roles/` directories relative to the playbook). Role tasks must be explicitly inlined or included via `include_tasks`.

### Ansible: no vault decryption

Ansible Vault-encrypted variables and files are not decrypted. Playbooks that rely on vault must decrypt values before passing them to the executor or supply plaintext variables at runtime.

### CLI via Cobra (not core/go CLI framework)

`build/buildcmd/` registers `core build` and `core release` sub-commands using `github.com/spf13/cobra` directly rather than the CLI framework from `forge.lthn.ai/core/go`. This creates a dependency divergence. Alignment with the core/go CLI framework is a future consideration.

### DigitalOcean not implemented

DigitalOcean was documented in early drafts of CLAUDE.md and FINDINGS.md but no types or implementation exist. The documentation references were removed in Phase 2. DigitalOcean support would require a new `infra/digitalocean.go` file using the `APIClient` abstraction.

---

## Future Considerations

- **Native Coolify client** — Replace `deploy/python/` and the embedded Python runtime with a native Go HTTP client for the Coolify v1 API. Eliminates the 50 MB runtime penalty and removes the `kluctl/go-embed-python` dependency.

- **Ansible role resolution** — Implement file system role discovery matching the Ansible convention (`roles/<name>/tasks/main.yml` relative to the playbook directory). Required for running the production DevOps playbooks without pre-inlining role tasks.

- **Ansible vault support** — Add vault decryption using the existing `forge.lthn.ai/core/go-crypt` package (which already manages SSH keys). Allow vault password to be supplied via environment variable or file path.

- **SSH alignment with go-crypt** — `ansible/ssh.go` uses `golang.org/x/crypto/ssh` directly. The `go-crypt` package provides key management. Aligning the two would centralise SSH key handling across the ecosystem.

- **Cobra to core/go CLI alignment** — Migrate `build/buildcmd/` from direct Cobra usage to the core/go CLI framework used by other commands. This is low risk but requires coordination with the parent CLI command tree.

- **DigitalOcean support** — Add `infra/digitalocean.go` implementing the `APIClient`-based pattern established in Phase 2. Required if Lethean infrastructure migrates workloads to DigitalOcean.

- **Coverage trending integration** — Wire `devkit.CoverageStore` into the CI pipeline to accumulate snapshots across runs and fail builds on regression. A `~/.core/coverage.json` or per-project store path would be natural.

- **Build tag isolation for hypervisor tests** — Add `//go:build linux` and `//go:build darwin` tags to `container/` tests that require platform-specific hypervisors, enabling clean CI across both platforms without mock exceptions.
