package vdom

// Element describes a DOM element: tag, attributes, styles,
// event listeners, and children.
type Element struct {
	Tag       string
	Key       string
	ID        string
	ClassList []string
	Attrs     map[string]string
	Styles    map[string]string
	Events    map[string]interface{} // handler type defined by pkg/component
	Children  []Node
}

func (e *Element) nodeType() nodeKind { return KindElement }

// NewElement creates a new virtual element node.
func NewElement(tag string) *Element {
	return &Element{
		Tag:    tag,
		Attrs:  make(map[string]string),
		Styles: make(map[string]string),
		Events: make(map[string]interface{}),
	}
}
