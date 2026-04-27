# Development Guide — go-devops

## Prerequisites

| Tool | Minimum version | Purpose |
|------|----------------|---------|
| Go | 1.26 | Build and test |
| Task | any | Taskfile automation (optional, used by some builders) |
| `govulncheck` | latest | Vulnerability scanning (`devkit.VulnCheck`) |
| `gitleaks` | any | Secret scanning (`devkit.ScanSecrets`) |
| `gocyclo` | any | External complexity tool (`devkit.Complexity`) |
| SSH access | — | Integration tests for `ansible/` package |

Install optional tools:

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
go install github.com/zricethezav/gitleaks/v8@latest
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
```

## Local Dependency

`go-devops` depends on `forge.lthn.ai/core/go` (the parent framework). The `go.mod` `replace` directive resolves this locally:

```
replace forge.lthn.ai/core/go => ../core
```

The `../core` path must exist relative to the `go-devops` checkout. If working in a Go workspace (`go.work`), add both modules:

```
go work init
go work use . ../core
```

Do not alter the `replace` directive path.

## Build and Test

```bash
# Run all tests
go test ./...

# Run all tests with race detector
go test -race ./...

# Run a single test by name
go test -v -run TestName ./...

# Run tests in one package
go test ./ansible/...

# Static analysis
go vet ./...

# Check for vulnerabilities
govulncheck ./...

# View test coverage
go test -cover ./...

# Generate a coverage profile
go test -coverprofile=cover.out ./...
go tool cover -html=cover.out
```

## Test Patterns

### Naming Convention

Tests use `_Good`, `_Bad`, and `_Ugly` suffixes:

| Suffix | Meaning |
|--------|---------|
| `_Good` | Happy-path test; expected success |
| `_Bad` | Expected error condition; error must be returned |
| `_Ugly` | Panic, edge case, or degenerate input |

Example:

```go
func TestParsePlaybook_Good(t *testing.T) { ... }
func TestParsePlaybook_Bad(t *testing.T) { ... }
func TestParsePlaybook_Ugly(t *testing.T) { ... }
```

### Assertion Library

Use `github.com/stretchr/testify`. Prefer `require` over `assert` when subsequent assertions depend on the previous one passing:

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestSomething_Good(t *testing.T) {
    result, err := SomeFunction()
    require.NoError(t, err)
    assert.Equal(t, "expected", result.Field)
}
```

### HTTP Test Servers

Use `net/http/httptest` for API client tests. The `infra/` tests demonstrate the pattern:

```go
srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write([]byte(`{"id": 1}`))
}))
defer srv.Close()

client := NewHCloudClient("token", WithHTTPClient(srv.Client()))
```

### SSH Mocking

The `ansible/` package uses an `sshRunner` interface to decouple module implementations from real SSH connections. `mock_ssh_test.go` provides `MockSSHClient` with:

- `expectCommand(pattern, stdout, stderr, rc)` — registers expected command patterns.
- `hasExecuted(pattern)` — asserts a command matching the pattern was called.
- `hasExecutedMethod(method)` — asserts a specific method (`Run`, `RunScript`, `Upload`) was called.
- In-memory filesystem simulation for file operation tests.

Use `MockSSHClient` for all `ansible/modules.go` tests. Real SSH connections are not used in unit tests.

### In-Memory Complexity Analysis

For `devkit` complexity tests, use `AnalyseComplexitySource` rather than writing temporary files:

```go
src := `package foo
func Complex(x int) int {
    if x > 0 { return x }
    return -x
}`
results, err := AnalyseComplexitySource(src, "foo.go", 1)
require.NoError(t, err)
```

### Coverage Store Tests

Use `t.TempDir()` to create temporary directories for `CoverageStore` persistence tests:

```go
dir := t.TempDir()
store := NewCoverageStore(filepath.Join(dir, "coverage.json"))
```

### Publisher Dry-Run Tests

All `release/publishers/` tests use `dryRun: true`. No external services are called. Tests verify:

- Correct command-line argument construction.
- Correct file generation (formula text, manifest JSON, PKGBUILD content).
- Interface compliance: the publisher's `Name()` is non-empty and `Publish` with a nil config does not panic.

---

## Coding Standards

### Language

Use **UK English** in all documentation, comments, identifiers, log messages, and error strings:

- colour (not color)
- organisation (not organization)
- centre (not center)
- behaviour (not behavior)
- licence (noun, not license)

### Strict Types

