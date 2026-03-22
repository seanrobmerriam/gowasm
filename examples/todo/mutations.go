package main

// addTodo appends a new item using the current inputText value.
// It is a no-op if the input is blank.
func addTodo() {
	text := inputText.Get()
	if text == "" {
		return
	}

	s := state.Get()
	state.Set(AppState{
		NextID: s.NextID + 1,
		Items: append(s.Items, TodoItem{
			ID:   s.NextID,
			Text: text,
			Done: false,
		}),
	})

	inputText.Set("") // clear the field after adding
}

// toggleTodo flips the Done flag for the item with the given ID.
func toggleTodo(id int) {
	s := state.Get()
	items := make([]TodoItem, len(s.Items))
	copy(items, s.Items)

	for i, item := range items {
		if item.ID == id {
			items[i].Done = !item.Done
			break
		}
	}

	state.Set(AppState{NextID: s.NextID, Items: items})
}

// deleteTodo removes the item with the given ID.
func deleteTodo(id int) {
	s := state.Get()
	items := make([]TodoItem, 0, len(s.Items))
	for _, item := range s.Items {
		if item.ID != id {
			items = append(items, item)
		}
	}
	state.Set(AppState{NextID: s.NextID, Items: items})
}

// clearCompleted removes every done item in one shot.
func clearCompleted() {
	s := state.Get()
	items := make([]TodoItem, 0, len(s.Items))
	for _, item := range s.Items {
		if !item.Done {
			items = append(items, item)
		}
	}
	state.Set(AppState{NextID: s.NextID, Items: items})
}

// toggleAll sets every item's Done flag to the given value.
// Used by the "check all" master checkbox.
func toggleAll(done bool) {
	s := state.Get()
	items := make([]TodoItem, len(s.Items))
	copy(items, s.Items)
	for i := range items {
		items[i].Done = done
	}
	state.Set(AppState{NextID: s.NextID, Items: items})
}
