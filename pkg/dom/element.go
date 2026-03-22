package dom

import "syscall/js"

// Element wraps a JS DOM element.
type Element struct {
	val js.Value
}

// NewElement creates a new DOM element with the given tag name.
func NewElement(tag string) Element {
	return Doc().CreateElement(tag)
}

// ElementFromID returns the element with the given ID, or (zero, false) if not found.
func ElementFromID(id string) (Element, bool) {
	return Doc().GetElementByID(id)
}

// SetAttr sets an HTML attribute on the element.
func (e Element) SetAttr(key, val string) {
	e.val.Call("setAttribute", key, val)
}

// RemoveAttr removes an HTML attribute from the element.
func (e Element) RemoveAttr(key string) {
	e.val.Call("removeAttribute", key)
}

// SetStyle sets a CSS style property on the element.
func (e Element) SetStyle(prop, val string) {
	e.val.Get("style").Call("setProperty", prop, val)
}

// AppendChild adds a child element to this element.
func (e Element) AppendChild(child Element) {
	e.val.Call("appendChild", child.val)
}

// RemoveChild removes a child element from this element.
func (e Element) RemoveChild(child Element) {
	e.val.Call("removeChild", child.val)
}

// Remove removes this element from its parent.
func (e Element) Remove() {
	e.val.Call("remove")
}

// SetInnerHTML sets the innerHTML of the element.
func (e Element) SetInnerHTML(html string) {
	e.val.Set("innerHTML", html)
}

// SetTextContent sets the text content of the element.
func (e Element) SetTextContent(text string) {
	e.val.Set("textContent", text)
}

// SetProperty sets a JS property on the element.
func (e Element) SetProperty(key string, val any) {
	e.val.Set(key, val)
}

// AddEventListener registers an event handler and returns a handle for removal.
func (e Element) AddEventListener(event string, fn EventHandler) ListenerHandle {
	wrapper := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) == 0 {
			return nil
		}
		fn(Event{val: args[0]})
		return nil
	})
	e.val.Call("addEventListener", event, wrapper)
	return ListenerHandle{fn: wrapper, event: event, target: e.val}
}

// RemoveEventListener removes an event listener using the handle.
func (e Element) RemoveEventListener(handle ListenerHandle) {
	handle.Release()
}

// ListenerHandle is an opaque handle for removing event listeners.
type ListenerHandle struct {
	fn     js.Func
	event  string
	target js.Value
}

// Release releases the listener function.
func (h ListenerHandle) Release() {
	h.target.Call("removeEventListener", h.event, h.fn)
	h.fn.Release()
}
