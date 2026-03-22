package component

import (
	"github.com/seanrobmerriam/gowasm/pkg/dom"
	"github.com/seanrobmerriam/gowasm/pkg/reactive"
	"github.com/seanrobmerriam/gowasm/pkg/vdom"
)

// Node is the unit of UI. Every element in the tree is a Node.
type Node interface {
	// Mount attaches this node under parent and returns the root DOM element.
	Mount(parent dom.Element) dom.Element
	// Patch updates the existing DOM in-place given the previous node description.
	// Returns false if the node type changed and a full remount is required.
	Patch(old Node) bool
	// Unmount removes this node from the DOM and releases resources.
	Unmount()
	getKey() string
}

// Component produces a Node tree describing its current UI.
type Component interface {
	Render() Node
}

// OnMounter is implemented by components that need to run after mounting.
type OnMounter interface {
	OnMount()
}

// OnUnmounter is implemented by components that need to run before unmounting.
type OnUnmounter interface {
	OnUnmount()
}

// ElementNode represents a DOM element node.
type ElementNode struct {
	vnode     *vdom.Element
	domEl     dom.Element
	listeners []dom.ListenerHandle
	children  []Node
}

// H creates a new element node with the given tag and options.
func H(tag string, opts ...Option) *ElementNode {
	n := &ElementNode{
		vnode:     vdom.NewElement(tag),
		listeners: make([]dom.ListenerHandle, 0),
		children:  make([]Node, 0),
	}
	for _, opt := range opts {
		opt(n)
	}
	return n
}

func (n *ElementNode) getKey() string { return n.vnode.Key }

// VNode returns the underlying virtual DOM description.
// Used by the SSR renderer — not needed in normal component code.
func (n *ElementNode) VNode() *vdom.Element { return n.vnode }

// SSRChildren returns the children of this ElementNode as a slice of
// interface{} values. Used by the SSR renderer to walk the tree without
// a typed import of pkg/component.
func (n *ElementNode) SSRChildren() []interface{} {
	result := make([]interface{}, len(n.children))
	for i, child := range n.children {
		result[i] = child
	}
	return result
}

// AdoptElement and AdoptTextNode are implemented in the js/wasm-only
// hydration support file so host builds remain free of browser runtime imports.

// Mount implements Node.
func (n *ElementNode) Mount(parent dom.Element) dom.Element {
	el := dom.NewElement(n.vnode.Tag)
	n.domEl = el

	if n.vnode.ID != "" {
		el.SetAttr("id", n.vnode.ID)
	}
	if len(n.vnode.ClassList) > 0 {
		el.SetAttr("class", joinStrings(n.vnode.ClassList, " "))
	}
	for k, v := range n.vnode.Attrs {
		el.SetAttr(k, v)
	}
	for k, v := range n.vnode.Styles {
		el.SetStyle(k, v)
	}
	for event, handler := range n.vnode.Events {
		if h, ok := handler.(dom.EventHandler); ok {
			handle := el.AddEventListener(event, h)
			n.listeners = append(n.listeners, handle)
		}
	}
	for _, child := range n.children {
		child.Mount(el)
	}
	parent.AppendChild(el)
	return el
}

// Patch implements Node.
func (n *ElementNode) Patch(old Node) bool {
	oldEl, ok := old.(*ElementNode)
	if !ok || oldEl.vnode.Tag != n.vnode.Tag {
		return false
	}

	n.domEl = oldEl.domEl

	if n.vnode.ID != oldEl.vnode.ID {
		if n.vnode.ID != "" {
			n.domEl.SetAttr("id", n.vnode.ID)
		} else {
			n.domEl.RemoveAttr("id")
		}
	}

	if len(n.vnode.ClassList) > 0 || len(oldEl.vnode.ClassList) > 0 {
		n.domEl.SetAttr("class", joinStrings(n.vnode.ClassList, " "))
	}

	for k := range oldEl.vnode.Attrs {
		if _, exists := n.vnode.Attrs[k]; !exists {
			n.domEl.RemoveAttr(k)
		}
	}
	for k, v := range n.vnode.Attrs {
		if oldEl.vnode.Attrs[k] != v {
			n.domEl.SetAttr(k, v)
		}
	}

	for k := range oldEl.vnode.Styles {
		if _, exists := n.vnode.Styles[k]; !exists {
			n.domEl.SetStyle(k, "")
		}
	}
	for k, v := range n.vnode.Styles {
		if oldEl.vnode.Styles[k] != v {
			n.domEl.SetStyle(k, v)
		}
	}

	for _, handle := range oldEl.listeners {
		handle.Release()
	}
	oldEl.listeners = nil
	n.listeners = nil

	for event, handler := range n.vnode.Events {
		if h, ok := handler.(dom.EventHandler); ok {
			handle := n.domEl.AddEventListener(event, h)
			n.listeners = append(n.listeners, handle)
		}
	}

	n.patchChildren(oldEl.children)

	return true
}

