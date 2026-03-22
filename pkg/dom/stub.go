//go:build !js || !wasm

package dom

import "errors"

var errUnavailable = errors.New("dom: not available outside js/wasm")

// Document is a stub browser document for non-js builds.
type Document struct{}

// Element is a stub DOM element for non-js builds.
type Element struct{}

// TextNode is a stub DOM text node for non-js builds.
type TextNode struct{}

// EventHandler is a function that handles a DOM event.
type EventHandler func(e Event)

// Event is a stub event type for non-js builds.
type Event struct{}

// Value is a stub value type that mirrors the small subset of methods used by
// the component package in non-js builds.
type Value struct{}

// ListenerHandle is an opaque handle for removing event listeners.
type ListenerHandle struct{}

// Doc returns a stub document.
func Doc() Document { return Document{} }

// NewElement is unavailable outside js/wasm.
func NewElement(tag string) Element { panic(errUnavailable) }

// ElementFromID is unavailable outside js/wasm.
func ElementFromID(id string) (Element, bool) { panic(errUnavailable) }

// GetElementByID is unavailable outside js/wasm.
func (d Document) GetElementByID(id string) (Element, bool) { panic(errUnavailable) }

// QuerySelector is unavailable outside js/wasm.
func (d Document) QuerySelector(selector string) (Element, bool) { panic(errUnavailable) }

// CreateElement is unavailable outside js/wasm.
func (d Document) CreateElement(tag string) Element { panic(errUnavailable) }

// CreateTextNode is unavailable outside js/wasm.
func (d Document) CreateTextNode(text string) TextNode { panic(errUnavailable) }

// Body is unavailable outside js/wasm.
func (d Document) Body() Element { panic(errUnavailable) }

// Head is unavailable outside js/wasm.
func (d Document) Head() Element { panic(errUnavailable) }

// SetAttr is unavailable outside js/wasm.
func (e Element) SetAttr(key, val string) { panic(errUnavailable) }

// RemoveAttr is unavailable outside js/wasm.
func (e Element) RemoveAttr(key string) { panic(errUnavailable) }

// SetStyle is unavailable outside js/wasm.
func (e Element) SetStyle(prop, val string) { panic(errUnavailable) }

// AppendChild is unavailable outside js/wasm.
func (e Element) AppendChild(child Element) { panic(errUnavailable) }

// RemoveChild is unavailable outside js/wasm.
func (e Element) RemoveChild(child Element) { panic(errUnavailable) }

// Remove is unavailable outside js/wasm.
func (e Element) Remove() { panic(errUnavailable) }

// SetInnerHTML is unavailable outside js/wasm.
func (e Element) SetInnerHTML(html string) { panic(errUnavailable) }

// SetTextContent is unavailable outside js/wasm.
func (e Element) SetTextContent(text string) { panic(errUnavailable) }

// SetProperty is unavailable outside js/wasm.
func (e Element) SetProperty(key string, val any) { panic(errUnavailable) }

// AddEventListener is unavailable outside js/wasm.
func (e Element) AddEventListener(event string, fn EventHandler) ListenerHandle {
	panic(errUnavailable)
}

// RemoveEventListener is unavailable outside js/wasm.
func (e Element) RemoveEventListener(handle ListenerHandle) { panic(errUnavailable) }

// Release is unavailable outside js/wasm.
func (h ListenerHandle) Release() {}

// Type is unavailable outside js/wasm.
func (e Event) Type() string { panic(errUnavailable) }

// Target is unavailable outside js/wasm.
func (e Event) Target() Element { panic(errUnavailable) }

// PreventDefault is unavailable outside js/wasm.
func (e Event) PreventDefault() { panic(errUnavailable) }

// StopPropagation is unavailable outside js/wasm.
func (e Event) StopPropagation() { panic(errUnavailable) }

// Value is unavailable outside js/wasm.
func (e Event) Value(key string) Value { panic(errUnavailable) }

// Get is unavailable outside js/wasm.
func (v Value) Get(key string) Value { panic(errUnavailable) }

// Int is unavailable outside js/wasm.
func (v Value) Int() int { panic(errUnavailable) }

// String is unavailable outside js/wasm.
func (v Value) String() string { panic(errUnavailable) }

// Bool is unavailable outside js/wasm.
func (v Value) Bool() bool { panic(errUnavailable) }

// InputValue is unavailable outside js/wasm.
func InputValue(e Event) string { panic(errUnavailable) }

// CheckboxChecked is unavailable outside js/wasm.
func CheckboxChecked(e Event) bool { panic(errUnavailable) }

// SelectValue is unavailable outside js/wasm.
func SelectValue(e Event) string { panic(errUnavailable) }

// InputValueFromElement is unavailable outside js/wasm.
func InputValueFromElement(el Element) string { panic(errUnavailable) }

// CheckboxCheckedFromElement is unavailable outside js/wasm.
func CheckboxCheckedFromElement(el Element) bool { panic(errUnavailable) }

// SetText is unavailable outside js/wasm.
func (t TextNode) SetText(text string) { panic(errUnavailable) }

// Remove is unavailable outside js/wasm.
func (t TextNode) Remove() { panic(errUnavailable) }

// AppendTo is unavailable outside js/wasm.
func (t TextNode) AppendTo(parent Element) { panic(errUnavailable) }
