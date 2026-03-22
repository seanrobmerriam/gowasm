package main

import "github.com/seanrobmerriam/gowasm/pkg/reactive"

// TodoItem is a single task.
type TodoItem struct {
	ID   int
	Text string
	Done bool
}

// Filter controls which subset of items is visible.
type Filter string

const (
	FilterAll       Filter = "all"
	FilterActive    Filter = "active"
	FilterCompleted Filter = "completed"
)

// AppState is the single source of truth for the todo app.
// All fields are plain values; reactivity is handled by the
// Signal that wraps the *slice* of items.
type AppState struct {
	NextID int
	Items  []TodoItem
}

// ── Reactive roots ──────────────────────────────────────────

// state holds every todo item.
var state = reactive.NewSignal(AppState{NextID: 1})

// inputText is the controlled value of the "new todo" text field.
var inputText = reactive.NewSignal("")

// activeFilter drives which items are displayed.
var activeFilter = reactive.NewSignal(FilterAll)

// ── Derived (Computed) values ───────────────────────────────

// visibleItems re-evaluates whenever state or activeFilter changes.
var visibleItems = reactive.NewComputed(func() []TodoItem {
	s := state.Get()
	f := activeFilter.Get()
	if f == FilterAll {
		return s.Items
	}
	out := make([]TodoItem, 0, len(s.Items))
	for _, item := range s.Items {
		if f == FilterCompleted && item.Done {
			out = append(out, item)
		}
		if f == FilterActive && !item.Done {
			out = append(out, item)
		}
	}
	return out
})

// activeCount and completedCount are used in the footer.
var activeCount = reactive.NewComputed(func() int {
	n := 0
	for _, item := range state.Get().Items {
		if !item.Done {
			n++
		}
	}
	return n
})

var completedCount = reactive.NewComputed(func() int {
	n := 0
	for _, item := range state.Get().Items {
		if item.Done {
			n++
		}
	}
	return n
})
