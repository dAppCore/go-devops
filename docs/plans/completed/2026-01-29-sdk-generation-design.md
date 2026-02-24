# SDK Generation Design -- Completion Summary

**Status:** COMPLETED
**Date Completed:** 2026-01-29 (initial implementation), verified 2026-02-24
**Plan:** Generate typed API clients from OpenAPI specs for TypeScript, Python, Go, and PHP with breaking change detection

## What Was Built

Full `sdk/` package with generator interface, four language generators, OpenAPI spec detection, and breaking change detection via oasdiff.

### Key Files

- `sdk/sdk.go` -- SDK struct, Config types, Generate/GenerateLanguage/SetVersion
- `sdk/detect.go` -- OpenAPI spec detection (config, common paths, Scramble)
- `sdk/diff.go` -- Breaking change detection using oasdiff (DiffResult, DiffExitCode)
- `sdk/generators/generator.go` -- Generator interface and Registry
- `sdk/generators/typescript.go` -- openapi-typescript-codegen (native/npx/Docker)
- `sdk/generators/python.go` -- openapi-python-client (native/Docker)
- `sdk/generators/go.go` -- oapi-codegen (native/Docker)
- `sdk/generators/php.go` -- openapi-generator via Docker
- `cmd/sdk/cmd.go` -- CLI: `core sdk diff`, `core sdk validate`
- `build/buildcmd/cmd_sdk.go` -- CLI: `core build sdk`

### Test Coverage

- `sdk/sdk_test.go`, `sdk/detect_test.go`, `sdk/diff_test.go`
- `sdk/breaking_test.go`, `sdk/generation_test.go`
- `sdk/generators/{typescript,python,go,php}_test.go`
