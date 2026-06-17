# Contributing to gen-cobra-flags

Thanks for your interest in contributing! This document describes how to get set up and the
conventions we follow.

## Prerequisites

- Go 1.26.0 or newer.
- `make` (the `Makefile` wraps the common workflows).

The lint target installs a pinned version of `golangci-lint` (v2) into `$(go env GOPATH)/bin`
automatically; you do not need to install it separately.

## Repository layout

This repository is a Go workspace with three modules:

- `/` — the generator CLI (`cmd/gen-cobra-flags`) and its internals (`internal/generator`).
- `/sdk` — the runtime library imported by generated code.
- `/example` — a worked example / fixture used by tests. **Do not hand-edit `example/config.go`**;
  it is a read-only fixture. You may regenerate `example/generated/*` via `make generate`.

## Common tasks

```sh
make build      # build the CLI into ./bin
make test       # run tests across all modules
make vet        # go vet across all modules
make lint       # golangci-lint across all modules
make fmt        # gofmt + import grouping across all modules
make generate   # rebuild the CLI and regenerate the example output
make tidy       # go mod tidy across all modules
make check      # vet + lint + test across all modules
```

Before opening a pull request, run:

```sh
make check
```

and ensure it is green.

## Coding conventions

- Code must pass `make lint`. All exported symbols require GoDoc comments.
- Run `make fmt` to format code and group imports (local imports are grouped under
  `github.com/gocloud9/gen-cobra-flags`).
- New generator behavior should be covered by golden tests under
  `internal/generator/testdata/`. Update golden files with the test's `-update` flag and review
  the diff carefully.
- The generator must produce deterministic output. If you touch package or struct iteration,
  verify output stability by running the relevant tests repeatedly.

## Commit messages

Use [Conventional Commits](https://www.conventionalcommits.org/) style, for example:

```
feat: add support for custom flag adaptors
fix: stabilise source package selection
docs: document +cobra markers
```

## Pull requests

- Keep changes focused and include tests where applicable.
- Describe the motivation ("why") in the PR description, not just the change ("what").
- Make sure `make check` passes.

## Reporting security issues

Please do not file public issues for security vulnerabilities. See [SECURITY.md](SECURITY.md).
