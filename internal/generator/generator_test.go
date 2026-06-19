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

// TestDeriveSourceImport verifies the import path derived for a directory from
// its enclosing Go module. The test runs inside the gen-cobra-flags module, so
// the current directory resolves to a known import path.
func TestDeriveSourceImport(t *testing.T) {
	const moduleRoot = "github.com/gocloud9/gen-cobra-flags"

	// "." is internal/generator within the module.
	got, err := deriveSourceImport(".")
	if err != nil {
		t.Fatalf("deriveSourceImport(.) error: %v", err)
	}
	if want := moduleRoot + "/internal/generator"; got != want {
		t.Errorf("deriveSourceImport(.) = %q, want %q", got, want)
	}

	// The example fixture is a separate directory in the same module.
	got, err = deriveSourceImport(filepath.Join("..", "..", "example"))
	if err != nil {
		t.Fatalf("deriveSourceImport(example) error: %v", err)
	}
	if want := moduleRoot + "/example"; got != want {
		t.Errorf("deriveSourceImport(example) = %q, want %q", got, want)
	}

	// The module root resolves to the bare module path.
	got, err = deriveSourceImport(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("deriveSourceImport(root) error: %v", err)
	}
	if got != moduleRoot {
		t.Errorf("deriveSourceImport(root) = %q, want %q", got, moduleRoot)
	}

	// A directory with no enclosing module returns an error.
	if _, err := deriveSourceImport(t.TempDir()); err == nil {
		t.Error("deriveSourceImport(tempdir) expected error, got nil")
	}
}

// TestModulePath verifies module-path extraction from go.mod contents.
func TestModulePath(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "module github.com/acme/app\n\ngo 1.22\n", "github.com/acme/app"},
		{"trailing comment", "module github.com/acme/app // v2\n", "github.com/acme/app"},
		{"leading comment", "// header\nmodule github.com/acme/app\n", "github.com/acme/app"},
		{"none", "go 1.22\n", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := modulePath([]byte(tc.in)); got != tc.want {
				t.Errorf("modulePath() = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestSameDir verifies the directory-equality detection that drives the
// same-package behavior: a path compared against itself (including via "." and
// trailing slashes) is the same directory, while distinct directories are not.
func TestSameDir(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatalf("creating subdir: %v", err)
	}

	cases := []struct {
		name string
		a, b string
		want bool
	}{
		{"identical", dir, dir, true},
		{"trailing slash", dir, dir + string(filepath.Separator), true},
		{"dot-normalized", dir, filepath.Join(dir, "sub", ".."), true},
		{"distinct", dir, sub, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := sameDir(tc.a, tc.b)
			if err != nil {
				t.Fatalf("sameDir(%q, %q) error: %v", tc.a, tc.b, err)
			}
			if got != tc.want {
				t.Errorf("sameDir(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}
