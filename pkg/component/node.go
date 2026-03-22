package component

import (
	"github.com/yourname/gowasm/pkg/dom"
	"github.com/yourname/gowasm/pkg/reactive"
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
	tag       string
	attrs     map[string]string
	styles    map[string]string
	events    map[string]dom.EventHandler
	listeners []dom.ListenerHandle
	children  []Node
	classList []string
	id        string
	domEl     dom.Element
}

// H creates a new element node with the given tag and options.
func H(tag string, opts ...Option) *ElementNode {
	n := &ElementNode{
		tag:       tag,
		attrs:     make(map[string]string),
		styles:    make(map[string]string),
		events:    make(map[string]dom.EventHandler),
		listeners: make([]dom.ListenerHandle, 0),
		children:  make([]Node, 0),
		classList: make([]string, 0),
	}
	for _, opt := range opts {
		opt(n)
	}
	return n
}

// Mount implements Node.
func (n *ElementNode) Mount(parent dom.Element) dom.Element {
	el := dom.NewElement(n.tag)
	n.domEl = el

	// Apply ID
	if n.id != "" {
		el.SetAttr("id", n.id)
	}

	// Apply classes
	if len(n.classList) > 0 {
		el.SetAttr("class", joinStrings(n.classList, " "))
	}

	// Apply attributes
	for k, v := range n.attrs {
		el.SetAttr(k, v)
	}

	// Apply styles
	for k, v := range n.styles {
		el.SetStyle(k, v)
	}

	// Apply event listeners
	for event, handler := range n.events {
		handle := el.AddEventListener(event, handler)
		n.listeners = append(n.listeners, handle)
	}

	// Mount children
	for _, child := range n.children {
		child.Mount(el)
	}

	// Append to parent
	parent.AppendChild(el)
	return el
}

// Patch implements Node.
func (n *ElementNode) Patch(old Node) bool {
	oldEl, ok := old.(*ElementNode)
	if !ok || oldEl.tag != n.tag {
		return false
	}

	// Update ID
	if n.id != oldEl.id {
		if n.id != "" {
			n.domEl.SetAttr("id", n.id)
		} else {
			n.domEl.RemoveAttr("id")
		}
	}

	// Update classes
	if len(n.classList) > 0 || len(oldEl.classList) > 0 {
		n.domEl.SetAttr("class", joinStrings(n.classList, " "))
	}

	// Update attributes
	for k := range oldEl.attrs {
		if _, exists := n.attrs[k]; !exists {
			n.domEl.RemoveAttr(k)
		}
	}
	for k, v := range n.attrs {
		if oldEl.attrs[k] != v {
			n.domEl.SetAttr(k, v)
		}
	}

	// Update styles
	for k := range oldEl.styles {
		if _, exists := n.styles[k]; !exists {
			n.domEl.SetStyle(k, "")
		}
	}
	for k, v := range n.styles {
		if oldEl.styles[k] != v {
			n.domEl.SetStyle(k, v)
		}
	}

	// Update event listeners - simplified approach
	for event, handler := range n.events {
		handle := n.domEl.AddEventListener(event, handler)
		n.listeners = append(n.listeners, handle)
	}

	// Reconcile children
	n.patchChildren(oldEl.children)

	return true
}

// patchChildren reconciles the children array.
func (n *ElementNode) patchChildren(oldChildren []Node) {
	// Remove old children that are no longer present
	for i := len(oldChildren) - 1; i >= 0; i-- {
		oldChildren[i].Unmount()
	}

	// Mount new children
	for _, child := range n.children {
		child.Mount(n.domEl)
	}
}

// Unmount implements Node.
func (n *ElementNode) Unmount() {
	// Remove event listeners
	for _, handle := range n.listeners {
		handle.Release()
	}
	n.listeners = nil

	// Unmount children
	for _, child := range n.children {
		child.Unmount()
	}

	// Remove from DOM
	n.domEl.Remove()
}

// TextNode represents a text content node.
type TextNode struct {
	text  string
	domEl dom.TextNode
}

// Text creates a new text node.
func Text(s string) *TextNode {
	return &TextNode{text: s}
}

// Mount implements Node.
func (n *TextNode) Mount(parent dom.Element) dom.Element {
	textNode := dom.Doc().CreateTextNode(n.text)
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
	if n.text != oldText.text {
		n.domEl.SetText(n.text)
	}
	return true
}

// Unmount implements Node.
func (n *TextNode) Unmount() {
	n.domEl.Remove()
}

// ComponentNode wraps a Component so it can appear as a Node in a tree.
type ComponentNode struct {
	component Component
	rootNode  Node
	effect    *reactive.Effect
	mounted   bool
}

// C creates a new component node.
func C(component Component) *ComponentNode {
	return &ComponentNode{
		component: component,
	}
}

// Mount implements Node.
func (cn *ComponentNode) Mount(parent dom.Element) dom.Element {
	// Call Render to get initial node tree
	cn.rootNode = cn.component.Render()
	cn.rootNode.Mount(parent)

	// Call OnMount if present
	if m, ok := cn.component.(OnMounter); ok {
		m.OnMount()
	}

	// Create effect to re-render when signals change
	cn.effect = reactive.NewEffect(func() {
		// Re-render the component
		oldRoot := cn.rootNode
		cn.rootNode = cn.component.Render()

		// Patch the old root
		if !cn.rootNode.Patch(oldRoot) {
			// Node type changed, do full remount
			oldRoot.Unmount()
			cn.rootNode.Mount(parent)
		}
	})

	cn.mounted = true
	return parent
}

// Patch implements Node.
func (cn *ComponentNode) Patch(old Node) bool {
	// Components cannot be patched in place
	return false
}

// Unmount implements Node.
func (cn *ComponentNode) Unmount() {
	if !cn.mounted {
		return
	}

	// Call OnUnmount if present
	if u, ok := cn.component.(OnUnmounter); ok {
		u.OnUnmount()
	}

	// Dispose the effect
	if cn.effect != nil {
		cn.effect.Dispose()
		cn.effect = nil
	}

	// Unmount the root node
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
