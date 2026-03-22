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
serves `wasm_exec.js` automatically from your local Go installation.

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

**Embedding components in a tree**

```go
component.H("div", component.Children(
    component.C(NewSidebar()),
    component.C(NewContent()),
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

Hash-based client-side routing (`#/path`).

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

**Handlers**

A handler receives a `RouteContext` and returns a `component.Node`.

```go
func userHandler(ctx router.RouteContext) component.Node {
    id := ctx.Get("id")         // named param
    tab := ctx.Get("tab")       // query string: #/users/42?tab=posts
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

**Binary size**

The standard Go toolchain produces binaries in the 2-5MB range for typical
apps. TinyGo support is planned and will reduce this substantially.

---

## Examples

The `examples/` directory contains:

- `examples/counter` — minimal signal and event handling
- `examples/router-demo` — multi-page routing with params and links

Run either with:

```bash
gowasm dev -dir examples/counter
gowasm dev -dir examples/router-demo
```

---

## License

See LICENSE.md

---

## TODO

- `gowasm new` scaffold command to generate a starter project
- Form input helpers: `InputValue(e Event) string`, `CheckboxChecked(e Event) bool`
- Todo list example covering lists, conditionals, and form inputs
- TinyGo build target for smaller binaries
- Keyed list reconciliation to preserve element identity across re-renders
- History API router (`/path` instead of `#/path`) as an alternative to hash routing
- Server-side rendering and client hydration
- `go test` harness for unit testing components outside the browser