// patchChildren reconciles the children array.
func (n *ElementNode) patchChildren(oldChildren []Node) {
	oldKeyed := make(map[string]Node)
	oldUnkeyed := make([]Node, 0, len(oldChildren))

	for _, child := range oldChildren {
		if k := child.getKey(); k != "" {
			oldKeyed[k] = child
		} else {
			oldUnkeyed = append(oldUnkeyed, child)
		}
	}

	consumed := make(map[string]bool)
	unkeyedIdx := 0

	for _, newChild := range n.children {
		k := newChild.getKey()

		if k != "" {
			if oldChild, ok := oldKeyed[k]; ok {
				consumed[k] = true
				if oldEl, ok2 := oldChild.(*ElementNode); ok2 {
					if newEl, ok3 := newChild.(*ElementNode); ok3 {
						newEl.domEl = oldEl.domEl
					}
				}
				if !newChild.Patch(oldChild) {
					oldChild.Unmount()
					newChild.Mount(n.domEl)
				}
			} else {
				newChild.Mount(n.domEl)
			}
		} else {
			if unkeyedIdx < len(oldUnkeyed) {
				oldChild := oldUnkeyed[unkeyedIdx]
				unkeyedIdx++

				if oldEl, ok := oldChild.(*ElementNode); ok {
					if newEl, ok2 := newChild.(*ElementNode); ok2 {
						newEl.domEl = oldEl.domEl
					}
				}
				if oldTxt, ok := oldChild.(*TextNode); ok {
					if newTxt, ok2 := newChild.(*TextNode); ok2 {
						newTxt.domEl = oldTxt.domEl
					}
				}

				if !newChild.Patch(oldChild) {
					oldChild.Unmount()
					newChild.Mount(n.domEl)
				}
			} else {
				newChild.Mount(n.domEl)
			}
		}
	}

	for i := unkeyedIdx; i < len(oldUnkeyed); i++ {
		oldUnkeyed[i].Unmount()
	}

	for k, child := range oldKeyed {
		if !consumed[k] {
			child.Unmount()
		}
	}
}

// Unmount implements Node.
func (n *ElementNode) Unmount() {
	for _, handle := range n.listeners {
		handle.Release()
	}
	n.listeners = nil

	for _, child := range n.children {
		child.Unmount()
	}

	n.domEl.Remove()
}

// TextNode represents a text content node.
type TextNode struct {
	vnode *vdom.Text
	domEl dom.TextNode
}

// Text creates a new text node.
func Text(s string) *TextNode {
	return &TextNode{vnode: vdom.NewText(s)}
}

func (n *TextNode) getKey() string { return "" }

// VNode returns the underlying virtual DOM description.
// Used by the SSR renderer — not needed in normal component code.
func (n *TextNode) VNode() *vdom.Text { return n.vnode }

// Mount implements Node.
func (n *TextNode) Mount(parent dom.Element) dom.Element {
	textNode := dom.Doc().CreateTextNode(n.vnode.Content)
	n.domEl = textNode
	textNode.AppendTo(parent)
	return parent
}

// Patch implements Node.
func (n *TextNode) Patch(old Node) bool {
	oldText, ok := old.(*TextNode)
	if !ok {
		return false
	}
	if n.vnode.Content != oldText.vnode.Content {
		n.domEl.SetText(n.vnode.Content)
	}
	return true
}

// Unmount implements Node.
func (n *TextNode) Unmount() {
	n.domEl.Remove()
}

// ComponentNode wraps a Component so it can appear as a Node in a tree.
type ComponentNode struct {
	vnode    *vdom.Component
	rootNode Node
	effect   *reactive.Effect
	mounted  bool
}

// C creates a new component node.
func C(component Component) *ComponentNode {
	return &ComponentNode{
		vnode: vdom.NewComponent("", component),
	}
}

// CKeyed creates a component node with a reconciliation key.
// Use when rendering dynamic lists of components.
func CKeyed(key string, component Component) *ComponentNode {
	return &ComponentNode{
		vnode: vdom.NewComponent(key, component),
	}
}

func (cn *ComponentNode) getKey() string { return cn.vnode.Key }

// VNode returns the underlying virtual DOM description.
// Used by the SSR renderer — not needed in normal component code.
func (cn *ComponentNode) VNode() *vdom.Component { return cn.vnode }

func (cn *ComponentNode) component() Component {
	return cn.vnode.Renderer.(Component)
}

// Mount implements Node.
func (cn *ComponentNode) Mount(parent dom.Element) dom.Element {
	var firstRun = true

	cn.effect = reactive.NewEffect(func() {
		newRoot := cn.component().Render()

		if firstRun {
			firstRun = false
			cn.rootNode = newRoot
			cn.rootNode.Mount(parent)

			if m, ok := cn.component().(OnMounter); ok {
				m.OnMount()
			}
		} else {
			oldRoot := cn.rootNode
			cn.rootNode = newRoot
			if !cn.rootNode.Patch(oldRoot) {
				oldRoot.Unmount()
				cn.rootNode.Mount(parent)
			}
		}
	})

	cn.mounted = true
	return parent
}

// Patch implements Node.
func (cn *ComponentNode) Patch(old Node) bool {
	return false
}

// Unmount implements Node.
func (cn *ComponentNode) Unmount() {
	if !cn.mounted {
		return
	}

	if u, ok := cn.component().(OnUnmounter); ok {
		u.OnUnmount()
	}

	if cn.effect != nil {
		cn.effect.Dispose()
		cn.effect = nil
	}

	if cn.rootNode != nil {
		cn.rootNode.Unmount()
	}

	cn.mounted = false
}

// joinStrings joins a slice of strings with a separator.
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
