package main

import (
	"fmt"
	"os"

	"github.com/gocloud9/gen-cobra-flags/example"
)

func main() {
	cmd := example.NewExampleCommand()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
