package vdom

// Component describes a component node in the virtual tree.
// The Renderer field holds the actual component.Component interface value —
// typed as interface{} to avoid a circular import with pkg/component.
type Component struct {
	Key      string
	Renderer interface{} // holds a component.Component
}

func (c *Component) nodeType() nodeKind { return KindComponent }

// NewComponent creates a new virtual component node.
func NewComponent(key string, renderer interface{}) *Component {
	return &Component{Key: key, Renderer: renderer}
}
