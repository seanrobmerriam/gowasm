# gowasm

A frontend framework for Go that compiles to WebAssembly. Write components in
idiomatic Go, run them in the browser.

No JavaScript required in application code. No code generation. No separate
build pipeline beyond the standard Go toolchain.

---

## Requirements

- Go 1.21 or later
- A browser with WebAssembly support (all modern browsers)

---

## Installation

```bash
go get github.com/seanrobmerriam/gowasm
```

Install the CLI:

```bash
go install github.com/seanrobmerriam/gowasm/cmd/gowasm@latest
```

---

## Quick start

Create a new directory for your app and add a `main.go`:

```go
//go:build js && wasm

package main

import (
    "github.com/seanrobmerriam/gowasm/pkg/component"
    "github.com/seanrobmerriam/gowasm/pkg/dom"
    "github.com/seanrobmerriam/gowasm/pkg/reactive"
)

type Counter struct {
    count *reactive.Signal[int]
}

func NewCounter() *Counter {
    return &Counter{count: reactive.NewSignal(0)}
}

func (c *Counter) Render() component.Node {
    return component.H("div",
        component.Children(
            component.H("p", component.Children(
                component.Text(itoa(c.count.Get())),
            )),
            component.H("button",
                component.On("click", func(e dom.Event) {
                    c.count.Set(c.count.Peek() + 1)
                }),
                component.Children(component.Text("+")),
            ),
        ),
    )
}

func main() {
    component.Mount("app", NewCounter())
}
```

Add an `index.html`:

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>My App</title>
</head>
<body>
  <div id="app"></div>
  <script src="/wasm_exec.js"></script>
  <script>
    const go = new Go();
    WebAssembly.instantiateStreaming(fetch("/app.wasm"), go.importObject)
      .then(result => go.run(result.instance));
  </script>
  <!-- livereload -->
</body>
</html>
```

Start the dev server:

```bash
gowasm dev -dir .
```

Open `http://localhost:8080`.

---

## CLI

### `gowasm build`

Compiles your Go package to WebAssembly.

```bash
gowasm build -dir ./myapp -out app.wasm
```

| Flag | Default | Description |
|------|---------|-------------|
| `-dir` | `.` | Directory containing `main.go` |
| `-out` | `app.wasm` | Output filename |

### `gowasm serve`

Serves a pre-built app over HTTP. Sets the required COOP/COEP headers and
serves `wasm_exec.js` automatically from your local Go installation. Unknown
paths fall back to `index.html`, which is required for deep links when using
the History API router.

```bash
gowasm serve -dir ./myapp -out app.wasm -port 8080
```

| Flag | Default | Description |
|------|---------|-------------|
| `-dir` | `.` | Directory containing `index.html` |
| `-out` | `app.wasm` | Path to compiled `.wasm` file |
| `-port` | `8080` | HTTP port |
| `-host` | `localhost` | HTTP host |

### `gowasm dev`

Watches for file changes, rebuilds, and live-reloads the browser. The initial
build runs automatically. If the first build fails the watcher still starts —
fix the error and save to trigger a rebuild.

```bash
gowasm dev -dir ./myapp
```

Accepts the same flags as `serve`.

---

## Testing

The project now has two test tiers:

