package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/seanrobmerriam/gowasm/cmd/gowasm/internal/build"
	"github.com/seanrobmerriam/gowasm/cmd/gowasm/internal/pidfile"
	"github.com/seanrobmerriam/gowasm/cmd/gowasm/internal/reload"
	"github.com/seanrobmerriam/gowasm/cmd/gowasm/internal/scaffold"
	"github.com/seanrobmerriam/gowasm/cmd/gowasm/internal/serve"
	"github.com/seanrobmerriam/gowasm/cmd/gowasm/internal/watch"
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
	case "new":
		runNew()
	case "stop":
		runStop()
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
	fmt.Fprintf(os.Stderr, "  new     Scaffold a new project\n")
	fmt.Fprintf(os.Stderr, "  stop    Stop a running serve or dev server\n")
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

	// Write PID file
	if err := pidfile.Write(*dir); err != nil {
		fmt.Fprintf(os.Stderr, "error writing pid file: %v\n", err)
		os.Exit(1)
	}
	defer pidfile.Remove(*dir)

	// Handle signals for clean PID file removal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		pidfile.Remove(*dir)
		os.Exit(0)
	}()

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

	// Write PID file
	if err := pidfile.Write(*dir); err != nil {
		fmt.Fprintf(os.Stderr, "error writing pid file: %v\n", err)
		os.Exit(1)
	}
	defer pidfile.Remove(*dir)

	// Handle signals for clean PID file removal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		pidfile.Remove(*dir)
		os.Exit(0)
	}()

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

func runNew() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: gowasm new <name> [-module <mod>]\n")
		os.Exit(1)
	}

	name := os.Args[2]

	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	defaultModule := "github.com/seanrobmerriam/" + name
	module := fs.String("module", defaultModule, "Go module path")

	if err := fs.Parse(os.Args[3:]); err != nil {
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	frameworkDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	opts := scaffold.Options{
		Name:         name,
		Module:       *module,
		TargetDir:    filepath.Join(cwd, name),
		FrameworkDir: frameworkDir,
	}

	if err := scaffold.Run(opts); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nCreated %s\n\n", name)
	fmt.Printf("Next steps:\n")
	fmt.Printf("  cd %s\n", name)
	fmt.Printf("  go mod tidy\n")
	fmt.Printf("  gowasm dev -dir .\n")
	fmt.Printf("\nNote: go.mod contains a replace directive pointing at the local\n")
	fmt.Printf("gowasm source. Update or remove it once the framework is published.\n")
}

func runStop() {
	fs := flag.NewFlagSet("stop", flag.ContinueOnError)
	dir := fs.String("dir", ".", "project directory")
	if err := fs.Parse(os.Args[2:]); err != nil {
		os.Exit(1)
	}

	pid, err := pidfile.Read(*dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "no server running in %s (no pid file found)\n", *dir)
		os.Exit(1)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not find process %d: %v\n", pid, err)
		pidfile.Remove(*dir)
		os.Exit(1)
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		fmt.Fprintf(os.Stderr, "could not signal process %d: %v\n", pid, err)
		pidfile.Remove(*dir)
		os.Exit(1)
	}

	fmt.Printf("stopped server (pid %d)\n", pid)
}
