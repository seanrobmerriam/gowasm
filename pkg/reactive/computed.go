package reactive

// computedObserver wraps a Computed and implements the observer interface.
type computedObserver[T any] struct {
	c *Computed[T]
}

func (co *computedObserver[T]) notify() {
	co.c.signal.mu.Lock()
	wasDirty := co.c.signal.dirty
	co.c.signal.dirty = true
	co.c.signal.mu.Unlock()
	if !wasDirty {
		co.c.signal.notifySubscribers()
	}
}

func (co *computedObserver[T]) addDependency(sig *SignalBase) {
	co.c.addDependency(sig)
}

func (co *computedObserver[T]) removeDependency(sig *SignalBase) {
	co.c.removeDependency(sig)
}

// Computed represents a computed value that automatically tracks its dependencies.
type Computed[T any] struct {
	signal   SignalBase
	fn       func() T
	cached   T
	hasValue bool
	obs      *computedObserver[T]
	deps     []*SignalBase
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
		current.addDependency(&c.signal)
		c.signal.addSubscriber(current)
	}

	// Check if we need to recompute
	if c.signal.dirty || !c.hasValue {
		for _, dep := range c.deps {
			dep.removeSubscriber(c.obs)
		}
		c.deps = nil

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
	for _, dep := range c.deps {
		if dep == sig {
			return
		}
	}
	c.deps = append(c.deps, sig)
}

// removeDependency implements the observer interface.
func (c *Computed[T]) removeDependency(sig *SignalBase) {
	for i, dep := range c.deps {
		if dep == sig {
			c.deps = append(c.deps[:i], c.deps[i+1:]...)
			return
		}
	}
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
