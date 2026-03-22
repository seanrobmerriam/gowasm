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
		component.Class("counter"),
		component.Children(
			component.H("h1", component.Children(
				component.Text("gowasm counter"),
			)),
			component.H("p", component.Children(
				component.Text("Count: "),
				component.Text(itoa(c.count.Get())),
			)),
			component.H("button",
				component.On("click", func(e dom.Event) {
					e.PreventDefault()
					c.count.Set(c.count.Peek() - 1)
				}),
				component.Children(component.Text("-")),
			),
			component.H("button",
				component.On("click", func(e dom.Event) {
					e.PreventDefault()
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

// itoa converts an int to string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := [20]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte(n%10) + "0"[0]
		n /= 10
	}
	if neg {
		pos--
		buf[pos] = "-"[0]
	}
	return string(buf[pos:])
}
