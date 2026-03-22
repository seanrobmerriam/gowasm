package main

import (
	"github.com/seanrobmerriam/gowasm/pkg/component"
	"github.com/seanrobmerriam/gowasm/pkg/dom"
)

func main() {
	// Mount the root App component into <div id="app"> in index.html.
	app := component.Mount(
		dom.QuerySelector("#app"),
		AppComponent(),
	)

	// Block forever — the Wasm module must not exit.
	// Mount returns a teardown function; we ignore it here because
	// the app owns the page for its entire lifetime.
	_ = app
	select {}
}
