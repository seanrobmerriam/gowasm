//go:build js && wasm

package dom

import "syscall/js"

// JSValue returns the underlying js.Value of an Element.
// This is intended for use in tests only.
func (e Element) JSValue() js.Value {
	return e.val
}

// JSValue returns the underlying js.Value of a TextNode.
func (t TextNode) JSValue() js.Value {
	return t.val
}
