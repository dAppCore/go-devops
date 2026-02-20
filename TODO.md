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

- [ ] **Test copy/template/file/lineinfile/blockinfile/stat** — Verify:
  - `copy`: calls `client.Upload()` with content, applies chown/chgrp
  - `file`: handles state branches (directory/absent/touch/link) with correct mkdir/chmod/chown/ln commands
  - `lineinfile`: builds correct sed commands for line manipulation
  - `blockinfile`: marker-based block management with heredoc escaping
  - `stat`: calls `client.Stat()`, returns file info map
  - `template`: uses `e.TemplateFile()` then `client.Upload()`

### Step 1.3: Service & package modules (7 modules, ~180 LOC)

- [ ] **Test service/systemd/apt/apt_key/apt_repository/package/pip** — Verify:
  - `service`: correct systemctl start/stop/restart/enable/disable commands
  - `systemd`: daemon_reload + delegation to service
  - `apt`: correct apt-get install/remove/update commands
  - `package`: auto-detection of apt vs yum

### Step 1.4: User/group & advanced modules (10 modules, ~385 LOC)

- [ ] **Test user/group/cron/authorized_key/git/unarchive/uri/ufw/docker_compose** — Verify:
  - `user`: conditional useradd vs usermod based on `id` check
  - `cron`: crontab list/edit/delete with comment markers
  - `authorized_key`: SSH key management, grep-based idempotency
  - `git`: clone vs fetch+checkout logic based on FileExists
  - `unarchive`: Upload + tar/zip extraction

### Step 1.5: Error propagation & become

- [ ] **Error propagation** — Verify all SSH errors are wrapped with `core.E()` including host context. Test SSH failures in Run/Upload/Download paths.
- [ ] **Become/sudo** — Test privilege escalation: `SetBecome(true, "root", "password")` → verify `sudo -S` prefix on commands. Test passwordless sudo (`-n` flag).
- [ ] **Fact gathering** — Test fact collection mocking `/etc/os-release` for Ubuntu, CentOS, Alpine. Verify distro detection.
- [ ] **Idempotency checks** — Verify `changed: false` when no action needed for file, service, user, apt modules.

## Phase 2: Infrastructure API Robustness

- [ ] **Retry logic** — Add configurable retry with exponential backoff for Hetzner Cloud/Robot and CloudNS API calls. Cloud APIs are flaky.
- [ ] **Rate limiting** — Hetzner Cloud has rate limits. Detect 429 responses, queue and retry.
- [ ] **DigitalOcean support** — Currently referenced in config but no implementation. Either implement or remove.
- [ ] **API client abstraction** — Extract common HTTP client pattern from hetzner.go and cloudns.go into shared infra client.

## Phase 3: Release Pipeline Testing

- [ ] **Publisher integration tests** — Mock GitHub API for release creation, Docker registry for image push, Homebrew tap for formula update. Verify dry-run mode produces correct output without side effects.
- [ ] **SDK generation tests** — Generate TypeScript/Go/Python clients from a test OpenAPI spec. Verify output compiles/type-checks.
- [ ] **Breaking change detection** — Test oasdiff integration: modify a spec with breaking change, verify detection and failure mode.

## Phase 4: DevKit Expansion

- [ ] **Vulnerability scanning** — Integrate `govulncheck` output parsing into devkit findings.
- [ ] **Complexity thresholds** — Configurable cyclomatic complexity threshold. Flag functions exceeding it.
- [ ] **Coverage trending** — Store coverage snapshots, detect regressions between runs.

---

## Workflow

1. Virgil in core/go writes tasks here after research
2. This repo's dedicated session picks up tasks in phase order
3. Mark `[x]` when done, note commit hash
