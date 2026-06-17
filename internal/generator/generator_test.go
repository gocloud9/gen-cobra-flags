package generator

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

// update regenerates the golden files when set: go test ./internal/generator -update
var update = flag.Bool("update", false, "update golden files")

// exampleDir resolves the read-only example fixture module directory.
func exampleDir(t *testing.T) string {
	t.Helper()
	// internal/generator -> repo root -> example
	dir, err := filepath.Abs(filepath.Join("..", "..", "example"))
	if err != nil {
		t.Fatalf("resolving example dir: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "config.go")); err != nil {
		t.Fatalf("example fixture not found at %s: %v", dir, err)
	}
	return dir
}

// TestGenerate_ExampleFixture runs the generator against the example fixture
// and compares the output to checked-in golden files. The example module is
// treated as read-only input.
func TestGenerate_ExampleFixture(t *testing.T) {
	src := exampleDir(t)
	out := t.TempDir()

	err := Generate(Options{
		InputDir:     src,
		OutputDir:    out,
		Package:      "generated",
		SourceImport: "github.com/gocloud9/gen-cobra-flags/example",
	})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	goldenDir := filepath.Join("testdata", "golden")

	// Collect generated files.
	generated, err := filepath.Glob(filepath.Join(out, "*.go"))
	if err != nil {
		t.Fatalf("globbing generated files: %v", err)
	}
	if len(generated) == 0 {
		t.Fatal("generator produced no .go files")
	}

	for _, gen := range generated {
		name := filepath.Base(gen)
		got, err := os.ReadFile(gen)
		if err != nil {
			t.Fatalf("reading generated file %s: %v", name, err)
		}

		goldenPath := filepath.Join(goldenDir, name+".golden")

		if *update {
			if err := os.MkdirAll(goldenDir, 0o755); err != nil {
				t.Fatalf("creating golden dir: %v", err)
			}
			if err := os.WriteFile(goldenPath, got, 0o644); err != nil {
				t.Fatalf("writing golden file %s: %v", goldenPath, err)
			}
			continue
		}

		want, err := os.ReadFile(goldenPath)
		if err != nil {
			t.Fatalf("reading golden file %s (run with -update to create): %v", goldenPath, err)
		}

		if string(got) != string(want) {
			t.Errorf("generated %s does not match golden file %s\n--- got ---\n%s\n--- want ---\n%s",
				name, goldenPath, got, want)
		}
	}
}
