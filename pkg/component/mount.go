package component

import (
	"github.com/seanrobmerriam/gowasm/pkg/dom"
	"github.com/seanrobmerriam/gowasm/pkg/reactive"
	"github.com/seanrobmerriam/gowasm/pkg/vdom"
)

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

// HydrateMount is used by the hydration client to mount a component
// with a pre-existing initial render that has already been adopted
// from server-rendered DOM. This skips the first DOM mount and goes
// directly into the reactive patch loop.
func HydrateMount(
	parent dom.Element,
	comp Component,
	initial Node,
) {
	cn := &ComponentNode{
		vnode:    vdom.NewComponent("", comp),
		rootNode: initial,
		mounted:  true,
	}

	var firstRun = true

	cn.effect = reactive.NewEffect(func() {
		if firstRun {
			firstRun = false
			// Initial node already adopted — skip mount, just track signals.
			_ = comp.Render()
			return
		}

		newRoot := comp.Render()
		oldRoot := cn.rootNode
		cn.rootNode = newRoot
		if !cn.rootNode.Patch(oldRoot) {
			oldRoot.Unmount()
			cn.rootNode.Mount(parent)
		}
	})
}
