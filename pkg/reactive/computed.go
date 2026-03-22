package reactive

// Computed represents a computed value that automatically tracks its dependencies.
type Computed[T any] struct {
	signal   SignalBase
	fn       func() T
	cached   T
	hasValue bool
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
	return c
}

// Get returns the computed value, re-running the function if any dependency changed.
func (c *Computed[T]) Get() T {
	// Register as subscriber of dependencies
	current := currentObserver()
	if current != nil {
		c.signal.addSubscriber(current)
	}

	// Check if we need to recompute
	if c.signal.dirty || !c.hasValue {
		// Push this computed onto the observer stack so dependencies register with it
		pushObserver(&c.signal)
		c.cached = c.fn()
		popObserver()

		c.hasValue = true
		c.signal.dirty = false
	}

	return c.cached
}

// addDependency implements the observer interface (no-op for computed).
func (c *Computed[T]) addDependency(sig *SignalBase) {
	// Computed values track their own dependencies via the observer stack
}

// removeDependency implements the observer interface (no-op for computed).
func (c *Computed[T]) removeDependency(sig *SignalBase) {
	// Computed values track their own dependencies via the observer stack
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
