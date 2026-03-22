package router

import "syscall/js"

// history is the package-level hash state manager.
var history = &hashHistory{}

type hashHistory struct {
	current     string
	listeners   []func(path string)
	handler     js.Func
	initialized bool
}

// init installs the hashchange listener and reads the initial path.
func (h *hashHistory) init() {
	// Guard against double-initialisation
	if h.initialized {
		return
	}

	h.current = parsePath(js.Global().Get("location").Get("hash").String())

	h.handler = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		newPath := parsePath(js.Global().Get("location").Get("hash").String())
		h.current = newPath
		for _, l := range h.listeners {
			l(newPath)
		}
		return nil
	})

	js.Global().Call("addEventListener", "hashchange", h.handler)
	h.initialized = true
}

// navigate pushes a new hash path.
func (h *hashHistory) navigate(path string) {
	js.Global().Get("location").Set("hash", path)
}

// onchange registers a listener called on every path change.
func (h *hashHistory) onchange(fn func(path string)) {
	h.listeners = append(h.listeners, fn)
}

// parsePath strips the leading # and ensures a leading slash.
// "" → "/"    "#" → "/"    "#/about" → "/about"
func parsePath(hash string) string {
	if len(hash) == 0 {
		return "/"
	}
	if hash[0] == '#' {
		hash = hash[1:]
	}
	if len(hash) == 0 || hash[0] != '/' {
		hash = "/" + hash
	}
	return hash
}