Every Go file must use strict typing. Avoid `any` at API boundaries where a concrete type is knowable. `map[string]any` is acceptable for Ansible task arguments and YAML-decoded data where the schema is dynamic.

### Error Handling

Use the `core.E` helper from `forge.lthn.ai/core/go` for contextual errors:

```go
return core.E("ansible.Executor.runTask", "failed to upload file", err)
```

For packages that do not import `core/go`, use `fmt.Errorf` with `%w`:

```go
return fmt.Errorf("infra.HCloudClient.ListServers: %w", err)
```

Error strings must not be capitalised and must not end with punctuation (Go convention).

### Import Order

Three groups, each separated by a blank line:

1. Standard library
2. `forge.lthn.ai/core/...` packages
3. Third-party packages

```go
import (
    "context"
    "fmt"

    "forge.lthn.ai/core/go/pkg/io"

    "gopkg.in/yaml.v3"
    "golang.org/x/crypto/ssh"
)
```

### File Headers

Source files do not require a licence header comment beyond the package declaration. The `devkit/` package uses a trailing `// LEK-1 | lthn.ai | EUPL-1.2` comment; maintain this convention in `devkit/` files only.

### Interface Placement

Define interfaces in the package that consumes them, not the package that implements them. The `Builder`, `Publisher`, `Signer`, `Generator`, `Hypervisor`, and `ImageSource` interfaces each live in the package that calls them.

---

## Conventional Commits

All commits follow the Conventional Commits specification.

**Format**: `type(scope): description`

**Scopes** map to package names:

| Scope | Package |
|-------|---------|
| `ansible` | `ansible/` |
| `build` | `build/`, `build/builders/`, `build/signing/` |
| `container` | `container/` |
| `devkit` | `devkit/` |
| `devops` | `devops/` |
| `infra` | `infra/` |
| `release` | `release/`, `release/publishers/` |
| `sdk` | `sdk/`, `sdk/generators/` |
| `deploy` | `deploy/` |

**Examples**:

```
feat(ansible): add docker_compose module support
fix(infra): handle nil Retry-After header in rate limiter
refactor(build): extract archive creation into separate function
test(devkit): expand coverage trending snapshot comparison tests
chore: update go.sum after dependency upgrade
```

**Co-author line**: every commit must include:

```
Co-Authored-By: Virgil <virgil@lethean.io>
```

---

## Licence

All source files are licensed under the **European Union Public Licence 1.2 (EUPL-1.2)**. Do not introduce dependencies with licences incompatible with EUPL-1.2. The `github.com/kluctl/go-embed-python` dependency (Apache 2.0) and `golang.org/x/crypto` (BSD-3-Clause) are compatible. Verify new dependencies before adding them.

---

## Forge Repository

- **Remote**: `ssh://git@forge.lthn.ai:2223/core/go-devops.git`
- **Push**: `git push forge main`
- HTTPS authentication is not supported on the Forge instance; SSH is required.

---

## Adding a New Module to ansible/

1. Add the module name(s) to `KnownModules` in `types.go`.
2. Implement a function `executeModuleName(ctx, ssh, args, vars) TaskResult` in `modules.go`.
3. Add a `case "modulename":` branch in the dispatch switch in `executor.go`.
4. Add a shim to `mock_ssh_test.go`'s `sshRunner` interface (if the module requires file operations, use `sshFileRunner`).
5. Write tests in `modules_*_test.go` using the mock infrastructure. Cover at minimum: success case, changed vs. unchanged, argument validation failure, and SSH error propagation.

## Adding a New Release Publisher

1. Create `release/publishers/myplatform.go`.
2. Implement `Publisher`:
   - `Name() string` — return the platform name.
   - `Publish(ctx, release, pubCfg, relCfg, dryRun) error` — when `dryRun` is true, log intent and return nil.
3. Register the publisher in `release/config.go` alongside existing publishers.
4. Write `release/publishers/myplatform_test.go` with dry-run tests. Follow the pattern of existing publisher tests: verify command arguments, generated file content, and interface compliance.

## Adding a New Builder

1. Create `build/builders/mylang.go`.
2. Implement `Builder`:
   - `Name() string`
   - `Detect(fs io.Medium, dir string) (bool, error)` — check for a marker file.
   - `Build(ctx, cfg, targets) ([]Artifact, error)`
3. Register the builder in `build/buildcmd/`.
4. Write tests verifying `Detect` (marker present/absent) and `Build` (at minimum with a mock `io.Medium`).
