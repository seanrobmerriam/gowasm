package router

import (
	"strings"

	"github.com/seanrobmerriam/gowasm/pkg/component"
	"github.com/seanrobmerriam/gowasm/pkg/reactive"
)

// Handler is a function that returns a Node given the matched route context.
type Handler func(ctx RouteContext) component.Node

// RouterOption configures a Router.
type RouterOption func(*Router)

// WithHistoryAPI switches the router from hash-based to pushState-based
// navigation. The server must serve index.html for all routes.
func WithHistoryAPI() RouterOption {
	return func(r *Router) {
		r.useHistoryAPI = true
	}
}

// route is a single registered route.
type route struct {
	pattern     string
	segments    []segment
	handler     Handler
	specificity int // higher = more specific
}

// segment is one path segment, classified during registration.
type segment struct {
	kind  segmentKind // static | param | wildcard
	value string      // literal text or param name
}

type segmentKind int

const (
	segStatic   segmentKind = iota
	segParam                // :name
	segWildcard             // *
)

// Router matches URL paths to handler functions.
type Router struct {
	routes        []route
	notFound      Handler
	signal        *reactive.Signal[RouteContext]
	current       RouteContext
	useHistoryAPI bool
}

// New creates a new Router and installs the hashchange listener.
// New must be called once per application.
func New(opts ...RouterOption) *Router {
	r := &Router{
		routes: []route{},
		notFound: func(ctx RouteContext) component.Node {
			return component.H("div", component.Children(
				component.H("h1", component.Children(component.Text("404 Not Found"))),
				component.H("p", component.Children(component.Text("Path: "+ctx.Path))),
			))
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(r)
	}

	// Initialise the appropriate history backend
	if r.useHistoryAPI {
		histAPI.init()
	} else {
		history.init()
	}

	// Build initial context with query parsing
	var initialPath string
	if r.useHistoryAPI {
		initialPath = histAPI.current
	} else {
		initialPath = history.current
	}
	initialCtx := RouteContext{Path: initialPath}
	if idx := strings.Index(initialPath, "?"); idx >= 0 {
		initialCtx.Path = initialPath[:idx]
		initialCtx.Query = parseQuery(initialPath[idx+1:])
	} else {
		initialCtx.Query = make(map[string]string)
	}
	r.current = initialCtx
	r.signal = reactive.NewSignal[RouteContext](initialCtx)

	// Register change listener on the appropriate backend
	changeHandler := func(path string) {
		ctx := RouteContext{Path: path}
		if idx := strings.Index(path, "?"); idx >= 0 {
			ctx.Path = path[:idx]
			ctx.Query = parseQuery(path[idx+1:])
		} else {
			ctx.Query = make(map[string]string)
		}
		r.current = ctx
		r.signal.Set(ctx)
	}

	if r.useHistoryAPI {
		histAPI.onchange(changeHandler)
	} else {
		history.onchange(changeHandler)
	}

	return r
}

// Handle registers pattern to handler.
// Patterns:
//
//	/static/path
//	/users/:id
//	/docs/*
//	/org/:org/repo/:repo
func (r *Router) Handle(pattern string, handler Handler) *Router {
	segments := parsePattern(pattern)
	specificity := computeSpecificity(segments)
	r.routes = append(r.routes, route{
		pattern:     pattern,
		segments:    segments,
		handler:     handler,
		specificity: specificity,
	})
	return r
}

// NotFound sets the handler used when no route matches.
func (r *Router) NotFound(handler Handler) *Router {
	r.notFound = handler
	return r
}

// View returns a Component that renders the currently matched route.
func (r *Router) View() *routerView {
	return &routerView{router: r}
}

// Navigate navigates to the given path, e.g. "/users/42".
func (r *Router) Navigate(path string) {
	if r.useHistoryAPI {
		histAPI.navigate(path)
	} else {
		history.navigate(path)
	}
}

// Current returns the current RouteContext.
func (r *Router) Current() RouteContext {
	return r.current
}

// match finds the best matching route for the given path.
// Returns the matched route and updated context with params.
func (r *Router) match(path string) (*route, RouteContext) {
	var bestMatch *route
	bestSpecificity := -1
	var bestParams map[string]string

	// Extract query params from path
	ctx := RouteContext{Path: path}
	if idx := strings.Index(path, "?"); idx >= 0 {
		ctx.Path = path[:idx]
		ctx.Query = parseQuery(path[idx+1:])
	}

	for i := range r.routes {
		route := &r.routes[i]
		params, ok := matchSegments(route.segments, ctx.Path)
		if !ok {
			continue
		}
		// Prefer higher specificity
		if route.specificity > bestSpecificity {
			bestSpecificity = route.specificity
			bestMatch = route
			bestParams = params
		}
	}

	if bestMatch != nil {
		ctx.Params = bestParams
		ctx.Pattern = bestMatch.pattern
		return bestMatch, ctx
	}
	return nil, ctx
}

// parsePattern parses a route pattern into segments.
func parsePattern(pattern string) []segment {
	if pattern == "" || pattern[0] != '/' {
		pattern = "/" + pattern
	}

	var segments []segment
	parts := strings.Split(pattern, "/")

	for i := 1; i < len(parts); i++ { // skip empty first part
		part := parts[i]
		if part == "" {
			continue
		}
		if part == "*" {
			segments = append(segments, segment{kind: segWildcard, value: "*"})
		} else if part != "" && part[0] == ':' {
			segments = append(segments, segment{kind: segParam, value: part[1:]})
		} else {
			segments = append(segments, segment{kind: segStatic, value: part})
		}
	}

	return segments
}

// computeSpecificity calculates route specificity (higher = more specific).
func computeSpecificity(segments []segment) int {
	spec := 0
	for _, seg := range segments {
		switch seg.kind {
		case segStatic:
			spec += 100
		case segParam:
			spec += 10
		case segWildcard:
			spec += 1
		}
	}
	return spec
}

// matchSegments attempts to match segments against a path.
// Returns params map if successful.
func matchSegments(segments []segment, path string) (map[string]string, bool) {
	// Strip leading slash
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	// Split into parts; empty path means root — zero parts
	var parts []string
	if path != "" {
		parts = strings.Split(path, "/")
	}

	params := make(map[string]string)

	// Root route: no segments, no parts → match
	if len(segments) == 0 && len(parts) == 0 {
		return params, true
	}

	segIdx := 0
	pathIdx := 0

	for pathIdx < len(parts) {
		if segIdx >= len(segments) {
			// More path segments than route segments
			return nil, false
		}

		seg := segments[segIdx]
		part := parts[pathIdx]

		switch seg.kind {
		case segStatic:
			if seg.value != part {
				return nil, false
			}
		case segParam:
			params[seg.value] = part
		case segWildcard:
			// Consume remaining path
			params["*"] = strings.Join(parts[pathIdx:], "/")
			return params, true
		}

		segIdx++
		pathIdx++
	}

	// Check if we consumed all segments
	if segIdx < len(segments) {
		// Route has more segments than path - check if remaining are optional/wildcard
		for ; segIdx < len(segments); segIdx++ {
			if segments[segIdx].kind != segWildcard {
				return nil, false
			}
		}
	}

	return params, true
}

// routerView is defined in view.go
