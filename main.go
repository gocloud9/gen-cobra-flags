package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var (
		inputFile  = flag.String("input", "", "Input file containing struct definition")
		outputFile = flag.String("output", "", "Output file for generated code")
		structName = flag.String("struct", "", "Name of the struct to generate flags for")
		pkgName    = flag.String("package", "", "Package name for generated code")
	)

	flag.Parse()

	if *inputFile == "" || *structName == "" {
		fmt.Fprintf(os.Stderr, "Usage: gen-cobra-flags -input <file> -struct <name> [-output <file>] [-package <name>]\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Set defaults
	if *outputFile == "" {
		*outputFile = "flags_gen.go"
	}
	if *pkgName == "" {
		*pkgName = "main"
	}

	err := generateFlags(*inputFile, *outputFile, *structName, *pkgName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated flags in %s\n", *outputFile)
}
