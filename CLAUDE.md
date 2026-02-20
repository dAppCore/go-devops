# CLAUDE.md — go-devops Agent Instructions

You are a dedicated domain expert for `forge.lthn.ai/core/go-devops`. Virgil (in core/go) orchestrates your work via TODO.md. Pick up tasks in phase order, mark `[x]` when done, commit and push.

## Commands

```bash
go test ./...                    # Run all tests
go test -v -run TestName ./...   # Single test
go test -race ./...              # Race detector
go vet ./...                     # Static analysis
```

## Local Dependencies

| Module | Local Path | Notes |
|--------|-----------|-------|
| `forge.lthn.ai/core/go` | `../core` | Framework (core.E, io.Medium, config, i18n, log) |

**Do NOT change the replace directive path.**

## Key Interfaces

```go
// build/builders/
type Builder interface {
    Name() string
    Detect(fs io.Medium, dir string) (bool, error)
    Build(ctx context.Context, cfg *Config, targets []Target) ([]Artifact, error)
}

// release/publishers/
type Publisher interface {
    Name() string
    Publish(ctx context.Context, release *Release, pubCfg PublisherConfig, relCfg ReleaseConfig, dryRun bool) error
}

// container/
type Hypervisor interface {
    Name() string
    Available() bool
    Run(ctx context.Context, opts RunOptions) (*process.Handle, error)
}

// devops/sources/
type ImageSource interface {
    Name() string
    Available() bool
    Download(ctx context.Context, name, version string, progress func(downloaded, total int64)) (string, error)
}

// build/signing/
type Signer interface {
    Name() string
    Available() bool
    Sign(filePath, keyID string) ([]byte, error)
}

// sdk/generators/
type Generator interface {
    Language() string
    Generate(ctx context.Context, spec, outputDir string, config *Config) error
}
```

## Coding Standards

- **UK English**: colour, organisation, centre
- **Tests**: testify assert/require, `_Good`/`_Bad`/`_Ugly` naming convention
- **Conventional commits**: `feat(ansible):`, `fix(infra):`, `refactor(build):`
- **Co-Author**: `Co-Authored-By: Virgil <virgil@lethean.io>`
- **Licence**: EUPL-1.2
- **Imports**: stdlib → forge.lthn.ai → third-party, each group separated by blank line

## Forge

- **Repo**: `forge.lthn.ai/core/go-devops`
- **Push via SSH**: `git push forge main` (remote: `ssh://git@forge.lthn.ai:2223/core/go-devops.git`)

## Documentation

- Architecture: `docs/architecture.md`
- Development guide: `docs/development.md`
- Project history: `docs/history.md`

## Task Queue

See `TODO.md` for prioritised work.
