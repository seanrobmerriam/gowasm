package ssr_test

import (
	"strings"
	"testing"

	"github.com/seanrobmerriam/gowasm/pkg/component"
	"github.com/seanrobmerriam/gowasm/pkg/ssr"
)

type staticComponent struct {
	node component.Node
}

func (s *staticComponent) Render() component.Node {
	return s.node
}

type helloComponent struct{}

func (h *helloComponent) Render() component.Node {
	return component.H("p", component.Children(
		component.Text("hello from SSR"),
	))
}

func render(t *testing.T, c interface{}) string {
	t.Helper()
	r := ssr.New()
	html, err := r.RenderToString(c)
	if err != nil {
		t.Fatalf("RenderToString failed: %v", err)
	}
	return html
}

func TestRenderText(t *testing.T) {
	html := render(t, &staticComponent{node: component.Text("hello")})
	if html != "hello" {
		t.Fatalf("expected hello, got %q", html)
	}
}

func TestRenderElement(t *testing.T) {
	html := render(t, &staticComponent{node: component.H("div")})
	if html != "<div></div>" {
		t.Fatalf("expected <div></div>, got %q", html)
	}
}

func TestRenderElementWithAttrs(t *testing.T) {
	html := render(t, &staticComponent{node: component.H("div", component.Attr("id", "foo"))})
	if html != `<div id="foo"></div>` {
		t.Fatalf("expected id attr, got %q", html)
	}
}

func TestRenderElementWithClass(t *testing.T) {
	html := render(t, &staticComponent{node: component.H("div", component.Class("a", "b"))})
	if html != `<div class="a b"></div>` {
		t.Fatalf("expected class attr, got %q", html)
	}
}

func TestRenderElementWithStyle(t *testing.T) {
	html := render(t, &staticComponent{node: component.H("div", component.Style("color", "red"))})
	if html != `<div style="color: red;"></div>` {
		t.Fatalf("expected style attr, got %q", html)
	}
}

func TestRenderVoidElement(t *testing.T) {
	html := render(t, &staticComponent{node: component.H("input")})
	if html != `<input/>` {
		t.Fatalf("expected <input/>, got %q", html)
	}
}

func TestRenderNested(t *testing.T) {
	html := render(t, &staticComponent{node: component.H("div", component.Children(
		component.H("span", component.Children(component.Text("hi"))),
	))})
	if html != `<div><span>hi</span></div>` {
		t.Fatalf("expected nested html, got %q", html)
	}
}

func TestRenderTextEscaping(t *testing.T) {
	html := render(t, &staticComponent{node: component.Text(`<>&"'`)})
	if html != `&lt;&gt;&amp;&quot;&#39;` {
		t.Fatalf("expected escaped text, got %q", html)
	}
}

func TestRenderAttrEscaping(t *testing.T) {
	html := render(t, &staticComponent{node: component.H("div", component.Attr("title", `<>&"'`))})
	if html != `<div title="&lt;&gt;&amp;&quot;&#39;"></div>` {
		t.Fatalf("expected escaped attr, got %q", html)
	}
}

func TestRenderComponent(t *testing.T) {
	html := render(t, &helloComponent{})
	if !strings.Contains(html, "<p>hello from SSR</p>") {
		t.Fatalf("expected component html, got %q", html)
	}
}

func TestRenderKey(t *testing.T) {
	html := render(t, &staticComponent{node: component.H("div", component.Key("abc"))})
	if html != `<div data-gowasm-key="abc"></div>` {
		t.Fatalf("expected key attr, got %q", html)
	}
}

func TestRenderPage(t *testing.T) {
	page := ssr.RenderPage("<p>hello</p>", ssr.PageOptions{
		Title:  "Test",
		RootID: "app",
	})
	if !strings.Contains(page, `<div id="app" data-ssr="true"><p>hello</p></div>`) {
		t.Fatalf("expected SSR root div, got %q", page)
	}
	if !strings.Contains(page, "<title>Test</title>") {
		t.Fatalf("expected title, got %q", page)
	}
}

func TestRenderDeterministic(t *testing.T) {
	comp := &staticComponent{node: component.H("div",
		component.Attr("b", "2"),
		component.Attr("a", "1"),
		component.Style("color", "red"),
		component.Style("background", "white"),
	)}
	first := render(t, comp)
	second := render(t, comp)
	if first != second {
		t.Fatalf("expected deterministic output, got %q != %q", first, second)
	}
}
