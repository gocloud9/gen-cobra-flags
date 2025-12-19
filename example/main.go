package example

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewExampleCommand() *cobra.Command {
	config := &Config{}

	cmd := &cobra.Command{
		Use:   "example",
		Short: "Example command demonstrating generated flags",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Configuration:\n")
			fmt.Printf("  Host: %s\n", config.Host)
			fmt.Printf("  Port: %d\n", config.Port)
			fmt.Printf("  Debug: %v\n", config.Debug)
			fmt.Printf("  Verbose: %v\n", config.Verbose)
			fmt.Printf("  LogLevel: %s\n", config.LogLevel)
		},
	}

	// Add generated flags
	AddConfigFlags(cmd, config)

	return cmd
}
