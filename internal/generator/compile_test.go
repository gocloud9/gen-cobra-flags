package generator

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestGenerate_OutputCompiles is a hermetic end-to-end check: it generates code
// from the example fixture into a throwaway Go module that contains a copy of
// the source struct package, wires the generated package to depend on it, and
// runs `go build`. This proves the generated output is valid, compiling Go
// independent of the state of the example module's other files.
func TestGenerate_OutputCompiles(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compile check in -short mode")
	}

	repoRoot := repoRoot(t)
	src := filepath.Join(repoRoot, "example")

	tmp := t.TempDir()

	const modPath = "example.test/m"
	const sourcePkgImport = modPath + "/source"

	// 1. Copy the source struct package (config.go only) into the temp module.
	srcPkgDir := filepath.Join(tmp, "source")
	mustMkdir(t, srcPkgDir)
	copyFile(t, filepath.Join(src, "config.go"), filepath.Join(srcPkgDir, "config.go"))
	// Strip the go:generate directive line so it doesn't reference the binary.
	stripGoGenerate(t, filepath.Join(srcPkgDir, "config.go"))

	// 2. Write a go.mod that replaces the sdk with the in-repo copy. This must
	// exist before parsing/generating so the source package is loadable. The
	// go directive must match the running toolchain's version, otherwise
	// GOTOOLCHAIN=auto may select a different toolchain to satisfy it and the
	// build fails with a stdlib version mismatch.
	sdkDir := filepath.Join(repoRoot, "sdk")
	goMod := strings.Join([]string{
		"module " + modPath,
		"",
		"go " + toolchainGoVersion(),
		"",
		"require (",
		"\tgithub.com/gocloud9/gen-cobra-flags/sdk v0.0.0-pre",
		"\tgithub.com/spf13/cobra v1.10.2",
		")",
		"",
		"replace github.com/gocloud9/gen-cobra-flags/sdk => " + sdkDir,
		"",
	}, "\n")
	mustWrite(t, filepath.Join(tmp, "go.mod"), goMod)

	// 3. Generate the flags package into the temp module.
	genDir := filepath.Join(tmp, "generated")
	mustMkdir(t, genDir)
	if err := Generate(Options{
		InputDir:     srcPkgDir,
		OutputDir:    genDir,
		Package:      "generated",
		SourceImport: sourcePkgImport,
	}); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Guard against the generator silently producing no per-struct output.
	genFiles, _ := filepath.Glob(filepath.Join(genDir, "*_gen.go"))
	if len(genFiles) == 0 {
		t.Fatal("generator produced no *_gen.go files; the source package may not have been parsed")
	}

	// 4. Resolve remaining dependencies (cobra, yaml) and build everything.
	runGo(t, tmp, "mod", "tidy")
	runGo(t, tmp, "build", "./...")
}

// TestGenerate_FixturesCompile generates each same-package fixture into a
// throwaway Go module (alongside a copy of the fixture source) and runs
// `go build`, proving the generated output for those scenarios is valid Go.
func TestGenerate_FixturesCompile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping compile check in -short mode")
	}

	fixtures := []struct {
		name    string
		fixture string
		pkg     string
	}{
		{name: "simple", fixture: "simple", pkg: "simple"},
		{name: "multi", fixture: "multi", pkg: "multi"},
	}

	root := repoRoot(t)
	sdkDir := filepath.Join(root, "sdk")

	for _, f := range fixtures {
		t.Run(f.name, func(t *testing.T) {
			tmp := t.TempDir()
			modPath := "fixture.test/" + f.name

			goMod := strings.Join([]string{
				"module " + modPath,
				"",
				"go " + toolchainGoVersion(),
				"",
				"require (",
				"\tgithub.com/gocloud9/gen-cobra-flags/sdk v0.0.0-pre",
				"\tgithub.com/spf13/cobra v1.10.2",
				")",
				"",
				"replace github.com/gocloud9/gen-cobra-flags/sdk => " + sdkDir,
				"",
			}, "\n")
			mustWrite(t, filepath.Join(tmp, "go.mod"), goMod)

			// Same-package fixtures: source struct and generated code share one
			// package directory.
			pkgDir := filepath.Join(tmp, f.pkg)
			mustMkdir(t, pkgDir)

			fixtureSrc := filepath.Join(root, "internal", "generator", "testdata", "fixtures", f.fixture)
			srcFiles, _ := filepath.Glob(filepath.Join(fixtureSrc, "*.go"))
			if len(srcFiles) == 0 {
				t.Fatalf("no source files in fixture %s", fixtureSrc)
			}
			for _, sf := range srcFiles {
				copyFile(t, sf, filepath.Join(pkgDir, filepath.Base(sf)))
			}

			if err := Generate(Options{
				InputDir:            pkgDir,
				OutputDir:           pkgDir,
				Package:             f.pkg,
				SamePackageAsSource: true,
			}); err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			genFiles, _ := filepath.Glob(filepath.Join(pkgDir, "*_gen.go"))
			if len(genFiles) == 0 {
				t.Fatal("generator produced no *_gen.go files")
			}

			runGo(t, tmp, "mod", "tidy")
			runGo(t, tmp, "build", "./...")
		})
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolving repo root: %v", err)
	}
	return root
}

// toolchainGoVersion returns the running toolchain's go version as a
// "major.minor" string (e.g. "1.26") suitable for a go.mod go directive.
// Pinning the temp module to the exact running toolchain prevents
// GOTOOLCHAIN=auto from selecting a different version (which would otherwise
// fail with a stdlib "does not match go tool version" error).
func toolchainGoVersion() string {
	v := strings.TrimPrefix(runtime.Version(), "go") // e.g. "1.26.0" or "1.26"
	parts := strings.Split(v, ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return v
}

func mustMkdir(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func copyFile(t *testing.T, from, to string) {
	t.Helper()
	data, err := os.ReadFile(from)
	if err != nil {
		t.Fatalf("read %s: %v", from, err)
	}
	mustWrite(t, to, string(data))
}

// stripGoGenerate removes any //go:generate lines from a Go source file.
func stripGoGenerate(t *testing.T, path string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	lines := strings.Split(string(data), "\n")
	kept := lines[:0]
	for _, l := range lines {
		if strings.HasPrefix(strings.TrimSpace(l), "//go:generate") {
			continue
		}
		kept = append(kept, l)
	}
	mustWrite(t, path, strings.Join(kept, "\n"))
}

func runGo(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go %s failed: %v\n%s", strings.Join(args, " "), err, out)
	}
}
