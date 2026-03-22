package reactive

import "sync"

var batchDepth int
var batchQueue []func()
var batchMu sync.Mutex

// Batch defers all subscriber notifications until fn returns, then flushes once.
func Batch(fn func()) {
	batchMu.Lock()
	batchDepth++
	batchMu.Unlock()

	fn()

	batchMu.Lock()
	batchDepth--
	if batchDepth == 0 {
		queue := batchQueue
		batchQueue = nil
		batchMu.Unlock()
		for _, fn := range queue {
			fn()
		}
		return
	}
	batchMu.Unlock()
}

// Effect represents an effect that re-runs when its dependencies change.
type Effect struct {
	fn         func()
	subscribed []*SignalBase
	mu         sync.Mutex
	disposed   bool
	running    bool
}

// NewEffect creates a new effect that runs fn immediately and re-runs when dependencies change.
func NewEffect(fn func()) *Effect {
	e := &Effect{fn: fn}

	pushObserver(e)
	fn()
	popObserver()

	return e
}

// run executes the effect function and updates its subscriptions.
func (e *Effect) run() {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return
	}
	e.running = true
	oldDeps := e.subscribed
	e.subscribed = nil
	e.mu.Unlock()

	for _, sig := range oldDeps {
		sig.removeSubscriber(e)
	}

	pushObserver(e)
	e.fn()
	popObserver()

	e.mu.Lock()
	e.running = false
	e.mu.Unlock()
}

// notify implements the observer interface.
func (e *Effect) notify() {
	e.mu.Lock()
	isDone := e.disposed
	e.mu.Unlock()
	if isDone {
		return
	}
	e.run()
}

// addDependency adds a signal to the effect's dependency list.
func (e *Effect) addDependency(sig *SignalBase) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, s := range e.subscribed {
		if s == sig {
			return
		}
	}
	e.subscribed = append(e.subscribed, sig)
}

// removeDependency removes a signal from the effect's dependency list.
func (e *Effect) removeDependency(sig *SignalBase) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for i, s := range e.subscribed {
		if s == sig {
			e.subscribed = append(e.subscribed[:i], e.subscribed[i+1:]...)
			return
		}
	}
}

// Dispose stops the effect and releases all subscriptions.
func (e *Effect) Dispose() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.disposed {
		return
	}

	e.disposed = true

	for _, sig := range e.subscribed {
		sig.removeSubscriber(e)
	}
	e.subscribed = nil
}
