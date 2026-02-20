# TODO.md — go-devops

Dispatched from core/go orchestration. Pick up tasks in order.

---

## Phase 0: Test Coverage & Hardening

- [ ] **Expand ansible/ tests** — Only `ssh_test.go` exists. Add: `executor_test.go` (run a minimal playbook with mock SSH, verify task order + handler notification), `modules_test.go` (test each module: service start/stop, file copy, template render, command exec — use mocked SSH session), `parser_test.go` (parse valid playbook YAML, invalid YAML, empty plays, nested vars), `types_test.go` (Facts merge, Inventory host grouping).
- [ ] **Expand infra/ tests** — Only `config_test.go` exists. Add: `hetzner_test.go` (mock HTTP responses for server list/create/delete, load balancer ops, snapshot management), `cloudns_test.go` (mock DNS zone/record CRUD, ACME challenge create/cleanup, error responses). Use `httptest.NewServer` for API mocking.
- [ ] **Expand build/ tests** — Add: builder detection tests (each builder's `Detect()` with matching/non-matching directory structures), archive round-trip (create tar.gz → extract → compare), signing mock tests (verify `Sign()` called with correct paths).
- [ ] **Expand release/ tests** — Add: version detection from git tags / package.json / go.mod, changelog generation from conventional commits (mock git log output), publisher dry-run tests.
- [ ] **Race condition tests** — `go test -race ./...` across all packages. Ansible executor runs concurrent handlers — verify thread safety.
- [ ] **`go vet ./...` clean** — Fix any warnings.

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
