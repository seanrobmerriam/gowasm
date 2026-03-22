//go:build js && wasm

package hydrate

import (
	"github.com/seanrobmerriam/gowasm/pkg/component"
	"github.com/seanrobmerriam/gowasm/pkg/dom"
)

// attachEvents walks the node tree produced by comp.Render() and registers
// event listeners on the corresponding claimed DOM elements.
// It does not modify DOM structure — only adds listeners.
func attachEvents(
	node component.Node,
	claimed claimMap,
	path []int,
	handles *[]dom.ListenerHandle,
) {
	switch n := node.(type) {
	case *component.ElementNode:
		vnode := n.VNode()
		if claimedNode, ok := claimed[positionPath(path)]; ok {
			el := dom.ElementFromJSValue(claimedNode)
			for event, handler := range vnode.Events {
				if h, ok := handler.(dom.EventHandler); ok {
					handle := el.AddEventListener(event, h)
					*handles = append(*handles, handle)
				}
			}
		}

		children := n.SSRChildren()
		for i, child := range children {
			if c, ok := child.(component.Node); ok {
				attachEvents(c, claimed, append(append([]int(nil), path...), i), handles)
			}
		}
	case *component.ComponentNode:
		rendered := n.VNode().Renderer.(component.Component).Render()
		attachEvents(rendered, claimed, path, handles)
	case *component.TextNode:
		return
	}
}
