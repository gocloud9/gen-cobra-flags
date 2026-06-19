// Command gen-cobra-flags generates Cobra flag-binding boilerplate from Go
// structs annotated with +cobra:* markers.
//
// Usage:
//
//	gen-cobra-flags -input <dir> -output <dir> -package <name> [-struct <Name>] [-source-import <path>]
//
// When -output resolves to the same directory as -input, the generated code is
// emitted into the source package (no source-package import or qualifier).
// It is typically invoked via a //go:generate directive.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gocloud9/gen-cobra-flags/internal/generator"
)

func main() {
	var (
		input        = flag.String("input", ".", "directory to parse for annotated structs")
		output       = flag.String("output", ".", "directory to write generated files to")
		pkg          = flag.String("package", "", "package name for the generated files (required)")
		structName   = flag.String("struct", "", "restrict generation to a single struct (default: all annotated structs)")
		sourceImport = flag.String("source-import", "", "import path of the package containing the source structs")
	)
	flag.Parse()

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
