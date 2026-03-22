//go:build js && wasm

package main

import (
	"strings"

	"github.com/seanrobmerriam/gowasm/pkg/component"
	"github.com/seanrobmerriam/gowasm/pkg/dom"
	"github.com/seanrobmerriam/gowasm/pkg/reactive"
)

type Todo struct {
	ID        int
	Text      string
	Completed bool
}

type Filter int

const (
	FilterAll Filter = iota
	FilterActive
	FilterCompleted
)

type TodoApp struct {
	todos  *reactive.Signal[[]Todo]
	input  *reactive.Signal[string]
	filter *reactive.Signal[Filter]
	nextID int
}

func NewTodoApp() *TodoApp {
	return &TodoApp{
		todos:  reactive.NewSignal([]Todo{}),
		input:  reactive.NewSignal(""),
		filter: reactive.NewSignal(FilterAll),
		nextID: 1,
	}
}

func (a *TodoApp) addTodo() {
	text := strings.TrimSpace(a.input.Peek())
	if text == "" {
		return
	}

	todos := a.todos.Peek()
	newTodos := make([]Todo, len(todos)+1)
	copy(newTodos, todos)
	newTodos[len(todos)] = Todo{
		ID:        a.nextID,
		Text:      text,
		Completed: false,
	}

	a.nextID++
	a.todos.Set(newTodos)
	a.input.Set("")
}

func (a *TodoApp) toggleTodo(id int) {
	todos := a.todos.Peek()
	newTodos := make([]Todo, len(todos))
	copy(newTodos, todos)

	for i := range newTodos {
		if newTodos[i].ID == id {
			newTodos[i].Completed = !newTodos[i].Completed
			break
		}
	}

	a.todos.Set(newTodos)
}

func (a *TodoApp) deleteTodo(id int) {
	todos := a.todos.Peek()
	newTodos := make([]Todo, 0, len(todos))

	for _, todo := range todos {
		if todo.ID != id {
			newTodos = append(newTodos, todo)
		}
	}

	a.todos.Set(newTodos)
}

func (a *TodoApp) clearCompleted() {
	todos := a.todos.Peek()
	newTodos := make([]Todo, 0, len(todos))

	for _, todo := range todos {
		if !todo.Completed {
			newTodos = append(newTodos, todo)
		}
	}

	a.todos.Set(newTodos)
}

func (a *TodoApp) filteredTodos() []Todo {
	todos := a.todos.Peek()
	filter := a.filter.Peek()

	switch filter {
	case FilterActive:
		filtered := make([]Todo, 0, len(todos))
		for _, todo := range todos {
			if !todo.Completed {
				filtered = append(filtered, todo)
			}
		}
		return filtered
	case FilterCompleted:
		filtered := make([]Todo, 0, len(todos))
		for _, todo := range todos {
			if todo.Completed {
				filtered = append(filtered, todo)
			}
		}
		return filtered
	default:
		filtered := make([]Todo, len(todos))
		copy(filtered, todos)
		return filtered
	}
}

func (a *TodoApp) Render() component.Node {
	todos := a.todos.Get()

	var listNode component.Node = component.H("div")
	var footerNode component.Node = component.H("div")
	if len(todos) > 0 {
		listNode = component.C(&TodoList{app: a})
		footerNode = component.C(&TodoFooter{app: a})
	}

	return component.H("div",
		component.Class("todoapp"),
		component.Children(
			component.H("header", component.Children(
				component.H("h1", component.Children(component.Text("todos"))),
				component.C(&TodoInput{app: a}),
			)),
			listNode,
			footerNode,
		),
	)
}

type TodoInput struct {
	app *TodoApp
}

func (t *TodoInput) Render() component.Node {
	return component.H("div",
		component.Class("todo-input"),
		component.Children(
			component.Input(component.InputProps{
				Value:       t.app.input,
				Placeholder: "What needs to be done?",
				Type:        "text",
				OnKeyDown: func(e dom.Event) {
					if e.Value("keyCode").Int() == 13 {
						t.app.addTodo()
					}
				},
			}),
			component.H("button",
				component.On("click", func(e dom.Event) {
					t.app.addTodo()
				}),
				component.Children(component.Text("Add")),
			),
		),
	)
}

type TodoList struct {
	app *TodoApp
}

func (t *TodoList) Render() component.Node {
	t.app.todos.Get()
	t.app.filter.Get()
	todos := t.app.filteredTodos()
	children := make([]component.Node, 0, len(todos))
	for _, todo := range todos {
		children = append(children, component.C(NewTodoItem(t.app, todo)))
	}

	return component.H("ul",
		component.Class("todo-list"),
		component.Children(children...),
	)
}

type TodoItem struct {
	app  *TodoApp
	todo Todo
}

func NewTodoItem(app *TodoApp, todo Todo) *TodoItem {
	return &TodoItem{app: app, todo: todo}
}

func (t *TodoItem) Render() component.Node {
	itemClass := "todo-item"
	if t.todo.Completed {
		itemClass = "todo-item completed"
	}

	checkboxOpts := []component.Option{
		component.Attr("type", "checkbox"),
		component.On("change", func(e dom.Event) {
			t.app.toggleTodo(t.todo.ID)
		}),
	}
	if t.todo.Completed {
		checkboxOpts = append(checkboxOpts, component.Attr("checked", "true"))
	}

	return component.H("li",
		component.Class(itemClass),
		component.Children(
			component.H("input", checkboxOpts...),
			component.H("span",
				component.Class("todo-text"),
				component.Children(component.Text(t.todo.Text)),
			),
			component.H("button",
				component.Class("delete"),
				component.On("click", func(e dom.Event) {
					t.app.deleteTodo(t.todo.ID)
				}),
				component.Children(component.Text("x")),
			),
		),
	)
}

type TodoFooter struct {
	app *TodoApp
}

func (t *TodoFooter) Render() component.Node {
	todos := t.app.todos.Get()
	currentFilter := t.app.filter.Peek()

	itemsLeft := 0
	hasCompleted := false
	for _, todo := range todos {
		if todo.Completed {
			hasCompleted = true
			continue
		}
		itemsLeft++
	}

	itemLabel := "items"
	if itemsLeft == 1 {
		itemLabel = "item"
	}

	clearNode := component.H("div")
	if hasCompleted {
		clearNode = component.H("button",
			component.Class("clear-completed"),
			component.On("click", func(e dom.Event) {
				t.app.clearCompleted()
			}),
			component.Children(component.Text("Clear completed")),
		)
	}

	return component.H("footer",
		component.Class("todo-footer"),
		component.Children(
			component.H("span",
				component.Class("todo-count"),
				component.Children(component.Text(itoa(itemsLeft)+" "+itemLabel+" left")),
			),
			component.H("div",
				component.Class("filters"),
				component.Children(
					filterButton(t.app, "All", FilterAll, currentFilter),
					filterButton(t.app, "Active", FilterActive, currentFilter),
					filterButton(t.app, "Completed", FilterCompleted, currentFilter),
				),
			),
			clearNode,
		),
	)
}

func filterButton(app *TodoApp, label string, target Filter, current Filter) component.Node {
	opts := []component.Option{
		component.On("click", func(e dom.Event) {
			app.filter.Set(target)
		}),
		component.Children(component.Text(label)),
	}
	if current == target {
		opts = append(opts, component.Class("active"))
	}
	return component.H("button", opts...)
}

func main() {
	component.Mount("app", NewTodoApp())
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := [20]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte(n%10) + "0"[0]
		n /= 10
	}
	if neg {
		pos--
		buf[pos] = "-"[0]
	}
	return string(buf[pos:])
}
