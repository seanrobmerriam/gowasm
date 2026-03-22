package router

import (
	"github.com/seanrobmerriam/gowasm/pkg/component"
	"github.com/seanrobmerriam/gowasm/pkg/dom"
)

// LinkProps configures a Link.
type LinkProps struct {
	To       string           // target path, e.g. "/about"
	Children []component.Node // inner content
	Class    string           // optional CSS class
	Active   string           // class to add when this link is the current route
}

// Link returns a component.Node rendering an <a> tag that navigates on click
// without a full page reload.
func Link(r *Router, props LinkProps) component.Node {
	classes := props.Class

	// Append active class if current path matches
	if props.Active != "" && r.Current().Path == props.To {
		if classes != "" {
			classes += " "
		}
		classes += props.Active
	}

	href := "#" + props.To
	if r.useHistoryAPI {
		href = props.To
	}

	opts := []component.Option{
		component.Attr("href", href),
		component.On("click", func(e dom.Event) {
			e.PreventDefault()
			r.Navigate(props.To)
		}),
	}
	if classes != "" {
		opts = append(opts, component.Attr("class", classes))
	}
	opts = append(opts, component.Children(props.Children...))

	return component.H("a", opts...)
}

// Navigate navigates to path. Package-level convenience wrapper.
// Requires a router instance — prefer r.Navigate(path) directly.
func Navigate(r *Router, path string) {
	r.Navigate(path)
}
