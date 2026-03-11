---
title: SDK Generation
description: OpenAPI client generation for TypeScript, Python, Go, and PHP with breaking change detection.
---

# SDK Generation

The `sdk/` package generates typed API client libraries from OpenAPI specifications. It supports four languages, auto-detects spec files, and uses oasdiff for semantic breaking change detection. Configuration lives in the `sdk:` section of `.core/release.yaml`.

## How it works

```
core build sdk                    # Generate SDKs using .core/release.yaml config
core build sdk --spec api.yaml    # Explicit spec file
core build sdk --lang typescript  # Single language only
```

The generation pipeline:

1. **Detect spec** — find the OpenAPI file (config, common paths, or Laravel Scramble).
2. **Run diff** — if enabled, compare current spec against the previous release and flag breaking changes.
3. **Generate** — run the appropriate generator for each configured language.
4. **Output** — write client libraries to `sdk/{language}/`.

## OpenAPI spec detection

`detect.go` probes locations in priority order:

1. Path configured in `.core/release.yaml` (`sdk.spec` field).
2. Common file paths:
   - `api/openapi.yaml`, `api/openapi.json`
   - `openapi.yaml`, `openapi.json`
   - `docs/api.yaml`, `docs/openapi.yaml`
   - `swagger.yaml`
3. Laravel Scramble — if `scramble/scramble` is in `composer.json`, runs `php artisan scramble:export` to generate the spec.
4. Error if no spec found.

## Generator interface

Each language generator implements:

```go
type Generator interface {
    Language() string
    Generate(ctx context.Context, spec, outputDir string, config *Config) error
}
```

Generators are registered in a string-keyed registry, allowing overrides.

## Language generators

Each generator uses a native tool when available and falls back to Docker-based generation:

| Language | Native tool | Fallback | Install |
|----------|------------|----------|---------|
| TypeScript | `openapi-typescript-codegen` | openapi-generator (Docker) | `npm i -g openapi-typescript-codegen` |
| Python | `openapi-python-client` | openapi-generator (Docker) | `pip install openapi-python-client` |
| Go | `oapi-codegen` | openapi-generator (Docker) | `go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest` |
| PHP | openapi-generator (Docker) | — | Requires Docker |

### Fallback strategy

When the native tool is not installed, generators automatically fall back to running openapi-generator inside Docker:

```go
func (g *TypeScriptGenerator) Generate(ctx context.Context, opts GenerateOptions) error {
    if g.Available() {
        return g.generateNative(ctx, opts)
    }
    return g.generateDocker(ctx, opts)
}
```

## Output structure

Each generator writes to `sdk/{language}/`:

```
sdk/
├── typescript/
│   ├── package.json
│   ├── src/
│   │   ├── index.ts
│   │   ├── client.ts
│   │   └── models/
│   └── tsconfig.json
├── python/
│   ├── setup.py
│   ├── myapi/
│   │   ├── __init__.py
│   │   ├── client.py
│   │   └── models/
│   └── requirements.txt
├── go/
│   ├── go.mod
│   ├── client.go
│   └── models.go
└── php/
    ├── composer.json
    ├── src/
    │   ├── Client.php
    │   └── Models/
    └── README.md
```

## Breaking change detection

`diff.go` uses [oasdiff](https://github.com/Tufin/oasdiff) to compare two OpenAPI spec versions semantically. It detects:

- Removed endpoints
- Changed required parameters
- Modified response schemas
- Changed authentication requirements

### Usage

```bash
core sdk diff                        # Compare current spec vs last release
core sdk diff --spec api.yaml --base v1.0.0
core sdk validate                    # Validate the spec without generating
```

### CI exit codes

| Exit code | Meaning |
|-----------|---------|
| 0 | No breaking changes |
| 1 | Breaking changes detected |
| 2 | Error (invalid spec, missing file) |

### Diff configuration

The `sdk.diff` section in `release.yaml` controls behaviour:

```yaml
sdk:
  diff:
    enabled: true           # Run diff check before generation
    fail_on_breaking: true  # Abort on breaking changes (CI-friendly)
```

When `fail_on_breaking` is `false`, breaking changes produce a warning but generation continues.

## Release integration

SDK generation integrates with the release pipeline. When the `sdk:` section is present in `release.yaml`, `core build release` runs SDK generation after building artefacts:

```
core build release
  -> build artefacts
  -> generate SDKs (if sdk: configured)
  -> run diff check (warns or fails on breaking)
  -> publish to GitHub release
  -> publish SDKs (optional)
```

SDK generation can also run independently:

```bash
core build sdk                           # Generate using release.yaml config
core build sdk --version v1.2.3          # Explicit version
core build sdk --dry-run                 # Preview without generating
```

## Configuration reference

In `.core/release.yaml`:

```yaml
sdk:
  spec: api/openapi.yaml       # OpenAPI spec path (auto-detected if omitted)

  languages:                    # Languages to generate
    - typescript
    - python
    - go
    - php

  output: sdk                   # Output directory (default: sdk/)

  package:
    name: myapi                 # Base package name
    version: "{{.Version}}"     # Template — uses release version

  diff:
    enabled: true               # Run breaking change detection
    fail_on_breaking: true      # Fail on breaking changes (for CI)

  publish:                      # Optional monorepo publishing
    repo: myorg/sdks
    path: packages/myapi
```

### Field reference

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `sdk.spec` | `string` | auto-detect | Path to OpenAPI spec |
| `sdk.languages` | `[]string` | — | Languages to generate (`typescript`, `python`, `go`, `php`) |
| `sdk.output` | `string` | `sdk/` | Output directory |
| `sdk.package.name` | `string` | — | Base package name for generated clients |
| `sdk.package.version` | `string` | — | Version template (supports `{{.Version}}`) |
| `sdk.diff.enabled` | `bool` | `false` | Run breaking change detection |
| `sdk.diff.fail_on_breaking` | `bool` | `false` | Abort on breaking changes |
| `sdk.publish.repo` | `string` | — | Monorepo target (`owner/repo`) |
| `sdk.publish.path` | `string` | — | Path within the monorepo |

## Dependencies

SDK generation relies on:

| Dependency | Purpose |
|-----------|---------|
| `github.com/getkin/kin-openapi` | OpenAPI 3.x spec parsing and validation |
| `github.com/oasdiff/oasdiff` | Semantic API diff and breaking change detection |
