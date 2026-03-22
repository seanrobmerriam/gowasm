//go:build js && wasm

package hydrate

import (
	"strings"
	"syscall/js"

	"github.com/seanrobmerriam/gowasm/pkg/component"
)

// adoptDOM walks the component node tree and injects claimed DOM references
// into ElementNode and TextNode instances so their domEl fields point at the
// existing server-rendered nodes rather than being zero values.
func adoptDOM(
	node component.Node,
	claimed claimMap,
	path []int,
) {
	switch n := node.(type) {
	case *component.ElementNode:
		if claimedNode, ok := claimed[positionPath(path)]; ok {
			got := strings.ToLower(claimedNode.Get("tagName").String())
			expected := n.VNode().Tag
			if got != expected {
				warnMismatch(positionPath(path), expected, got)
				return
			}
			n.AdoptElement(claimedNode)
		}

		children := n.SSRChildren()
		for i, child := range children {
			if c, ok := child.(component.Node); ok {
				adoptDOM(c, claimed, append(append([]int(nil), path...), i))
			}
		}
	case *component.ComponentNode:
		rendered := n.VNode().Renderer.(component.Component).Render()
		adoptDOM(rendered, claimed, path)
	case *component.TextNode:
		if claimedNode, ok := claimed[positionPath(path)]; ok {
			n.AdoptTextNode(claimedNode)
		}
	}
}

func warnMismatch(path string, expected, got string) {
	js.Global().Get("console").Call("warn",
		"gowasm hydrate: mismatch at "+path+
			" expected <"+expected+"> got <"+got+">")
}
