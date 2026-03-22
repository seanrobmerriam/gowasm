package ssr

import (
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"

	"github.com/seanrobmerriam/gowasm/pkg/vdom"
)

var voidElements = map[string]bool{
	"area": true, "base": true, "br": true, "col": true,
	"embed": true, "hr": true, "img": true, "input": true,
	"link": true, "meta": true, "param": true, "source": true,
	"track": true, "wbr": true,
}

// renderNode writes the HTML for a single node to w.
func renderNode(node interface{}, w io.Writer, indent bool, depth int) error {
	if node == nil {
		return nil
	}

	vnode, err := callVNode(node)
	if err != nil {
		return err
	}

	switch v := vnode.(type) {
	case *vdom.Element:
		return renderElement(node, v, w, indent, depth)
	case *vdom.Text:
		if indent {
			writeIndent(w, depth)
		}
		if _, err := io.WriteString(w, escapeHTML(v.Content)); err != nil {
			return err
		}
		if indent {
			_, err := io.WriteString(w, "\n")
			return err
		}
		return nil
	case *vdom.Component:
		child, err := callRender(v.Renderer)
		if err != nil {
			return err
		}
		return renderNode(child, w, indent, depth)
	default:
		return fmt.Errorf("ssr: unsupported vnode type %T", vnode)
	}
}

func renderElement(node interface{}, vnode *vdom.Element, w io.Writer, indent bool, depth int) error {
	if indent {
		writeIndent(w, depth)
	}

	if _, err := io.WriteString(w, "<"+vnode.Tag); err != nil {
		return err
	}

	attrs := make(map[string]string, len(vnode.Attrs)+4)
	for k, v := range vnode.Attrs {
		attrs[k] = v
	}
	if vnode.ID != "" {
		attrs["id"] = vnode.ID
	}
	if len(vnode.ClassList) > 0 {
		attrs["class"] = strings.Join(vnode.ClassList, " ")
	}
	if style := serializeStyles(vnode.Styles); style != "" {
		attrs["style"] = style
	}
	if vnode.Key != "" {
		attrs["data-gowasm-key"] = vnode.Key
	}

	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if _, err := io.WriteString(w, fmt.Sprintf(` %s="%s"`, k, escapeHTML(attrs[k]))); err != nil {
			return err
		}
	}

	children, err := callSSRChildren(node)
	if err != nil {
		return err
	}

	if voidElements[vnode.Tag] {
		if _, err := io.WriteString(w, "/>"); err != nil {
			return err
		}
		if indent {
			_, err := io.WriteString(w, "\n")
			return err
		}
		return nil
	}

	if _, err := io.WriteString(w, ">"); err != nil {
		return err
	}
	if indent && len(children) > 0 {
		if _, err := io.WriteString(w, "\n"); err != nil {
			return err
		}
	}

	for _, child := range children {
		if err := renderNode(child, w, indent, depth+1); err != nil {
			return err
		}
	}

	if indent && len(children) > 0 {
		writeIndent(w, depth)
	}
	if _, err := io.WriteString(w, "</"+vnode.Tag+">"); err != nil {
		return err
	}
	if indent {
		_, err := io.WriteString(w, "\n")
		return err
	}
	return nil
}

func callVNode(node interface{}) (interface{}, error) {
	v := reflect.ValueOf(node)
	method := v.MethodByName("VNode")
	if !method.IsValid() {
		return nil, fmt.Errorf("ssr: node %T does not expose VNode()", node)
	}
	results := method.Call(nil)
	if len(results) != 1 {
		return nil, fmt.Errorf("ssr: VNode() must return exactly one value")
	}
	return results[0].Interface(), nil
}

func callSSRChildren(node interface{}) ([]interface{}, error) {
	v := reflect.ValueOf(node)
	method := v.MethodByName("SSRChildren")
	if !method.IsValid() {
		return nil, nil
	}
	results := method.Call(nil)
	if len(results) != 1 {
		return nil, fmt.Errorf("ssr: SSRChildren() must return exactly one value")
	}
	children, ok := results[0].Interface().([]interface{})
	if !ok {
		return nil, fmt.Errorf("ssr: SSRChildren() must return []interface{}")
	}
	return children, nil
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

func serializeStyles(styles map[string]string) string {
	if len(styles) == 0 {
		return ""
	}
	keys := make([]string, 0, len(styles))
	for k := range styles {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k)
		b.WriteString(": ")
		b.WriteString(styles[k])
		b.WriteString(";")
		if len(keys) > 0 {
			b.WriteString(" ")
		}
	}
	return strings.TrimSpace(b.String())
}

func writeIndent(w io.Writer, depth int) {
	for i := 0; i < depth; i++ {
		_, _ = io.WriteString(w, "  ")
	}
}
