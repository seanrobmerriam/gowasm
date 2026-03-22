package component

import "github.com/seanrobmerriam/gowasm/pkg/dom"

// Reconcile updates the DOM by comparing old and new node trees.
// It returns the updated root node.
func Reconcile(old Node, new Node, parent dom.Element) Node {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		new.Mount(parent)
		return new
	}
	if new == nil {
		old.Unmount()
		return nil
	}
	if !new.Patch(old) {
		old.Unmount()
		new.Mount(parent)
	}
	return new
}
