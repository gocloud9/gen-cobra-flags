package main

import (
	"os"
	"strings"
	"testing"
)

func TestGenerateFlags(t *testing.T) {
	// Create a temporary test file
	testInput := `package test

type TestConfig struct {
	Name string ` + "`" + `flag:"name" short:"n" usage:"Test name" default:"test"` + "`" + `
	Count int ` + "`" + `flag:"count" short:"c" usage:"Test count" default:"10"` + "`" + `
	Enabled bool ` + "`" + `flag:"enabled" short:"e" usage:"Enable feature"` + "`" + `
}
`
	inputFile := "test_input.go"
	outputFile := "test_output.go"

	// Write test input
	if err := os.WriteFile(inputFile, []byte(testInput), 0644); err != nil {
		t.Fatalf("Failed to write test input: %v", err)
	}
	defer os.Remove(inputFile)
	defer os.Remove(outputFile)

	// Generate flags
	err := generateFlags(inputFile, outputFile, "TestConfig", "test")
	if err != nil {
		t.Fatalf("generateFlags failed: %v", err)
	}

	// Check output file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatalf("Output file was not created")
	}

	// Read and verify output
	output, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}

	outputStr := string(output)

	// Verify key elements exist in output
	expectedStrings := []string{
		"package test",
		"AddTestConfigFlags",
		"cmd.Flags().StringVarP",
		"cmd.Flags().IntVarP",
		"cmd.Flags().BoolVarP",
		`"name"`,
		`"count"`,
		`"enabled"`,
		`"Test name"`,
		`"Test count"`,
	}

	for _, expected := range expectedStrings {
		if !contains(outputStr, expected) {
			t.Errorf("Output missing expected string: %q", expected)
		}
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
