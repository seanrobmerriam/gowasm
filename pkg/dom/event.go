//go:build js && wasm

package dom

import "syscall/js"

// EventHandler is a function that handles a DOM event.
type EventHandler func(e Event)

// Event wraps a JS event object.
type Event struct {
	val js.Value
}

// Type returns the event type string (e.g., "click", "input").
func (e Event) Type() string {
	return e.val.Get("type").String()
}

// Target returns the Element that dispatched the event.
func (e Event) Target() Element {
	return Element{val: e.val.Get("target")}
}

// PreventDefault calls preventDefault on the JS event.
func (e Event) PreventDefault() {
	e.val.Call("preventDefault")
}

// StopPropagation calls stopPropagation on the JS event.
func (e Event) StopPropagation() {
	e.val.Call("stopPropagation")
}

// Value returns the underlying js.Value for advanced interop.
// Use with caution: this bypasses the type-safe wrapper.
func (e Event) Value(key string) js.Value {
	return e.val.Get(key)
}
