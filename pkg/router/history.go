package router

import "syscall/js"

// history is the package-level hash state manager.
var history = &hashHistory{}
var histAPI = &historyAPI{}

type hashHistory struct {
	current     string
	listeners   []func(path string)
	handler     js.Func
	initialized bool
}

// historyAPI manages pushState-based navigation.
type historyAPI struct {
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

// init installs the popstate listener and reads the initial path.
func (h *historyAPI) init() {
	if h.initialized {
		return
	}

	h.current = currentLocationPath()

	h.handler = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		newPath := currentLocationPath()
		h.current = newPath
		for _, l := range h.listeners {
			l(newPath)
		}
		return nil
	})

	js.Global().Call("addEventListener", "popstate", h.handler)
	h.initialized = true
}

// navigate pushes a new path using pushState.
func (h *historyAPI) navigate(path string) {
	js.Global().Get("history").Call("pushState", nil, "", path)
	// pushState does not fire popstate — notify listeners manually
	h.current = path
	for _, l := range h.listeners {
		l(path)
	}
}

// onchange registers a listener called on every path change.
func (h *historyAPI) onchange(fn func(path string)) {
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

func currentLocationPath() string {
	location := js.Global().Get("location")
	path := location.Get("pathname").String()
	search := location.Get("search").String()
	if search != "" {
		path += search
	}
	if path == "" {
		return "/"
	}
	return path
}
