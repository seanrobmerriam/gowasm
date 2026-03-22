package build

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Options configures a build run.
type Options struct {
	Dir     string // source directory (must contain package main)
	OutFile string // filename, e.g. "app.wasm"
}

// Run compiles the package at opts.Dir to WebAssembly.
// It sets GOOS=js and GOARCH=wasm and runs: go build -o <OutFile> <Dir>
// Stdout and stderr from the compiler are forwarded to os.Stderr.
// Returns a non-nil error if compilation fails.
func Run(opts Options) error {
	// Resolve the Go binary path
	goBin, err := exec.LookPath("go")
	if err != nil {
		// Fallback to GOROOT/bin/go
		goroot := runtime.GOROOT()
		goBin = filepath.Join(goroot, "bin", "go")
		if _, err := os.Stat(goBin); err != nil {
			return err
		}
	}

	// Build command
	cmd := exec.Command(goBin, "build", "-o", opts.OutFile, opts.Dir)
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd.Run()
}
