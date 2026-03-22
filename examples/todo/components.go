package main

import (
	"fmt"

    "github.com/seanrobmerriam/gowasm/pkg/component"
    "github.com/seanrobmerriam/gowasm/pkg/dom"
)

// HeaderComponent renders the top bar: the "toggle all" checkbox
// and the controlled text input for new todos.
func HeaderComponent() component.Component {
    return component.El("header",
        component.Attr("class", "header"),

        // ── Toggle-all checkbox ──────────────────────────────
        // Only render it when there is at least one item.
        component.Computed(func() component.Component {
            if len(state.Get().Items) == 0 {
                return component.El("span") // empty placeholder
            }
            allDone := activeCount.Get() == 0
            return component.Checkbox(
                inputText, // ignored — see note below
                component.Attr("id", "toggle-all"),
                component.Attr("class", "toggle-all"),
                component.Checked(allDone),
                component.OnChange(func(e dom.Event) {
                    toggleAll(!allDone)
                }),
            )
        }),

        // ── New-todo text input ──────────────────────────────
        component.Input(
            inputText, // two-way binding: signal → value, value → signal
            component.Attr("class", "new-todo"),
            component.Attr("placeholder", "What needs to be done?"),
            component.Attr("autofocus", "true"),
            component.OnKeyDown(func(e dom.Event) {
                if e.Key() == "Enter" {
                    addTodo()
                }
            }),
        ),
    )
    // TodoItemComponent renders one row: checkbox, label, delete button.
    // It receives a plain TodoItem value — no signal — because the
    // parent re-renders the whole list whenever visibleItems changes.
    func TodoItemComponent(item TodoItem) component.Component {
        class := "todo-item"
        if item.Done {
            class += " completed"
        }

        return component.El("li",
            component.Attr("class", class),

            // Done checkbox — toggle on change
            component.Checkbox(
                reactive.NewSignal(item.Done), // local read-only signal for initial value
                component.Attr("class", "toggle"),
                component.OnChange(func(e dom.Event) {
                    toggleTodo(item.ID)
                }),
            ),

            // Label — the text of the task
            component.El("label",
                component.Text(item.Text),
            ),

            // Delete button
            component.El("button",
                component.Attr("class", "destroy"),
                component.OnClick(func(e dom.Event) {
                    deleteTodo(item.ID)
                }),
                component.Text("✕"),
            ),
        )
    }

    // ListComponent renders the <ul> of visible todo items.
    // It re-evaluates whenever visibleItems changes.
    func ListComponent() component.Component {
        return component.El("section",
            component.Attr("class", "main"),

            component.Computed(func() component.Component {
                items := visibleItems.Get()
                children := make([]component.Option, 0, len(items))
                for _, item := range items {
                    // Capture loop variable explicitly — required in Go <1.22
                    item := item
                    children = append(children, TodoItemComponent(item))
                }
                return component.El("ul",
                    component.Attr("class", "todo-list"),
                    component.Children(children...),
                )
            }),
        )
    }

    // FooterComponent renders counts and the filter tab bar.
    func FooterComponent() component.Component {
        return component.Computed(func() component.Component {
            active := activeCount.Get()
            completed := completedCount.Get()

            // Hide the footer entirely when the list is empty.
            if active+completed == 0 {
                return component.El("span")
            }

            itemWord := "items"
            if active == 1 {
                itemWord = "item"
            }

            return component.El("footer",
                component.Attr("class", "footer"),

                // Item count
                component.El("span",
                    component.Attr("class", "todo-count"),
                    component.Text(fmt.Sprintf("%d %s left", active, itemWord)),
                ),

                // Filter tabs
                component.El("ul",
                    component.Attr("class", "filters"),
                    filterTab("All",       FilterAll),
                    filterTab("Active",    FilterActive),
                    filterTab("Completed", FilterCompleted),
                ),

                // Clear completed (only shown when there are some)
                component.Computed(func() component.Component {
                    if completedCount.Get() == 0 {
                        return component.El("span")
                    }
                    return component.El("button",
                        component.Attr("class", "clear-completed"),
                        component.OnClick(func(e dom.Event) { clearCompleted() }),
                        component.Text("Clear completed"),
                    )
                }),
            )
        })
    }

    // filterTab builds one <li><a> filter link and highlights the active one.
    func filterTab(label string, f Filter) component.Option {
        return component.Computed(func() component.Component {
            class := ""
            if activeFilter.Get() == f {
                class = "selected"
            }
            return component.El("li",
                component.El("a",
                    component.Attr("class", class),
                    component.Attr("href", "#"),
                    component.OnClick(func(e dom.Event) {
                        e.PreventDefault()
                        activeFilter.Set(f)
                    }),
                    component.Text(label),
                ),
            )
        })
    }

    // AppComponent is the root — it simply stacks the three sections.
    func AppComponent() component.Component {
        return component.El("div",
            component.Attr("class", "todoapp"),
            HeaderComponent(),
            ListComponent(),
            FooterComponent(),
        )
    }
}
