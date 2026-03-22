package vdom

// Text describes a DOM text node.
type Text struct {
	Content string
}

func (t *Text) nodeType() nodeKind { return KindText }

// NewText creates a new virtual text node.
func NewText(content string) *Text {
	return &Text{Content: content}
}
