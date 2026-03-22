package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yourname/gowasm/cmd/gowasm/internal/build"
	"github.com/yourname/gowasm/cmd/gowasm/internal/reload"
	"github.com/yourname/gowasm/cmd/gowasm/internal/serve"
	"github.com/yourname/gowasm/cmd/gowasm/internal/watch"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "build":
		runBuild()
	case "serve":
		runServe()
	case "dev":
		runDev()
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <command> [flags]\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  build   Compile to WebAssembly\n")
	fmt.Fprintf(os.Stderr, "  serve   Serve a built app\n")
	fmt.Fprintf(os.Stderr, "  dev     Watch, rebuild, and live-reload\n")
	fmt.Fprintf(os.Stderr, "\nFlags (all subcommands):\n")
	fmt.Fprintf(os.Stderr, "  -dir    string   directory containing main.go (default \".\")\n")
	fmt.Fprintf(os.Stderr, "  -out    string   output .wasm filename (default \"app.wasm\")\n")
	fmt.Fprintf(os.Stderr, "  -port   int      HTTP port (default 8080)\n")
	fmt.Fprintf(os.Stderr, "  -host   string   HTTP host (default \"localhost\")\n")
}

func runBuild() {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	dir := fs.String("dir", ".", "source directory")
	out := fs.String("out", "app.wasm", "output .wasm filename")
	if err := fs.Parse(os.Args[2:]); err != nil {
		os.Exit(1)
	}

	opts := build.Options{
		Dir:     *dir,
		OutFile: *out,
	}
	if err := build.Run(opts); err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Built: %s\n", *out)
}

func runServe() {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	dir := fs.String("dir", ".", "directory to serve")
	out := fs.String("out", "app.wasm", "output .wasm filename")
	port := fs.Int("port", 8080, "HTTP port")
	host := fs.String("host", "localhost", "HTTP host")
	if err := fs.Parse(os.Args[2:]); err != nil {
		os.Exit(1)
	}

	opts := serve.Options{
		Dir:      *dir,
		WasmFile: *out,
		Port:     *port,
		Host:     *host,
		Reload:   nil,
	}
	if err := serve.ListenAndServe(opts); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}

func runDev() {
	fs := flag.NewFlagSet("dev", flag.ContinueOnError)
	dir := fs.String("dir", ".", "source directory")
	out := fs.String("out", "app.wasm", "output .wasm filename")
	port := fs.Int("port", 8080, "HTTP port")
	host := fs.String("host", "localhost", "HTTP host")
	if err := fs.Parse(os.Args[2:]); err != nil {
		os.Exit(1)
	}

	// Initial build
	buildOpts := build.Options{
		Dir:     *dir,
		OutFile: *out,
	}
	if err := build.Run(buildOpts); err != nil {
		fmt.Fprintf(os.Stderr, "initial build failed: %v\n", err)
		// Continue watching anyway
	}

	// Start reload hub
	hub := reload.New()

	// Start server in goroutine
	go func() {
		wasmPath, _ := os.Getwd()
		opts := serve.Options{
			Dir:      *dir,
			WasmFile: wasmPath + "/" + *out,
			Port:     *port,
			Host:     *host,
			Reload:   hub,
		}
		if err := serve.ListenAndServe(opts); err != nil {
			fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		}
	}()

	fmt.Printf("gowasm dev: http://%s:%d\n", *host, *port)

	// Start watcher
	watchOpts := watch.Options{
		Dir: *dir,
		Build: func() error {
			fmt.Fprintf(os.Stderr, "Building...\n")
			return build.Run(buildOpts)
		},
		Reload: func() {
			fmt.Fprintf(os.Stderr, "Reloading...\n")
			hub.Broadcast()
		},
	}
	if err := watch.Watch(watchOpts); err != nil {
		fmt.Fprintf(os.Stderr, "watch error: %v\n", err)
		os.Exit(1)
	}
}
