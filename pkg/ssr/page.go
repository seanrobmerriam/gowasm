package ssr

import "fmt"

// PageOptions configures the HTML page wrapper.
type PageOptions struct {
	Title       string
	Lang        string // default "en"
	RootID      string // default "app"
	Head        string // additional <head> content (raw HTML)
	WasmSrc     string // path to .wasm file, default "/app.wasm"
	WasmExecSrc string // path to wasm_exec.js, default "/wasm_exec.js"
}

// RenderPage wraps an HTML fragment in a full page document.
// The fragment is injected into a div with id=opts.RootID.
// The wasm bootstrap script is appended to load the client-side app
// for hydration.
func RenderPage(fragment string, opts PageOptions) string {
	lang := opts.Lang
	if lang == "" {
		lang = "en"
	}
	rootID := opts.RootID
	if rootID == "" {
		rootID = "app"
	}
	wasmSrc := opts.WasmSrc
	if wasmSrc == "" {
		wasmSrc = "/app.wasm"
	}
	wasmExecSrc := opts.WasmExecSrc
	if wasmExecSrc == "" {
		wasmExecSrc = "/wasm_exec.js"
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="%s">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>%s</title>
%s
</head>
<body>
  <div id="%s" data-ssr="true">%s</div>
  <script src="%s"></script>
  <script>
    const go = new Go();
    WebAssembly.instantiateStreaming(fetch("%s"), go.importObject)
      .then(result => go.run(result.instance));
  </script>
</body>
</html>
`, escapeHTML(lang), escapeHTML(opts.Title), opts.Head, escapeHTML(rootID), fragment, escapeHTML(wasmExecSrc), escapeHTML(wasmSrc))
}
