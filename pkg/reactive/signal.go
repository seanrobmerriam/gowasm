// Package reactive provides fine-grained reactivity with automatic dependency tracking.
package reactive

import "sync"

type observer interface {
	notify()
	addDependency(*SignalBase)
	removeDependency(*SignalBase)
}

// activeObserver is the currently executing observer (effect or computed).
var activeObserver observer

// observerStack is the stack of observers for nested tracking.
var observerStack []observer

// pushObserver pushes an observer onto the stack.
func pushObserver(o observer) {
	observerStack = append(observerStack, o)
}

// popObserver pops the current observer from the stack.
func popObserver() {
	n := len(observerStack)
	if n > 0 {
		observerStack = observerStack[:n-1]
	}
}

// currentObserver returns the top of the observer stack or nil.
func currentObserver() observer {
	n := len(observerStack)
	if n == 0 {
		return nil
	}
	return observerStack[n-1]
}

// SignalBase is the internal non-generic signal implementation.
type SignalBase struct {
	value   any
	mu      sync.RWMutex
	subs    []observer
	dirty   bool
	queued  bool
	version uint64
}

// newSignalBase creates a new signal base.
func newSignalBase(initial any) *SignalBase {
	return &SignalBase{
		value:   initial,
		dirty:   false,
		version: 0,
	}
}

// getValue returns the current value.
func (s *SignalBase) getValue() any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.value
}

// setValue sets the value and notifies subscribers.
func (s *SignalBase) setValue(v any) {
	s.mu.Lock()
	s.value = v
	s.version++
	s.dirty = false
	s.mu.Unlock()

	batchMu.Lock()
	if batchDepth > 0 {
		s.mu.Lock()
		if !s.queued {
			s.queued = true
			batchQueue = append(batchQueue, func() {
				s.mu.Lock()
				s.queued = false
				s.mu.Unlock()
				s.notifySubscribers()
			})
		}
		s.mu.Unlock()
		batchMu.Unlock()
		return
	}
	batchMu.Unlock()

	s.notifySubscribers()
}

// peekValue reads the value without tracking.
func (s *SignalBase) peekValue() any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.value
}

// addSubscriber adds an observer to the subscriber list.
func (s *SignalBase) addSubscriber(o observer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Check if already subscribed
	for _, sub := range s.subs {
		if sub == o {
			return
		}
	}
	s.subs = append(s.subs, o)
}

// removeSubscriber removes an observer from the subscriber list.
func (s *SignalBase) removeSubscriber(o observer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, sub := range s.subs {
		if sub == o {
			s.subs = append(s.subs[:i], s.subs[i+1:]...)
			return
		}
	}
}

// notifySubscribers notifies all subscribers that the signal changed.
func (s *SignalBase) notifySubscribers() {
	s.mu.RLock()
	subs := make([]observer, len(s.subs))
	copy(subs, s.subs)
	s.mu.RUnlock()

	for _, sub := range subs {
		sub.notify()
	}
}

// markDirty marks the signal as dirty (value has changed but subscribers not notified).
func (s *SignalBase) markDirty() {
	s.mu.Lock()
	s.dirty = true
	s.mu.Unlock()
}

// isDirty returns true if the signal is dirty.
func (s *SignalBase) isDirty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dirty
}

// version_ returns the current version.
func (s *SignalBase) version_() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.version
}

// notify implements the observer interface for SignalBase.
func (s *SignalBase) notify() {
	s.mu.Lock()
	s.dirty = true
	s.mu.Unlock()
}

// addDependency adds a signal as a dependency (no-op for signal base).
func (s *SignalBase) addDependency(sig *SignalBase) {
	// Signal bases don't track dependencies themselves
}

// removeDependency removes a signal dependency (no-op for signal base).
func (s *SignalBase) removeDependency(sig *SignalBase) {
	// Signal bases don't track dependencies themselves
}

// Signal is a reactive signal that notifies subscribers when its value changes.
type Signal[T any] struct {
	base *SignalBase
}

// NewSignal creates a new signal with the given initial value.
func NewSignal[T any](initial T) *Signal[T] {
	return &Signal[T]{
		base: newSignalBase(initial),
	}
}

// Get returns the current value and records this signal as a dependency
// of the currently executing observer.
func (s *Signal[T]) Get() T {
	// Record this signal as a dependency of the current observer
	if obs := currentObserver(); obs != nil {
		obs.addDependency(s.base)
		s.base.addSubscriber(obs)
	}
	return s.base.peekValue().(T)
}

// Set sets the signal's value and notifies all subscribers.
func (s *Signal[T]) Set(v T) {
	s.base.setValue(v)
}

// Peek returns the current value without recording a dependency.
func (s *Signal[T]) Peek() T {
	return s.base.peekValue().(T)
}