- `pkg/reactive` uses standard `go test` and runs without a browser
- `pkg/dom` and `pkg/component` run as real WebAssembly in headless Chrome via
  [`wasmbrowsertest`](https://github.com/agnivade/wasmbrowsertest)

Install the browser test runner once:

```bash
go install github.com/agnivade/wasmbrowsertest@latest
```

Run the full suite:

```bash
make test
```

Or run each tier separately:

```bash
make test-reactive
make test-dom
make test-component
```

`test-dom` and `test-component` require Chrome or Chromium to be installed on
the machine.

---

## Rendering Modes

gowasm now supports three rendering paths built on the same component model:

- Client mount: `component.Mount("app", App{})`
- Client hydration of SSR HTML: `hydrate.Hydrate("app", App{})`
- Server-side rendering: `ssr.New().RenderToString(App{})`

The virtual tree used by all three paths lives in `pkg/vdom`, which has no
browser-runtime dependency and can be used in standard Go server processes.

---

## Packages

### `pkg/component`

The component model. A component is any type that implements `Render() Node`.

```go
type Component interface {
    Render() Node
}
```

**Building a node tree**

```go
component.H("div",
    component.ID("main"),
    component.Class("container", "dark"),
    component.Attr("data-role", "main"),
    component.Style("padding", "1rem"),
    component.On("click", handler),
    component.Children(
        component.Text("Hello"),
        component.H("span", component.Children(component.Text("world"))),
    ),
)
```

**Lifecycle hooks**

```go
type OnMounter  interface { OnMount() }
type OnUnmounter interface { OnUnmount() }
```

Implement either or both on your component struct. `OnMount` is called after
the component's first render is attached to the DOM. `OnUnmount` is called
before it is removed.

**Mounting**

```go
// Mount blocks forever — call it as the last statement in main().
component.Mount("app", NewCounter())
```

For server-rendered pages, use `hydrate.Hydrate("app", NewCounter())`
instead. It adopts existing SSR HTML when `data-ssr="true"` is present and
falls back to a normal client mount otherwise.

**Embedding components in a tree**

```go
component.H("div", component.Children(
    component.C(NewSidebar()),
    component.C(NewContent()),
))
```

**Keyed list rendering**

Use keys for dynamic sibling lists so the reconciler can preserve identity
across insertions, deletions, and reordering.

```go
component.H("ul", component.Children(
    component.CKeyed("todo-1", NewTodoItem(todo1)),
    component.CKeyed("todo-2", NewTodoItem(todo2)),
    component.H("li",
        component.Key("summary"),
        component.Children(component.Text("2 items")),
    ),
))
```

---

### `pkg/reactive`

Fine-grained reactivity. Signals track reads and writes; effects and computed
values re-run automatically when their dependencies change.

**Signal**

```go
count := reactive.NewSignal(0)

count.Get()      // read — tracked inside effects and Render()
count.Peek()     // read — not tracked, safe inside event handlers
count.Set(count.Peek() + 1)
```

**Computed**

```go
doubled := reactive.NewComputed(func() int {
    return count.Get() * 2
})

doubled.Get() // re-evaluates only when count changes
```

**Effect**

```go
effect := reactive.NewEffect(func() {
    fmt.Println("count is now", count.Get())
})

effect.Dispose() // stop the effect and release subscriptions
```

**Batch**

Multiple signal writes inside `Batch` notify subscribers only once, after the
function returns.

```go
reactive.Batch(func() {
    firstName.Set("Ada")
    lastName.Set("Lovelace")
})
```

---

### `pkg/router`

Client-side routing with two modes:

- Hash mode: `#/path` via `router.New()`
- History API mode: `/path` via `router.New(router.WithHistoryAPI())`

**Setup**

```go
r := router.New()

r.Handle("/", homeHandler)
r.Handle("/about", aboutHandler)
r.Handle("/users/:id", userHandler)
r.Handle("/docs/*", docsHandler)

r.NotFound(func(ctx router.RouteContext) component.Node {
    return component.Text("not found")
})

component.Mount("app", r.View())
```

**History API mode**

```go
r := router.New(router.WithHistoryAPI())
```

Use this mode when you want clean URLs without `#`. `gowasm serve` and
`gowasm dev` serve `index.html` as a fallback for unknown routes so direct
navigation and refreshes continue to work.

If the page was server-rendered, pair the router with `hydrate.Hydrate(...)`
instead of `component.Mount(...)`.

**Handlers**

A handler receives a `RouteContext` and returns a `component.Node`.

```go
func userHandler(ctx router.RouteContext) component.Node {
    id := ctx.Get("id")         // named param
    tab := ctx.Get("tab")       // query string: #/users/42?tab=posts or /users/42?tab=posts
    return component.Text("user: " + id + " tab: " + tab)
}
```

**Navigation**

```go
r.Navigate("/users/42")
```

**Link**

```go
router.Link(r, router.LinkProps{
    To:       "/about",
    Children: []component.Node{component.Text("About")},
    Class:    "nav-link",
    Active:   "nav-link--active", // added when path matches
})
```

In hash mode the rendered href is `#/about`; in History API mode it is
`/about`.

**Route patterns**

| Pattern | Matches |
|---------|---------|
| `/about` | `/about` only |
| `/users/:id` | `/users/42`, `/users/ada` |
| `/docs/*` | `/docs/`, `/docs/intro/setup` |
| `/org/:org/repo/:repo` | `/org/acme/repo/core` |

When multiple patterns match, the most specific wins: static segments beat
params, params beat wildcards.

---

### `pkg/dom`

Low-level DOM bindings. Most application code should use `pkg/component`
instead. Use `pkg/dom` directly when you need fine-grained DOM control or are
writing framework-level code.

```go
el, ok := dom.ElementFromID("my-input")
if ok {
    el.SetAttr("disabled", "true")
    el.SetStyle("opacity", "0.5")
}

handle := el.AddEventListener("input", func(e dom.Event) {
    val := e.Value("target").Get("value").String()
    _ = val
})

handle.Release() // removes the listener and frees the JS function
```

---

### `pkg/vdom`

Pure virtual DOM node definitions shared by client rendering, hydration, and
SSR. This package has no browser-runtime dependency, which makes it usable in
plain Go server code.

---

### `pkg/ssr`

Server-side rendering for gowasm components. `pkg/ssr` walks the same
component tree and produces deterministic HTML strings without requiring
WebAssembly or a browser.

```go
renderer := ssr.New()
fragment, err := renderer.RenderToString(NewApp())
page := ssr.RenderPage(fragment, ssr.PageOptions{
    Title:  "My App",
    RootID: "app",
})
```

The generated page wrapper marks the app root with `data-ssr="true"` so the
client can hydrate it later.

---

### `pkg/hydrate`

Client-side hydration for SSR pages. `pkg/hydrate` claims the existing DOM,
attaches event listeners, adopts the claimed nodes into the component tree,
and then hands off to the normal reactive patch loop.

```go
hydrate.Hydrate("app", NewApp())
```

---

## Design notes

**Reactivity is automatic inside Render**

Any `Signal.Get()` call inside a component's `Render()` method automatically
subscribes that component to updates. When the signal changes, the component
re-renders and the DOM is patched in place. No manual subscription or
dependency declaration needed.

**Use Peek inside event handlers**

`Signal.Get()` records a dependency. Inside an event handler you almost always
want `Signal.Peek()` — you are writing a new value, not declaring a dependency
on the current one.

**wasm_exec.js is served automatically**

`gowasm serve` and `gowasm dev` read `wasm_exec.js` directly from your local Go
installation at runtime, so it always matches the compiler version. You do not
need to vendor it or keep it in sync manually.

**History API routes are server-backed**

When using `router.WithHistoryAPI()`, direct navigation to `/users/42` or a
refresh on that route still works because the server falls back to
`index.html` for unknown paths.

**SSR and hydration share the same tree**

The server renders `pkg/vdom` descriptions to HTML through `pkg/ssr`, and the
browser later reuses that HTML through `pkg/hydrate` instead of discarding it
and remounting from scratch.

**Binary size**

The standard Go toolchain produces binaries in the 2-5MB range for typical
apps. TinyGo support is planned and will reduce this substantially.

---

## Examples

The `examples/` directory contains:

- `examples/counter` — minimal signal and event handling
- `examples/router-demo` — multi-page routing with params and links
- `examples/router-history` — History API router with clean URLs and hydration entrypoint
- `examples/todo` — shared state, forms, filters, and keyed list rendering

Run either with:

```bash
gowasm dev -dir examples/counter
gowasm dev -dir examples/router-demo
gowasm dev -dir examples/router-history
gowasm dev -dir examples/todo
```

There is also a minimal host-side SSR demo server:

```bash
go run ./cmd/ssrserver
```

---

## License

See LICENSE.md

---

## TODO

- `gowasm new` scaffold command to generate a starter project
- Form input helpers: `InputValue(e Event) string`, `CheckboxChecked(e Event) bool`
- TinyGo build target for smaller binaries
- `go test` harness for unit testing components outside the browser
