package component

import (
	"github.com/seanrobmerriam/gowasm/pkg/dom"
	"github.com/seanrobmerriam/gowasm/pkg/reactive"
)

// InputProps configures a controlled text input.
type InputProps struct {
	Value       *reactive.Signal[string] // required — drives the input value
	Placeholder string
	Type        string // "text", "email", "password", etc. default: "text"
	Class       string
	ID          string
	Disabled    bool
	OnInput     dom.EventHandler // optional additional handler called after signal update
	OnChange    dom.EventHandler // optional additional handler
	OnKeyDown   dom.EventHandler // optional
}

// Input returns a controlled <input> element node.
// The input's value is driven by props.Value (a Signal[string]).
// User typing updates the signal via two-way binding.
func Input(props InputProps) Node {
	inputType := props.Type
	if inputType == "" {
		inputType = "text"
	}

	opts := []Option{
		Attr("type", inputType),
		Attr("value", props.Value.Get()),
	}

	if props.Placeholder != "" {
		opts = append(opts, Attr("placeholder", props.Placeholder))
	}
	if props.Class != "" {
		opts = append(opts, Attr("class", props.Class))
	}
	if props.ID != "" {
		opts = append(opts, Attr("id", props.ID))
	}
	if props.Disabled {
		opts = append(opts, Attr("disabled", "true"))
	}

	// Combine the binding handler with any user-provided handler
	var inputHandler dom.EventHandler
	if props.OnInput != nil {
		wrapper := props.OnInput
		inputHandler = func(e dom.Event) {
			dom.BindInput(props.Value)(e)
			wrapper(e)
		}
	} else {
		inputHandler = dom.BindInput(props.Value)
	}
	opts = append(opts, On("input", inputHandler))

	if props.OnChange != nil {
		opts = append(opts, On("change", props.OnChange))
	}
	if props.OnKeyDown != nil {
		opts = append(opts, On("keydown", props.OnKeyDown))
	}

	return H("input", opts...)
}

// TextareaProps configures a controlled textarea.
type TextareaProps struct {
	Value       *reactive.Signal[string]
	Placeholder string
	Rows        int
	Class       string
	ID          string
	Disabled    bool
	OnInput     dom.EventHandler
}

// Textarea returns a controlled <textarea> element node.
func Textarea(props TextareaProps) Node {
	opts := []Option{
		Attr("value", props.Value.Get()),
	}

	if props.Placeholder != "" {
		opts = append(opts, Attr("placeholder", props.Placeholder))
	}
	if props.Rows > 0 {
		opts = append(opts, Attr("rows", itoa(props.Rows)))
	}
	if props.Class != "" {
		opts = append(opts, Attr("class", props.Class))
	}
	if props.ID != "" {
		opts = append(opts, Attr("id", props.ID))
	}
	if props.Disabled {
		opts = append(opts, Attr("disabled", "true"))
	}

	var inputHandler dom.EventHandler
	if props.OnInput != nil {
		wrapper := props.OnInput
		inputHandler = func(e dom.Event) {
			dom.BindInput(props.Value)(e)
			wrapper(e)
		}
	} else {
		inputHandler = dom.BindInput(props.Value)
	}
	opts = append(opts, On("input", inputHandler))

	return H("textarea", opts...)
}

// CheckboxProps configures a controlled checkbox input.
type CheckboxProps struct {
	Checked  *reactive.Signal[bool]
	Label    string // if non-empty, wraps in <label>
	Class    string
	ID       string
	Disabled bool
	OnChange dom.EventHandler
}

// Checkbox returns a controlled <input type="checkbox"> node.
// If props.Label is non-empty, wraps the input in a <label> element.
func Checkbox(props CheckboxProps) Node {
	opts := []Option{
		Attr("type", "checkbox"),
	}

	if props.Checked.Get() {
		opts = append(opts, Attr("checked", "true"))
	}
	if props.Class != "" {
		opts = append(opts, Attr("class", props.Class))
	}
	if props.ID != "" {
		opts = append(opts, Attr("id", props.ID))
	}
	if props.Disabled {
		opts = append(opts, Attr("disabled", "true"))
	}

	var changeHandler dom.EventHandler
	if props.OnChange != nil {
		wrapper := props.OnChange
		changeHandler = func(e dom.Event) {
			dom.BindCheckbox(props.Checked)(e)
			wrapper(e)
		}
	} else {
		changeHandler = dom.BindCheckbox(props.Checked)
	}
	opts = append(opts, On("change", changeHandler))

	inputNode := H("input", opts...)

	if props.Label != "" {
		return H("label", nil, Children(inputNode, Text(props.Label)))
	}
	return inputNode
}

// SelectOption represents one option in a select.
type SelectOption struct {
	Value string
	Label string
}

// SelectProps configures a controlled select element.
type SelectProps struct {
	Value    *reactive.Signal[string] // current selected value
	Options  []SelectOption
	Class    string
	ID       string
	Disabled bool
	OnChange dom.EventHandler
}

// Select returns a controlled <select> element node.
// The selected option is driven by props.Value.
func Select(props SelectProps) Node {
	currentValue := props.Value.Get()

	opts := []Option{}

	if props.Class != "" {
		opts = append(opts, Attr("class", props.Class))
	}
	if props.ID != "" {
		opts = append(opts, Attr("id", props.ID))
	}
	if props.Disabled {
		opts = append(opts, Attr("disabled", "true"))
	}

	var changeHandler dom.EventHandler
	if props.OnChange != nil {
		wrapper := props.OnChange
		changeHandler = func(e dom.Event) {
			dom.BindSelect(props.Value)(e)
			wrapper(e)
		}
	} else {
		changeHandler = dom.BindSelect(props.Value)
	}
	opts = append(opts, On("change", changeHandler))

	// Build option children
	var children []Node
	for _, opt := range props.Options {
		optNode := H("option", Attr("value", opt.Value))
		if opt.Value == currentValue {
			optNode.attrs["selected"] = "true"
		}
		optNode.children = append(optNode.children, Text(opt.Label))
		children = append(children, optNode)
	}

	return H("select", append(opts, Children(children...))...)
}

// itoa converts an int to string.
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
		buf[pos] = byte(n%10) + '0'
		n /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
