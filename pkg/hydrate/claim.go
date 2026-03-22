//go:build js && wasm

package hydrate

import (
	"strconv"
	"strings"
	"syscall/js"

	"github.com/seanrobmerriam/gowasm/pkg/dom"
)

// claimMap maps a position path string to a live DOM js.Value.
// Position path format: "0", "0.1", "0.1.2" etc — child indices
// from the root, dot-separated.
type claimMap map[string]js.Value

// buildClaimMap walks the DOM tree rooted at el and records every
// element and text node by its position path.
func buildClaimMap(el dom.Element) claimMap {
	claimed := make(claimMap)
	root := el.JSValue()
	claimed[""] = root

	var walk func(node js.Value, path []int)
	walk = func(node js.Value, path []int) {
		childNodes := node.Get("childNodes")
		length := childNodes.Get("length").Int()
		for i := 0; i < length; i++ {
			child := childNodes.Index(i)
			nodeType := child.Get("nodeType").Int()
			childPath := append(append([]int(nil), path...), i)
			switch nodeType {
			case 1, 3:
				claimed[positionPath(childPath)] = child
			}
			if nodeType == 1 {
				walk(child, childPath)
			}
		}
	}

	walk(root, nil)
	return claimed
}

// positionPath builds a dot-separated path string from a slice of indices.
// e.g. []int{0, 1, 2} → "0.1.2"
func positionPath(indices []int) string {
	if len(indices) == 0 {
		return ""
	}
	parts := make([]string, len(indices))
	for i, idx := range indices {
		parts[i] = strconv.Itoa(idx)
	}
	return strings.Join(parts, ".")
}
