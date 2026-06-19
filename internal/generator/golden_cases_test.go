package generator

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// goldenCase describes a generator scenario exercised against checked-in
// golden output. Fixtures live under testdata/fixtures/<dir> and golden files
// under testdata/golden/<name>/.
type goldenCase struct {
	name    string
	fixture string // directory under testdata/fixtures
	opts    Options
}

func goldenCases() []goldenCase {
	return []goldenCase{
		{
			name:    "simple_same_package",
			fixture: "simple",
			opts: Options{
				Package:             "simple",
				SamePackageAsSource: true,
			},
		},
		{
			name:    "multiple_structs",
			fixture: "multi",
			opts: Options{
				Package:             "multi",
				SamePackageAsSource: true,
			},
		},
		{
			name:    "subcommands",
			fixture: "subcommands",
			opts: Options{
				Package:             "subcommands",
				SamePackageAsSource: true,
			},
		},
		{
			name:    "flagadaptor",
			fixture: "flagadaptor",
			opts: Options{
				Package:             "flagadaptor",
				SamePackageAsSource: true,
			},
		},
	}
}

// TestGenerate_GoldenCases generates each fixture and compares every produced
// file against checked-in golden output. Run with -update to regenerate.
func TestGenerate_GoldenCases(t *testing.T) {
	for _, tc := range goldenCases() {
		t.Run(tc.name, func(t *testing.T) {
			fixtureDir, err := filepath.Abs(filepath.Join("testdata", "fixtures", tc.fixture))
			if err != nil {
				t.Fatalf("resolving fixture dir: %v", err)
			}

			out := t.TempDir()
			opts := tc.opts
			opts.InputDir = fixtureDir
			opts.OutputDir = out

			if err := Generate(opts); err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			generated, err := filepath.Glob(filepath.Join(out, "*.go"))
			if err != nil {
				t.Fatalf("globbing generated files: %v", err)
			}
			if len(generated) == 0 {
				t.Fatal("generator produced no .go files")
			}
			sort.Strings(generated)

			goldenDir := filepath.Join("testdata", "golden", tc.name)

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

			// Verify no unexpected golden files linger (e.g. after a rename).
			if !*update {
				assertGoldenSetMatches(t, goldenDir, generated)
			}
		})
	}
}

// assertGoldenSetMatches fails if the golden directory contains golden files
// that were not produced by the current generator run.
func assertGoldenSetMatches(t *testing.T, goldenDir string, generated []string) {
	t.Helper()
	goldenFiles, err := filepath.Glob(filepath.Join(goldenDir, "*.golden"))
	if err != nil {
		t.Fatalf("globbing golden files: %v", err)
	}
	produced := map[string]bool{}
	for _, g := range generated {
		produced[filepath.Base(g)+".golden"] = true
	}
	for _, gf := range goldenFiles {
		base := filepath.Base(gf)
		if !produced[base] {
			t.Errorf("stale golden file %s has no corresponding generated output", base)
		}
	}
}
