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

- [ ] **Module test coverage** — `modules.go` is 1,434 LOC with zero tests. Each module (service, file, template, command, copy, apt, yum) needs unit tests with mocked SSH sessions.
- [ ] **Error propagation** — Verify all SSH errors are wrapped with `core.E()` including host context. Currently some errors may lose the host identifier.
- [ ] **Fact gathering** — Test fact collection from different Linux distros (Ubuntu, CentOS, Alpine). Mock `/etc/os-release` parsing.
- [ ] **Become/sudo** — Test privilege escalation paths. Verify password prompt handling.
- [ ] **Idempotency checks** — Modules should report `changed: false` when no action needed. Verify for file, service, template modules.

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
