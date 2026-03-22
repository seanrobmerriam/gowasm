package router

import "github.com/yourname/gowasm/pkg/component"

// routerView is the internal component returned by Router.View().
type routerView struct {
	router *Router
}

// Render implements component.Component.
// Reads the router signal — this is what causes re-renders on navigation.
func (rv *routerView) Render() component.Node {
	ctx := rv.router.signal.Get() // tracked by reactive system
	matched, ctx := rv.router.match(ctx.Path)
	if matched == nil {
		return rv.router.notFound(ctx)
	}
	return matched.handler(ctx)
}
