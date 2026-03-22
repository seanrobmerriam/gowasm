//go:build js && wasm

package dom

import (
	"syscall/js"
	"testing"
)

func newTestContainer() Element {
	el := Doc().CreateElement("div")
	Doc().Body().AppendChild(el)
	return el
}

func cleanupContainer(el Element) {
	el.Remove()
}

func TestDocGetElementByID(t *testing.T) {
	container := newTestContainer()
	defer cleanupContainer(container)
	container.SetAttr("id", "dom-test-id")

	el, ok := Doc().GetElementByID("dom-test-id")
	if !ok {
		t.Fatal("expected element to be found")
	}
	if got := el.JSValue().Get("id").String(); got != "dom-test-id" {
		t.Errorf("expected id dom-test-id, got %q", got)
	}
}

func TestDocGetElementByIDMiss(t *testing.T) {
	if _, ok := Doc().GetElementByID("does-not-exist"); ok {
		t.Fatal("expected lookup miss")
	}
}

func TestDocQuerySelector(t *testing.T) {
	container := newTestContainer()
	defer cleanupContainer(container)
	container.SetAttr("class", "query-target")

	el, ok := Doc().QuerySelector(".query-target")
	if !ok {
		t.Fatal("expected selector to match")
	}
	if got := el.JSValue().Get("className").String(); got != "query-target" {
		t.Errorf("expected class query-target, got %q", got)
	}
}

func TestDocCreateElement(t *testing.T) {
	el := Doc().CreateElement("section")
	if got := el.JSValue().Get("tagName").String(); got != "SECTION" {
		t.Errorf("expected SECTION, got %q", got)
	}
}

func TestDocCreateTextNode(t *testing.T) {
	node := Doc().CreateTextNode("hello")
	if got := node.JSValue().Get("nodeType").Int(); got != 3 {
		t.Errorf("expected text node type 3, got %d", got)
	}
}

func TestElementSetAttr(t *testing.T) {
	el := NewElement("div")
	el.SetAttr("data-test", "hello")
	if got := el.JSValue().Call("getAttribute", "data-test").String(); got != "hello" {
		t.Errorf("expected hello, got %q", got)
	}
}

func TestElementRemoveAttr(t *testing.T) {
	el := NewElement("div")
	el.SetAttr("data-test", "hello")
	el.RemoveAttr("data-test")
	if got := el.JSValue().Call("hasAttribute", "data-test").Bool(); got {
		t.Error("expected attribute to be removed")
	}
}

func TestElementSetStyle(t *testing.T) {
	el := NewElement("div")
	el.SetStyle("color", "red")
	if got := el.JSValue().Get("style").Get("color").String(); got != "red" {
		t.Errorf("expected red, got %q", got)
	}
}

func TestElementAppendChild(t *testing.T) {
	parent := NewElement("div")
	child := NewElement("span")
	parent.AppendChild(child)
	if got := parent.JSValue().Get("childNodes").Get("length").Int(); got != 1 {
		t.Errorf("expected one child, got %d", got)
	}
}

func TestElementRemove(t *testing.T) {
	container := newTestContainer()
	defer cleanupContainer(container)
	child := NewElement("div")
	container.AppendChild(child)
	child.Remove()
	if got := container.JSValue().Get("childNodes").Get("length").Int(); got != 0 {
		t.Errorf("expected child to be removed, got %d children", got)
	}
}

func TestElementSetInnerHTML(t *testing.T) {
	el := NewElement("div")
	el.SetInnerHTML("<span>hi</span>")
	if got := el.JSValue().Get("innerHTML").String(); got != "<span>hi</span>" {
		t.Errorf("expected innerHTML to be set, got %q", got)
	}
}

func TestElementSetTextContent(t *testing.T) {
	el := NewElement("div")
	el.SetTextContent("hello")
	if got := el.JSValue().Get("textContent").String(); got != "hello" {
		t.Errorf("expected hello, got %q", got)
	}
}

func TestElementSetProperty(t *testing.T) {
	el := NewElement("input")
	el.SetProperty("value", "abc")
	if got := el.JSValue().Get("value").String(); got != "abc" {
		t.Errorf("expected abc, got %q", got)
	}
}

func TestAddRemoveEventListener(t *testing.T) {
	el := NewElement("button")
	count := 0
	handle := el.AddEventListener("click", func(e Event) {
		count++
	})

	click := js.Global().Get("Event").New("click")
	el.JSValue().Call("dispatchEvent", click)
	if count != 1 {
		t.Fatalf("expected click handler to run once, got %d", count)
	}

	handle.Release()
	el.JSValue().Call("dispatchEvent", click)
	if count != 1 {
		t.Errorf("expected released listener not to run again, got %d", count)
	}
}

func TestTextNodeSetText(t *testing.T) {
	node := Doc().CreateTextNode("hello")
	node.SetText("updated")
	if got := node.JSValue().Get("textContent").String(); got != "updated" {
		t.Errorf("expected updated, got %q", got)
	}
}

func TestTextNodeAppendTo(t *testing.T) {
	parent := NewElement("div")
	node := Doc().CreateTextNode("hello")
	node.AppendTo(parent)
	if got := parent.JSValue().Get("textContent").String(); got != "hello" {
		t.Errorf("expected hello, got %q", got)
	}
}

func TestTextNodeRemove(t *testing.T) {
	parent := NewElement("div")
	node := Doc().CreateTextNode("hello")
	node.AppendTo(parent)
	node.Remove()
	if got := parent.JSValue().Get("childNodes").Get("length").Int(); got != 0 {
		t.Errorf("expected text node to be removed, got %d children", got)
	}
}
