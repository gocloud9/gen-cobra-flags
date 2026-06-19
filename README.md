# gen-cobra-flags

`gen-cobra-flags` is a Go code generator that produces [Cobra](https://github.com/spf13/cobra)
CLI flag-binding boilerplate from plain Go structs annotated with `+cobra:*` marker comments.

Annotate a struct once, run `go generate`, and get:

- `Add<Struct>Flags(cmd)` — registers all flags on a `*cobra.Command`.
- `<Struct>ConfigFromFlags(cmd)` — reads the flags back into a config struct.
- `(*<Struct>Config).To<Struct>(...)` — converts the config into your domain type.

This removes the repetitive, error-prone work of wiring `cmd.Flags().StringP(...)`,
`cmd.Flags().GetString(...)`, and type conversions by hand.

## Installation

```sh
go install github.com/gocloud9/gen-cobra-flags/cmd/gen-cobra-flags@latest
```

This installs the `gen-cobra-flags` binary into `$(go env GOPATH)/bin`. Make sure that
directory is on your `PATH`.

## Quick start

1. Annotate a struct with `+cobra:*` markers and add a `//go:generate` directive:

   ```go
   package simple

   //go:generate gen-cobra-flags -input ./ -struct SimpleRequest -output ./ -package simple

   // SimpleRequest is a minimal annotated struct.
   // +cobra:flag=simple
   // +cobra:short=s
   // +cobra:usage=A simple request
   type SimpleRequest struct {
       // +cobra:flag=title
       // +cobra:short=t
       // +cobra:usage=The title
       // +cobra:default=""
       Title string

       // +cobra:flag=count
       // +cobra:usage=The count
       // +cobra:default=0
       Count int

       // +cobra:flag=enabled
       // +cobra:usage=Whether enabled
       // +cobra:default=false
       Enabled bool
   }
   ```

2. Run the generator:

   ```sh
   go generate ./...
   ```

3. Wire the generated helpers into your command:

   ```go
   cmd := &cobra.Command{
       Use: "create",
       RunE: func(cmd *cobra.Command, args []string) error {
           cfg, err := simple.SimpleRequestConfigFromFlags(cmd)
           if err != nil {
               return err
           }
           req, err := cfg.ToSimpleRequest()
           if err != nil {
               return err
           }
           // use req ...
           return nil
       },
   }
   simple.AddSimpleRequestFlags(cmd)
   ```

## CLI reference

```
gen-cobra-flags [flags]
```

| Flag             | Default | Description                                                          |
| ---------------- | ------- | -------------------------------------------------------------------- |
| `-input`         | `.`     | Directory or file to parse for annotated structs.                    |
| `-output`        | `.`     | Directory (or file) to write generated files to.                     |
| `-package`       | —       | Package name for the generated files. **Required.**                  |
| `-struct`        | (all)   | Restrict generation to a single struct. Defaults to all annotated.   |
| `-source-import` | (derived) | Import path of the package containing the source structs. Auto-derived from the input directory's Go module when generating into a different package; set it to override the derived value. |

When `-output` resolves to the same directory as `-input`, the generated code is emitted into
the source package, so no source-package import or qualifier is produced. When generating into
a separate directory, the generator imports the source package (its import path is derived
automatically from the input directory's Go module, or taken from `-source-import` when set).


## Markers

The full set of supported `+cobra:*` markers is documented in
[docs/markers.md](docs/markers.md).

## Runtime SDK

Generated code imports a small runtime library that provides type adaptors and helpers:

```
github.com/gocloud9/gen-cobra-flags/sdk
```

It contains:

- `sdk/pkg/adaptors` — conversion functions between flag string values and Go types
  (e.g. `JsonOrYamlToStruct`, `StringToInteger`, `StringToFloat`, `StringToIP`).
- `sdk/pkg/defaults` — parsers for default values (`ParseDuration`, `ParseTime`, `ParseCIDR`).
- `sdk/pkg/it` — small helpers such as `Must`.

Add it to your module:

```sh
go get github.com/gocloud9/gen-cobra-flags/sdk
```

API docs: <https://pkg.go.dev/github.com/gocloud9/gen-cobra-flags/sdk>

## Repository layout

This repository is a Go workspace composed of three modules:

| Module                                          | Path       | Purpose                                                              |
| ----------------------------------------------- | ---------- | -------------------------------------------------------------------- |
| `github.com/gocloud9/gen-cobra-flags`           | `/`        | The generator CLI (`cmd/gen-cobra-flags`) and its internals.         |
| `github.com/gocloud9/gen-cobra-flags/sdk`       | `/sdk`     | Runtime library imported by generated code.                          |
| `github.com/gocloud9/gen-cobra-flags/example`   | `/example` | A worked example / fixture demonstrating the generator end to end.   |

## Development

A `Makefile` wraps the common workflows across all three modules:

```sh
make build      # build the CLI
make test       # run tests in every module
make lint       # run golangci-lint (auto-installs the pinned version)
make fmt        # format and tidy imports
make generate   # regenerate the example output
make check      # vet + lint + test across all modules
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines.

## License

Licensed under the [MIT License](LICENSE).
