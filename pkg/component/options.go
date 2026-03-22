package component

import "github.com/seanrobmerriam/gowasm/pkg/dom"

// Option is a functional option for configuring an ElementNode.
type Option func(*ElementNode)

// Attr sets an HTML attribute on an element node.
func Attr(key, val string) Option {
	return func(n *ElementNode) {
		n.vnode.Attrs[key] = val
	}
}

// Style sets a CSS style property on an element node.
func Style(prop, val string) Option {
	return func(n *ElementNode) {
		n.vnode.Styles[prop] = val
	}
}

// On registers an event handler on an element node.
func On(event string, h dom.EventHandler) Option {
	return func(n *ElementNode) {
		n.vnode.Events[event] = h
	}
}

// Children sets the child nodes of an element node.
func Children(nodes ...Node) Option {
	return func(n *ElementNode) {
		n.children = append(n.children, nodes...)
		for _, child := range nodes {
			switch c := child.(type) {
			case *ElementNode:
				n.vnode.Children = append(n.vnode.Children, c.VNode())
			case *TextNode:
				n.vnode.Children = append(n.vnode.Children, c.VNode())
			case *ComponentNode:
				n.vnode.Children = append(n.vnode.Children, c.VNode())
			}
		}
	}
}

// Class adds CSS class names to an element node.
func Class(names ...string) Option {
	return func(n *ElementNode) {
		n.vnode.ClassList = append(n.vnode.ClassList, names...)
	}
}

// ID sets the id attribute of an element node.
func ID(id string) Option {
	return func(n *ElementNode) {
		n.vnode.ID = id
	}
}

// Key sets a stable reconciliation key on an element node.
// Use keys when rendering lists of dynamic children to help the reconciler
// match old and new nodes by identity rather than position.
// Keys must be unique among siblings.
func Key(k string) Option {
	return func(n *ElementNode) {
		n.vnode.Key = k
	}
}
