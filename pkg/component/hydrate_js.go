//go:build js && wasm

package component

import (
	"syscall/js"

	"github.com/seanrobmerriam/gowasm/pkg/dom"
)

// AdoptElement sets the live DOM element reference on this node.
// Used by the hydration client to attach server-rendered DOM nodes
// to the component tree without remounting.
func (n *ElementNode) AdoptElement(val interface{}) {
	if jsVal, ok := val.(js.Value); ok {
		n.domEl = dom.ElementFromJSValue(jsVal)
	}
}

// AdoptTextNode sets the live DOM text node reference on this node.
func (n *TextNode) AdoptTextNode(val interface{}) {
	if jsVal, ok := val.(js.Value); ok {
		n.domEl = dom.TextNodeFromJSValue(jsVal)
	}
}
