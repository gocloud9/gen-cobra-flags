package main

import (
	"fmt"
	"github.com/gocloud9/gen-cobra-flags/example/generated"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	cmd := NewExampleCommand()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func NewExampleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "example",
		Short: "Example command demonstrating generated flags",
		Run: func(cmd *cobra.Command, args []string) {
			requestConfig, err := generated.CreateFooBarRequestConfigFromFlags(cmd)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating CreateFooBarRequest: %v\n", err)
			}

			request, err := requestConfig.ToCreateFooBarRequest()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating CreateFooBarRequest: %v\n", err)
			}
			fmt.Println(request)
		},
	}
	generated.AddCreateFooBarRequestFlags(cmd)

	return cmd
}
