// Command gen-cobra-flags generates Cobra flag-binding boilerplate from Go
// structs annotated with +cobra:* markers.
//
// Usage:
//
//	gen-cobra-flags -input <dir> -output <dir> [-package <name>] [-struct <Name>] [-source-import <path>]
//	gen-cobra-flags -version
//
// When -output resolves to the same directory as -input, the generated code is
// emitted into the source package (no source-package import or qualifier) and
// -package is optional: the package name is derived from the input directory.
// When -output differs from -input, -package is required.
// It is typically invoked via a //go:generate directive.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/gocloud9/gen-cobra-flags/internal/generator"
)

// version is the CLI's release version. It defaults to "dev" for local
// builds and is overridden at build time via:
//
//	go build -ldflags "-X main.version=v1.2.3"
//
// as done by `make build`/`make install`. When installed via
// `go install .../gen-cobra-flags@vX.Y.Z` without ldflags, the version is
// instead recovered from the module's build info at runtime.
var version = "dev"

func main() {
	var (
		input        = flag.String("input", ".", "directory to parse for annotated structs")
		output       = flag.String("output", ".", "directory to write generated files to")
		pkg          = flag.String("package", "", "package name for the generated files (optional when input and output are the same directory)")
		structName   = flag.String("struct", "", "restrict generation to a single struct (default: all annotated structs)")
		sourceImport = flag.String("source-import", "", "import path of the package containing the source structs")
		showVersion  = flag.Bool("version", false, "print the gen-cobra-flags version and exit")
	)
	flag.Parse()

	if *showVersion {
		fmt.Println(resolveVersion())
		return
	}

	if err := run(generator.Options{
		InputDir:     *input,
		OutputDir:    *output,
		Package:      *pkg,
		Struct:       *structName,
		SourceImport: *sourceImport,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "gen-cobra-flags: %v\n", err)
		os.Exit(1)
	}
}

func run(opts generator.Options) error {
	return generator.Generate(opts)
}

// resolveVersion returns the CLI version to display. If version was not set
// at build time via -ldflags (i.e. it is still "dev"), it falls back to the
// version recorded in the binary's embedded module build info, which `go
// install pkg@version` populates automatically.
func resolveVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version
}
