# go-devops Agent Guide

This repository contains the DevOps command surface for the Core workspace. The
main packages live under `cmd/` and expose CLI groups for deployment, workspace
setup, documentation synchronization, Git helpers, and developer workflow
automation. The supporting packages under `deploy/`, `devkit/`, and `snapshot/`
hold integration code that those commands call.

Work in this repository should follow the Core v0.9 conventions. Import
`dappco.re/go` as `core` or dot-import it in tests when a Core wrapper exists,
including filesystem, path, JSON, formatting, environment, and string helpers.
Do not add direct standard library imports for those wrapper-covered areas in
new Go files or tests. Command execution should use the Core process wrapper so
call sites stay consistent with the rest of the tree.

Tests are organized beside the source file they exercise. Public functions and
methods require the Good, Bad, and Ugly test triplet in the sibling test file,
and examples belong in the sibling example test file with checkable output.
When a command touches external services such as GitHub, Forgejo, Coolify, or
embedded Python, tests should use local fakes or hooks unless the code is
explicitly testing environment-dependent behavior.

Before handing off a branch, run the v0.9 audit script from the repo root and
keep iterating until it reports compliance. The audit is intentionally strict
about file placement, examples, banned compatibility shims, and generated test
dump files because this project is used as a consumer canary for Core upgrade
patterns.
