# TODO.md — go-devops

Dispatched from core/go orchestration. Pick up tasks in order.

---

## Phase 0: Test Coverage & Hardening

- [x] **Expand ansible/ tests** — Added `parser_test.go` (17 tests: ParsePlaybook, ParseInventory, ParseTasks, GetHosts, GetHostVars, isModule, NormalizeModule), `types_test.go` (RoleRef/Task UnmarshalYAML, Inventory, Facts, TaskResult, KnownModules), `executor_test.go` (getHosts, matchesTags, evaluateWhen, templateString, applyFilter, resolveLoop, templateArgs, handleNotify, normalizeConditions, helper functions). All pass. Commit `6e346cb`.
- [x] **Expand infra/ tests** — Added `hetzner_test.go` (HCloudClient/HRobotClient construction, do() round-trip via httptest, API error handling, JSON serialisation for HCloudServer, HCloudLoadBalancer, HRobotServer) and `cloudns_test.go` (doRaw() round-trip, zone/record JSON, CRUD responses, ACME challenge, auth params, errors). Commit `6e346cb`.
- [x] **Expand build/ tests** — Added `archive_test.go` (archive round-trip for tar.gz/zip, multi-file archives, 249 LOC) and extended `signing_test.go` (mock signer tests, path verification, error handling, +181 LOC). Commit `5d22ed9`.
- [x] **Expand release/ tests** — Fixed nil pointer crash in `linuxkit.go:50` (added `release.FS == nil` guard). Added nil FS test case to `linuxkit_test.go` (+23 LOC). 862 tests pass across build/ and release/. Commit `5d22ed9`.
- [x] **Race condition tests** — `go test -race ./...` clean across ansible, infra, container, devops, build packages. Commit `6e346cb`.
- [x] **`go vet ./...` clean** — Fixed stale API calls in container/linuxkit_test.go, state_test.go, templates_test.go, devops/devops_test.go. go.mod replace directive fixed. Commit `6e346cb`.

## Phase 1: Ansible Engine Hardening

### Step 1.0: SSH mock infrastructure

- [x] **Create `ansible/mock_ssh_test.go`** — MockSSHClient with command registry (`expectCommand`), file system simulation (in-memory map), become state tracking, execution/upload logs, and assertion helpers (`hasExecuted`, `hasExecutedMethod`, `findExecuted`). Module shims via `sshRunner` interface for testability. 12 mock infrastructure tests. Commit `3330e55`.

### Step 1.1: Command execution modules (4 modules, ~100 LOC)

- [x] **Test command/shell/raw/script** — 36 module tests + 12 mock tests = 48 new tests. Verifies: command uses `Run()`, shell uses `RunScript()`, raw passes through without wrapping, script reads local file. Cross-module differentiation tests, dispatch routing, template variable resolution. Commit `3330e55`.

### Step 1.2: File operation modules (6 modules, ~280 LOC)

- [x] **Test copy/template/file/lineinfile/blockinfile/stat** — 54 new tests: copy (8), file (12), lineinfile (8), blockinfile (7), stat (5), template (6), dispatch (6), integration (2). Extended mock with `sshFileRunner` interface and 6 module shims. Fixed unsupported module test (copy→hostname). Total ansible tests: 208. Commit `c7da9ad`.

### Step 1.3: Service & package modules (7 modules, ~180 LOC)

- [x] **Test service/systemd/apt/apt_key/apt_repository/package/pip** — 56 new tests: service (12), systemd (4), apt (9), apt_key (6), apt_repository (8), package (3), pip (8), dispatch (7). 7 module shims in mock_ssh_test.go. Commit `9638e77`.

### Step 1.4: User/group & advanced modules (10 modules, ~385 LOC)

- [x] **Test user/group/cron/authorized_key/git/unarchive/uri/ufw/docker_compose** — 69 new tests: user (7), group (7), cron (5), authorized_key (7), git (8), unarchive (8), uri (6), ufw (8), docker_compose (7), dispatch (6). 9 module shims. Total ansible tests: 334. Commit `427929f`.

### Step 1.5: Error propagation & become

- [x] **Error propagation** — 68 tests: getHosts (8), matchesTags (7), evaluateWhen/evalCondition (22), templateString (14), applyFilter (9), resolveLoop (5), handleNotify (5), normalizeConditions (6), cross-cutting (7).
- [x] **Become/sudo** — 8 tests: enable/disable cycle, default user, passwordless, play-level become.
- [x] **Fact gathering** — 9 tests: Ubuntu/CentOS/Alpine/Debian os-release parsing, hostname, localhost.
- [x] **Idempotency checks** — 8 tests: group exists, authorized_key present, docker compose up-to-date, stat always unchanged.
- Total ansible tests: 438. Phase 1 complete. Commit `8ab8643`.

