//go:build js && wasm

package component

import (
	"syscall/js"
	"testing"

	"github.com/seanrobmerriam/gowasm/pkg/dom"
	"github.com/seanrobmerriam/gowasm/pkg/reactive"
)

func newTestContainer() dom.Element {
	el := dom.Doc().CreateElement("div")
	dom.Doc().Body().AppendChild(el)
	return el
}

func cleanupContainer(el dom.Element) {
	el.Remove()
}

func firstChild(container dom.Element) js.Value {
	return container.JSValue().Get("firstChild")
}

func TestHCreatesElement(t *testing.T) {
	container := newTestContainer()
	defer cleanupContainer(container)

	node := H("div")
	node.Mount(container)

	if got := firstChild(container).Get("tagName").String(); got != "DIV" {
		t.Errorf("expected DIV, got %q", got)
	}
}

func TestHAppliesAttrs(t *testing.T) {
	container := newTestContainer()
	defer cleanupContainer(container)

	node := H("div", Attr("data-test", "hello"))
	node.Mount(container)

	if got := firstChild(container).Call("getAttribute", "data-test").String(); got != "hello" {
		t.Errorf("expected hello, got %q", got)
	}
}

func TestHAppliesClass(t *testing.T) {
	container := newTestContainer()
	defer cleanupContainer(container)

	node := H("div", Class("a", "b"))
	node.Mount(container)

	if got := firstChild(container).Get("className").String(); got != "a b" {
		t.Errorf("expected class 'a b', got %q", got)
	}
}

func TestHAppliesStyle(t *testing.T) {
	container := newTestContainer()
	defer cleanupContainer(container)

	node := H("div", Style("color", "red"))
	node.Mount(container)

	if got := firstChild(container).Get("style").Get("color").String(); got != "red" {
		t.Errorf("expected red, got %q", got)
	}
}

func TestHAppliesID(t *testing.T) {
	container := newTestContainer()
	defer cleanupContainer(container)

	node := H("div", ID("my-id"))
	node.Mount(container)

	if got := firstChild(container).Get("id").String(); got != "my-id" {
		t.Errorf("expected my-id, got %q", got)
	}
}

func TestHMountsChildren(t *testing.T) {
	container := newTestContainer()
	defer cleanupContainer(container)

	node := H("div", Children(
		H("span", Children(Text("child"))),
	))
	node.Mount(container)

	if got := firstChild(container).Get("textContent").String(); got != "child" {
		t.Errorf("expected child, got %q", got)
	}
}

func TestHPatchAttrs(t *testing.T) {
	container := newTestContainer()
	defer cleanupContainer(container)

	oldNode := H("div", Attr("data-test", "old"))
	oldNode.Mount(container)
	oldEl := oldNode.domEl.JSValue()

	newNode := H("div", Attr("data-test", "new"))
	if !newNode.Patch(oldNode) {
		t.Fatal("expected patch to succeed")
	}

	if got := newNode.domEl.JSValue().Call("getAttribute", "data-test").String(); got != "new" {
		t.Errorf("expected new, got %q", got)
	}
	if !newNode.domEl.JSValue().Equal(oldEl) {
		t.Error("expected patch to reuse the same DOM element")
	}
}

func TestHPatchChildren(t *testing.T) {
	container := newTestContainer()
	defer cleanupContainer(container)

	oldNode := H("div", Children(Text("one"), Text("two")))
	oldNode.Mount(container)

	newNode := H("div", Children(Text("one"), Text("updated")))
	if !newNode.Patch(oldNode) {
		t.Fatal("expected patch to succeed")
	}

	if got := firstChild(container).Get("textContent").String(); got != "oneupdated" {
		t.Errorf("expected oneupdated, got %q", got)
	}
}

func TestHUnmount(t *testing.T) {
	container := newTestContainer()
	defer cleanupContainer(container)

	clicks := 0
	node := H("button", On("click", func(e dom.Event) {
		clicks++
	}))
	node.Mount(container)

	node.Unmount()
	if got := container.JSValue().Get("childNodes").Get("length").Int(); got != 0 {
		t.Fatalf("expected no children after unmount, got %d", got)
	}

	click := js.Global().Get("Event").New("click")
	node.domEl.JSValue().Call("dispatchEvent", click)
	if clicks != 0 {
		t.Errorf("expected listener to be released on unmount, got %d clicks", clicks)
	}
}

