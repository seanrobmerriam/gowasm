// Package dom provides ergonomic wrappers around syscall/js for working with
// the browser DOM in Go/WebAssembly.
package dom

import "syscall/js"

// Document is the global browser document singleton.
var doc js.Value

func init() {
	doc = js.Global().Get("document")
	if doc.IsNull() || doc.IsUndefined() {
		panic("dom: document is not available")
	}
}

// Doc returns the global Document wrapper.
func Doc() Document {
	return Document{val: doc}
}

// Document wraps the browser's Document object.
type Document struct {
	val js.Value
}

// GetElementByID returns the element with the given ID, or (zero, false) if not found.
func (d Document) GetElementByID(id string) (Element, bool) {
	el := d.val.Call("getElementById", id)
	if el.IsNull() || el.IsUndefined() {
		return Element{}, false
	}
	return Element{val: el}, true
}

// QuerySelector returns the first element matching the CSS selector, or (zero, false).
func (d Document) QuerySelector(selector string) (Element, bool) {
	el := d.val.Call("querySelector", selector)
	if el.IsNull() || el.IsUndefined() {
		return Element{}, false
	}
	return Element{val: el}, true
}

// CreateElement creates a new DOM element with the given tag name.
func (d Document) CreateElement(tag string) Element {
	return Element{val: d.val.Call("createElement", tag)}
}

// CreateTextNode creates a new text node.
func (d Document) CreateTextNode(text string) TextNode {
	return TextNode{val: d.val.Call("createTextNode", text)}
}

// Body returns the document body element.
func (d Document) Body() Element {
	return Element{val: d.val.Get("body")}
}

// Head returns the document head element.
func (d Document) Head() Element {
	return Element{val: d.val.Get("head")}
}
