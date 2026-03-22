// Package vdom defines the virtual DOM node tree used by gowasm.
// It has no browser-runtime dependency and can be used in both browser
// and server environments.
package vdom

// Node is a virtual DOM node — a description of UI, not a live DOM element.
type Node interface {
	nodeType() nodeKind
}

type nodeKind int

const (
	KindElement nodeKind = iota
	KindText
	KindComponent
)