## Phase 2: Infrastructure API Robustness

- [x] **API client abstraction** — Extracted shared `APIClient` struct in `infra/client.go` with `Do()` (JSON) and `DoRaw()` (bytes) methods. `HCloudClient`, `HRobotClient`, and `CloudNSClient` now delegate to `APIClient` via configurable auth functions and error prefixes. Options: `WithHTTPClient`, `WithRetry`, `WithAuth`, `WithPrefix`. 30 new `client_test.go` tests.
- [x] **Retry logic** — `APIClient` implements configurable exponential backoff with jitter. Retries on 5xx and transport errors; does NOT retry 4xx (except 429). `RetryConfig{MaxRetries, InitialBackoff, MaxBackoff}` with `DefaultRetryConfig()` (3 retries, 100ms, 5s). Context cancellation respected during backoff.
- [x] **Rate limiting** — Detects 429 responses, parses `Retry-After` header (seconds format, falls back to 1s). Queues subsequent requests behind a shared `blockedUntil` mutex. Rate limit wait is per-`APIClient` instance. Tested with real 1s Retry-After delays.
- [x] **DigitalOcean support** — Investigated: no types or implementation existed in `infra/` code, only stale documentation references in CLAUDE.md and FINDINGS.md. Removed the dead references. No code changes needed.

## Phase 3: Release Pipeline Testing

- [x] **Publisher integration tests** — Added `integration_test.go` (48 tests): GitHub dry-run/command-building/repo-detection/artifact-upload, Docker dry-run/buildx-args/config-parsing, Homebrew dry-run/formula-generation/class-naming, Scoop dry-run/manifest, AUR dry-run/PKGBUILD/SRCINFO, Chocolatey dry-run/nuspec, npm dry-run/package.json, LinuxKit dry-run multi-format/platform. Cross-publisher: name uniqueness, nil relCfg, checksum mapping, interface compliance. Commit `032d862`.
- [x] **SDK generation tests** — Added `generation_test.go` (38 tests): SDK orchestration (Generate iterates languages, output directory, no-spec error, unknown language), generator registry (register/get/overwrite/languages), generator interface compliance (language identifiers, install instructions, Available safety), SDK config (defaults, SetVersion, nil config), spec detection priority (configured > common > Scramble, all 8 common paths). Commit `032d862`.
- [x] **Breaking change detection** — Added `breaking_test.go` (30 tests): oasdiff integration for add-endpoint (non-breaking), remove-endpoint (breaking), add-required-param (breaking), add-optional-param (non-breaking), change-response-type (breaking), remove-HTTP-method (breaking), identical-specs, multiple-breaking-changes, JSON spec support. Error handling: non-existent base/revision, invalid YAML. DiffExitCode (0/1/2), DiffResult summary/human-readable changes. Commit `032d862`.

## Phase 4: DevKit Expansion

- [x] **Vulnerability scanning** — `VulnCheck()` runs `govulncheck -json` and `ParseVulnCheckJSON()` parses newline-delimited JSON into `VulnFinding` structs (ID, aliases, package, called function, description, fixed version, module path). Handles malformed lines, missing OSV entries, empty traces. 13 tests in `vulncheck_test.go`. Commit `e20083d`.
- [x] **Complexity thresholds** — `AnalyseComplexity()` walks Go source via `go/ast` with configurable threshold (default 15). Counts: if, for, range, case, comm, &&, ||, type switch, select. Skips vendor/, test files, hidden dirs. `AnalyseComplexitySource()` for in-memory parsing. 21 tests in `complexity_test.go`. Commit `e20083d`.
- [x] **Coverage trending** — `ParseCoverProfile()` parses coverprofile format, `ParseCoverOutput()` parses human-readable `go test -cover` output. `CoverageStore` with JSON persistence (Append/Load/Latest). `CompareCoverage()` diffs snapshots, flags regressions/improvements/new/removed packages. 19 tests in `coverage_test.go`. Commit `e20083d`.

---

## Workflow

1. Virgil in core/go writes tasks here after research
2. This repo's dedicated session picks up tasks in phase order
3. Mark `[x]` when done, note commit hash