func TestTextMounts(t *testing.T) {
	container := newTestContainer()
	defer cleanupContainer(container)

	node := Text("hello")
	node.Mount(container)

	if got := container.JSValue().Get("textContent").String(); got != "hello" {
		t.Errorf("expected hello, got %q", got)
	}
}

func TestTextPatch(t *testing.T) {
	container := newTestContainer()
	defer cleanupContainer(container)

	oldNode := Text("hello")
	oldNode.Mount(container)

	newNode := Text("updated")
	newNode.domEl = oldNode.domEl
	if !newNode.Patch(oldNode) {
		t.Fatal("expected text patch to succeed")
	}

	if got := container.JSValue().Get("textContent").String(); got != "updated" {
		t.Errorf("expected updated, got %q", got)
	}
}

func TestTextUnmount(t *testing.T) {
	container := newTestContainer()
	defer cleanupContainer(container)

	node := Text("hello")
	node.Mount(container)
	node.Unmount()

	if got := container.JSValue().Get("childNodes").Get("length").Int(); got != 0 {
		t.Errorf("expected no children after unmount, got %d", got)
	}
}

type testRenderComponent struct {
	val *reactive.Signal[string]
}

func (tc *testRenderComponent) Render() Node {
	return H("div", Children(Text(tc.val.Get())))
}

type lifecycleComponent struct {
	mounts   *int
	unmounts *int
}

func (lc *lifecycleComponent) Render() Node {
	return H("div", Children(Text("life")))
}

func (lc *lifecycleComponent) OnMount() {
	*lc.mounts = *lc.mounts + 1
}

func (lc *lifecycleComponent) OnUnmount() {
	*lc.unmounts = *lc.unmounts + 1
}

func TestComponentMounts(t *testing.T) {
	container := newTestContainer()
	defer cleanupContainer(container)

	tc := &testRenderComponent{val: reactive.NewSignal("hello")}
	node := C(tc)
	node.Mount(container)
	defer node.Unmount()

	if got := firstChild(container).Get("textContent").String(); got != "hello" {
		t.Errorf("expected hello, got %q", got)
	}
}

func TestComponentRerendersOnSignal(t *testing.T) {
	container := newTestContainer()
	defer cleanupContainer(container)

	tc := &testRenderComponent{val: reactive.NewSignal("hello")}
	node := C(tc)
	node.Mount(container)
	defer node.Unmount()

	tc.val.Set("updated")
	if got := firstChild(container).Get("textContent").String(); got != "updated" {
		t.Errorf("expected updated, got %q", got)
	}
}

func TestComponentOnMount(t *testing.T) {
	container := newTestContainer()
	defer cleanupContainer(container)

	mounts := 0
	unmounts := 0
	node := C(&lifecycleComponent{mounts: &mounts, unmounts: &unmounts})
	node.Mount(container)
	defer node.Unmount()

	if mounts != 1 {
		t.Errorf("expected OnMount once, got %d", mounts)
	}
}

func TestComponentOnUnmount(t *testing.T) {
	container := newTestContainer()
	defer cleanupContainer(container)

	mounts := 0
	unmounts := 0
	node := C(&lifecycleComponent{mounts: &mounts, unmounts: &unmounts})
	node.Mount(container)
	node.Unmount()

	if unmounts != 1 {
		t.Errorf("expected OnUnmount once, got %d", unmounts)
	}
}

func TestComponentUnmountDisposes(t *testing.T) {
	container := newTestContainer()
	defer cleanupContainer(container)

	tc := &testRenderComponent{val: reactive.NewSignal("hello")}
	node := C(tc)
	node.Mount(container)
	node.Unmount()

	if node.effect != nil {
		t.Fatal("expected effect to be nil after unmount")
	}

	tc.val.Set("updated")
	if got := container.JSValue().Get("childNodes").Get("length").Int(); got != 0 {
		t.Errorf("expected container to remain empty after unmount, got %d children", got)
	}
}
