package reactive

// computedObserver wraps a Computed and implements the observer interface.
type computedObserver[T any] struct {
	c *Computed[T]
}

func (co *computedObserver[T]) notify() {
	co.c.signal.mu.Lock()
	co.c.signal.dirty = true
	co.c.signal.mu.Unlock()
}

func (co *computedObserver[T]) addDependency(sig *SignalBase) {
	sig.addSubscriber(co)
}

func (co *computedObserver[T]) removeDependency(sig *SignalBase) {
	sig.removeSubscriber(co)
}

// Computed represents a computed value that automatically tracks its dependencies.
type Computed[T any] struct {
	signal   SignalBase
	fn       func() T
	cached   T
	hasValue bool
	obs      *computedObserver[T]
}

// NewComputed creates a new computed value that re-evaluates when dependencies change.
func NewComputed[T any](fn func() T) *Computed[T] {
	c := &Computed[T]{
		signal: SignalBase{
			dirty:   false,
			version: 0,
		},
		fn:       fn,
		hasValue: false,
	}
	c.obs = &computedObserver[T]{c: c}
	return c
}

// Get returns the computed value, re-running the function if any dependency changed.
func (c *Computed[T]) Get() T {
	// Register as subscriber so anyone reading this computed gets notified when it changes
	current := currentObserver()
	if current != nil {
		c.signal.addSubscriber(current)
	}

	// Check if we need to recompute
	if c.signal.dirty || !c.hasValue {
		// Push the computedObserver onto the stack so dependencies register with it
		pushObserver(c.obs)
		c.cached = c.fn()
		popObserver()

		c.hasValue = true
		c.signal.dirty = false
	}

	return c.cached
}

// addDependency implements the observer interface.
func (c *Computed[T]) addDependency(sig *SignalBase) {
	sig.addSubscriber(c.obs)
}

// removeDependency implements the observer interface.
func (c *Computed[T]) removeDependency(sig *SignalBase) {
	sig.removeSubscriber(c.obs)
}

// notify implements the observer interface.
func (c *Computed[T]) notify() {
	c.signal.mu.Lock()
	c.signal.dirty = true
	c.signal.mu.Unlock()
}

// version returns the version of the underlying signal.
func (c *Computed[T]) version() uint64 {
	return c.signal.version_()
}
