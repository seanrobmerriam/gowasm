package serve

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

// Options configures the dev/static server.
type Options struct {
	Dir      string // directory to serve static files from
	WasmFile string // absolute path to the compiled .wasm file
	Port     int
	Host     string
	Reload   ReloadHub // nil in serve mode, non-nil in dev mode
}

// ReloadHub is implemented by reload.Hub.
type ReloadHub interface {
	Handler() http.Handler // WebSocket upgrade handler at /~reload
}

// ListenAndServe starts the HTTP server and blocks.
func ListenAndServe(opts Options) error {
	addr := fmt.Sprintf("%s:%d", opts.Host, opts.Port)
	return http.ListenAndServe(addr, handler(opts))
}

func handler(opts Options) http.Handler {
	mux := http.NewServeMux()

	// wasm_exec.js
	mux.HandleFunc("/wasm_exec.js", func(w http.ResponseWriter, r *http.Request) {
		setHeaders(w)
		http.ServeFile(w, r, wasmExecPath())
	})

	// app.wasm
	mux.HandleFunc("/app.wasm", func(w http.ResponseWriter, r *http.Request) {
		setHeaders(w)
		w.Header().Set("Content-Type", "application/wasm")
		http.ServeFile(w, r, opts.WasmFile)
	})

	// index.html
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		setHeaders(w)
		path := filepath.Join(opts.Dir, r.URL.Path)
		if r.URL.Path == "/" {
			path = filepath.Join(opts.Dir, "index.html")
		}
		if _, err := os.Stat(path); os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
		data, err := os.ReadFile(path)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		if opts.Reload != nil {
			data = injectLivereload(data)
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(data)
	})

	// Reload WebSocket
	if opts.Reload != nil {
		mux.Handle("/~reload", opts.Reload.Handler())
	}

	return mux
}

func setHeaders(w http.ResponseWriter) {
	w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
	w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
}

func wasmExecPath() string {
	// Try Go 1.21+ location
	p := filepath.Join(runtime.GOROOT(), "misc", "wasm", "wasm_exec.js")
	if _, err := os.Stat(p); err == nil {
		return p
	}
	// Fallback to Go 1.24+ location
	p = filepath.Join(runtime.GOROOT(), "lib", "wasm", "wasm_exec.js")
	if _, err := os.Stat(p); err == nil {
		return p
	}
	// Return the primary path anyway; will 404 if missing
	return filepath.Join(runtime.GOROOT(), "misc", "wasm", "wasm_exec.js")
}

const livereloadScript = `<script>
(function() {
  var ws = new WebSocket("ws://" + location.host + "/~reload");
  ws.onmessage = function(e) { if (e.data === "reload") location.reload(); };
  ws.onclose = function() { setTimeout(function() { location.reload(); }, 500); };
})();
</script>`

func injectLivereload(data []byte) []byte {
	const marker = "<!-- livereload -->"
	script := []byte(livereloadScript)
	out := make([]byte, 0, len(data)+len(script))
	i := 0
	for {
		j := findSubstring(data, i, marker)
		if j < 0 {
			out = append(out, data[i:]...)
			break
		}
		out = append(out, data[i:j]...)
		out = append(out, script...)
		i = j + len(marker)
	}
	return out
}

func findSubstring(data []byte, start int, sub string) int {
	if len(sub) == 0 {
		return start
	}
	for i := start; i <= len(data)-len(sub); i++ {
		if string(data[i:i+len(sub)]) == sub {
			return i
		}
	}
	return -1
}

// writeWebSocketFrame writes a minimal WebSocket text frame.
func writeWebSocketFrame(w io.Writer, msg string) error {
	frame := make([]byte, 2+len(msg))
	frame[0] = 0x81 // FIN=1, opcode=1 (text)
	frame[1] = byte(len(msg))
	copy(frame[2:], msg)
	_, err := w.Write(frame)
	return err
}

// computeAcceptKey computes the WebSocket accept key.
func computeAcceptKey(key string) string {
	const magic = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	h := sha1.New()
	h.Write([]byte(key + magic))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
