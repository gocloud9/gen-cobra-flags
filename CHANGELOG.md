# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `+cobra:subcommand:config:json` / `+cobra:subcommand:config:yaml` markers to override the
  `json`/`yaml` struct tags of a parent required field when it is hoisted into a subcommand's
  composed config. Resolution precedence: the hoist override, then `+cobra:json`/`+cobra:yaml`,
  then the cobra flag name (previous default behavior is unchanged when no override is set).
- Extracted the generator into a testable `internal/generator` package behind a thin CLI shim
  (`cmd/gen-cobra-flags`).
- Golden-file and compilation tests for generated output, including same-package mode and
  multiple-struct fixtures.
- End-to-end CLI test that builds the binary and runs it `//go:generate`-style.
- Bug-fix test suite for the runtime SDK adaptors.
- Project documentation: `README.md`, `docs/markers.md`, `CONTRIBUTING.md`, `SECURITY.md`.
- Tooling: `.golangci.yml` (golangci-lint v2), `.editorconfig`, and a `Makefile` covering
  build/test/vet/lint/fmt/generate/tidy/check.
- GoDoc comments on all exported SDK symbols.

### Fixed

- Non-deterministic source package selection during generation.
- Same-package mode template bugs: dropped package qualifier on `To<Struct>`, an unused-variable
  loop for option-less structs, and an unconditional `reflect` import.
- Runtime SDK adaptor bugs in `JsonOrYamlToStruct`, `StringToInteger`, `StringToFloat`,
  `SliceToSlice`, and `IntegerToString`.

[Unreleased]: https://github.com/gocloud9/gen-cobra-flags/commits/main
