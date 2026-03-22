package component

import "github.com/yourname/gowasm/pkg/dom"

// Mount mounts component into the DOM element with the given ID.
// It blocks forever (required for WASM main goroutine).
func Mount(rootID string, component Component) {
	el, ok := dom.ElementFromID(rootID)
	if !ok {
		panic("component: no element found with id " + rootID)
	}

	root := C(component)
	root.Mount(el)

	// Block forever - required for WASM main goroutine
	select {}
}
