//go:build js && wasm

package main

import (
	"github.com/seanrobmerriam/gowasm/pkg/component"
	"github.com/seanrobmerriam/gowasm/pkg/router"
)

func main() {
	r := router.New(router.WithHistoryAPI())

	r.Handle("/", func(ctx router.RouteContext) component.Node {
		return component.H("div", component.Children(
			component.H("h1", component.Children(component.Text("Home"))),
			router.Link(r, router.LinkProps{
				To:       "/about",
				Children: []component.Node{component.Text("About")},
				Active:   "active",
			}),
			component.Text(" | "),
			router.Link(r, router.LinkProps{
				To:       "/users/42",
				Children: []component.Node{component.Text("User 42")},
				Active:   "active",
			}),
		))
	})

	r.Handle("/about", func(ctx router.RouteContext) component.Node {
		return component.H("div", component.Children(
			component.H("h1", component.Children(component.Text("About"))),
			router.Link(r, router.LinkProps{
				To:       "/",
				Children: []component.Node{component.Text("← Home")},
			}),
		))
	})

	r.Handle("/users/:id", func(ctx router.RouteContext) component.Node {
		return component.H("div", component.Children(
			component.H("h1", component.Children(
				component.Text("User: "+ctx.Get("id")),
			)),
			router.Link(r, router.LinkProps{
				To:       "/",
				Children: []component.Node{component.Text("← Home")},
			}),
		))
	})

	component.Mount("app", r.View())
}
