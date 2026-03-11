---
title: Doc Sync
description: Documentation sync command for collecting docs from multi-repo workspaces into a central location.
---

# Doc Sync

The `core docs` commands scan a multi-repo workspace for documentation and sync it to a central location. This enables unified documentation builds across federated monorepos.

## Commands

### core docs list

Scan all repos in the workspace and display a table showing documentation coverage:

```bash
core docs list                         # Uses auto-detected repos.yaml
core docs list --registry path/to/repos.yaml
```

Output shows which repos have:
- `README.md`
- `CLAUDE.md`
- `CHANGELOG.md`
- `docs/` directory (with file count)

### core docs sync

Copy `docs/` directories from all repos into a central output location:

```bash
core docs sync                         # Sync to default target (php)
core docs sync --dry-run               # Preview without writing
core docs sync --target gohelp         # Sync to go-help format
core docs sync --target zensical       # Sync to Zensical/Hugo format
core docs sync --output /path/to/dest  # Custom output directory
core docs sync --registry repos.yaml   # Explicit registry file
```

## Sync targets

The `--target` flag controls the output format and default destination:

### php (default)

Copies documentation to `core-php/docs/packages/{name}/`. Each repo's `docs/` directory is copied as-is, with directory names mapped:

- `core` maps to `packages/go/`
- `core-admin` maps to `packages/admin/`
- `core-api` maps to `packages/api/`

Skips `core-php` (the destination) and `core-template`.

### gohelp

Plain copy to a go-help content directory (`docs/content/` by default). No frontmatter injection. Directory names follow the same mapping as the php target.

### zensical

Copies to a Zensical/Hugo docs directory (`docs-site/docs/` by default) with automatic Hugo frontmatter injection. Repos are mapped to content sections:

| Repo pattern | Section | Folder |
|-------------|---------|--------|
| `cli` | `getting-started` | — |
| `core` | `cli` | — |
| `go-*` | `go` | repo name |
| `core-*` | `php` | name without prefix |

Frontmatter (title and weight) is added to markdown files that lack it. READMEs become `index.md` files. `KB/` directories are synced to a `kb/` section.

## How it works

1. **Load registry** — finds `repos.yaml` (auto-detected or explicit path) and loads the repo list. Respects `workspace.yaml` for custom `packages_dir`.
2. **Scan repos** — walks each repo looking for `README.md`, `CLAUDE.md`, `CHANGELOG.md`, `docs/*.md` (recursive), and `KB/*.md` (recursive). The `docs/plans/` subdirectory is skipped.
3. **Display plan** — shows which repos have documentation and where files will be written.
4. **Confirm** — prompts for confirmation (skipped in `--dry-run` mode).
5. **Copy** — clears existing output directories and copies files. For the zensical target, injects Hugo frontmatter where missing.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--registry` | auto-detect | Path to `repos.yaml` |
| `--dry-run` | `false` | Preview sync plan without writing files |
| `--output` | target-dependent | Output directory (overrides target default) |
| `--target` | `php` | Output format: `php`, `gohelp`, or `zensical` |

## Registry discovery

The registry (list of repos) is found by:

1. Explicit `--registry` path if provided.
2. Auto-detection: `repos.yaml` in the current directory or parent directories.
3. Fallback: scan the current directory for git repositories.

If a `workspace.yaml` file exists alongside the registry, its `packages_dir` setting overrides the default repo path resolution.

## RepoDocInfo

Each scanned repo produces a `RepoDocInfo` struct:

| Field | Type | Description |
|-------|------|-------------|
| `Name` | `string` | Repository name |
| `Path` | `string` | Absolute filesystem path |
| `HasDocs` | `bool` | Whether any documentation was found |
| `Readme` | `string` | Path to `README.md` (empty if missing) |
| `ClaudeMd` | `string` | Path to `CLAUDE.md` (empty if missing) |
| `Changelog` | `string` | Path to `CHANGELOG.md` (empty if missing) |
| `DocsFiles` | `[]string` | Relative paths of files in `docs/` |
| `KBFiles` | `[]string` | Relative paths of files in `KB/` |
