// Package ssr provides server-side rendering for gowasm components.
// It has no dependency on the browser runtime and runs in standard Go server processes.
package ssr

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
)

// Renderer walks a component tree and produces HTML.
type Renderer struct {
	// Indent controls whether output HTML is indented for readability.
	// Default false — production output is minified.
	Indent bool
}

// New creates a new Renderer with default settings.
func New() *Renderer {
	return &Renderer{}
}

// RenderToString renders a component to an HTML string.
// The returned string is a fragment — it does not include <html>, <head>,
// or <body> tags. Wrap it in a full page template as needed.
func (r *Renderer) RenderToString(component interface{}) (string, error) {
	var buf bytes.Buffer
	if err := r.RenderToWriter(component, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderToWriter renders a component and writes the HTML to w.
func (r *Renderer) RenderToWriter(component interface{}, w io.Writer) error {
	node, err := callRender(component)
	if err != nil {
		return err
	}
	return renderNode(node, w, r.Indent, 0)
}

func callRender(component interface{}) (interface{}, error) {
	v := reflect.ValueOf(component)
	method := v.MethodByName("Render")
	if !method.IsValid() {
		return nil, fmt.Errorf("ssr: component %T does not implement Render()", component)
	}
	results := method.Call(nil)
	if len(results) != 1 {
		return nil, fmt.Errorf("ssr: Render() must return exactly one value")
	}
	return results[0].Interface(), nil
}
