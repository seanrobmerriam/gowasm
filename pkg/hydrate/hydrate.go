//go:build js && wasm

// Package hydrate attaches a gowasm component to server-rendered HTML.
// It must be used instead of component.Mount when the page was rendered
// server-side via pkg/ssr.
package hydrate

import (
	"github.com/seanrobmerriam/gowasm/pkg/component"
	"github.com/seanrobmerriam/gowasm/pkg/dom"
)

var retainedHandles []dom.ListenerHandle

// Hydrate attaches component to the DOM element with the given ID.
// If the element has data-ssr="true", hydration mode is used.
// If not, falls back to a normal component.Mount.
// Blocks forever (required for WASM main goroutine).
func Hydrate(rootID string, comp component.Component) {
	el, ok := dom.ElementFromID(rootID)
	if !ok {
		panic("hydrate: no element found with id " + rootID)
	}
	HydrateElement(el, comp)
	select {}
}

// HydrateElement is the lower-level form that accepts a dom.Element directly.
func HydrateElement(root dom.Element, comp component.Component) {
	isSSR := root.JSValue().Get("dataset").Get("ssr").String() == "true"

	if !isSSR {
		cn := component.C(comp)
		cn.Mount(root)
		return
	}

	claimed := buildClaimMap(root)
	initialRender := comp.Render()

	var handles []dom.ListenerHandle
	attachEvents(initialRender, claimed, []int{0}, &handles)
	retainedHandles = append(retainedHandles, handles...)
	adoptDOM(initialRender, claimed, []int{0})

	component.HydrateMount(root, comp, initialRender)
}
