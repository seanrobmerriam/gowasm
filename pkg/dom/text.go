package dom

import "syscall/js"

// TextNode represents a DOM text node.
type TextNode struct {
	val js.Value
}

// SetText sets the text content of the text node.
func (t TextNode) SetText(text string) {
	t.val.Set("textContent", text)
}

// Remove removes this text node from its parent.
func (t TextNode) Remove() {
	t.val.Call("remove")
}

// AppendTo appends this text node to the given parent element.
func (t TextNode) AppendTo(parent Element) {
	parent.val.Call("appendChild", t.val)
}
