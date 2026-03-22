//go:build js && wasm

package main

import (
	"github.com/seanrobmerriam/gowasm/pkg/component"
	"github.com/seanrobmerriam/gowasm/pkg/router"
)

func main() {
	r := router.New()

	r.Handle("/", func(ctx router.RouteContext) component.Node {
		return component.H("div", component.Children(
			component.H("h1", component.Children(component.Text("Home"))),
			router.Link(r, router.LinkProps{
				To:       "/about",
				Children: []component.Node{component.Text("About")},
			}),
			component.Text(" | "),
			router.Link(r, router.LinkProps{
				To:       "/users/42",
				Children: []component.Node{component.Text("User 42")},
			}),
		))
	})

	r.Handle("/about", func(ctx router.RouteContext) component.Node {
		return component.H("div", component.Children(
			component.H("h1", component.Children(component.Text("About"))),
			component.Text("This is the about page."),
			component.H("br", nil),
			router.Link(r, router.LinkProps{
				To:       "/",
				Children: []component.Node{component.Text("← Home")},
			}),
		))
	})

	r.Handle("/users/:id", func(ctx router.RouteContext) component.Node {
		return component.H("div", component.Children(
			component.H("h1", component.Children(component.Text("User: "+ctx.Get("id")))),
			router.Link(r, router.LinkProps{
				To:       "/",
				Children: []component.Node{component.Text("← Home")},
			}),
		))
	})

	component.Mount("app", r.View())
}
