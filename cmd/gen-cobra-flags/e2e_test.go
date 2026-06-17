package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestCLI_EndToEnd builds the gen-cobra-flags binary, invokes it the way a
// //go:generate directive would, and compiles the produced output inside a
// throwaway module. This exercises the real `go install` + CLI path end to end
// (flag parsing, template embedding, file writing), which the in-process
// generator tests do not cover.
func TestCLI_EndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping end-to-end CLI test in -short mode")
	}

	root := repoRoot(t)
	tmp := t.TempDir()

	// 1. Build the CLI binary.
	bin := filepath.Join(tmp, "gen-cobra-flags")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	build := exec.Command("go", "build", "-o", bin, "./cmd/gen-cobra-flags")
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("building CLI failed: %v\n%s", err, out)
	}

	// 2. Lay out a throwaway module containing the source struct package.
	mod := filepath.Join(tmp, "mod")
	const modPath = "cli.test/m"
	const pkgName = "config"
	pkgDir := filepath.Join(mod, pkgName)
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	source := `package config

// Request is an annotated request struct.
// +cobra:flag=request
// +cobra:short=r
// +cobra:usage=A request
type Request struct {
	// +cobra:flag=label
	// +cobra:short=l
	// +cobra:usage=The label
	// +cobra:default=""
	Label string

	// +cobra:flag=retries
	// +cobra:usage=Retry count
	// +cobra:default=0
	Retries int
}
`
	writeFile(t, filepath.Join(pkgDir, "config.go"), source)

	sdkDir := filepath.Join(root, "sdk")
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
	writeFile(t, filepath.Join(mod, "go.mod"), goMod)

	// 3. Invoke the CLI exactly as a //go:generate directive would, in
	// same-package mode so the output lands beside the source struct.
	gen := exec.Command(bin,
		"-input", ".",
		"-output", ".",
		"-package", pkgName,
		"-same-package",
	)
	gen.Dir = pkgDir
	if out, err := gen.CombinedOutput(); err != nil {
		t.Fatalf("running CLI failed: %v\n%s", err, out)
	}

	genFiles, _ := filepath.Glob(filepath.Join(pkgDir, "*_gen.go"))
	if len(genFiles) == 0 {
		t.Fatal("CLI produced no *_gen.go files")
	}

	// 4. Resolve deps and compile the generated package.
	runGoIn(t, mod, "mod", "tidy")
	runGoIn(t, mod, "build", "./...")
}

func repoRoot(t *testing.T) string {
	t.Helper()
	// cmd/gen-cobra-flags -> repo root is two levels up.
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolving repo root: %v", err)
	}
	return root
}

func toolchainGoVersion() string {
	v := strings.TrimPrefix(runtime.Version(), "go")
	parts := strings.Split(v, ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return v
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func runGoIn(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOFLAGS=-mod=mod")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go %s failed: %v\n%s", strings.Join(args, " "), err, out)
	}
}
