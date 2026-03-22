package dom

import "github.com/seanrobmerriam/gowasm/pkg/reactive"

// BindInput returns an EventHandler that sets signal to the input's current
// value on every "input" event. Use with component.On("input", BindInput(sig)).
func BindInput(sig *reactive.Signal[string]) EventHandler {
	return func(e Event) {
		sig.Set(InputValue(e))
	}
}

// BindCheckbox returns an EventHandler that sets signal to the checkbox's
// checked state on every "change" event.
func BindCheckbox(sig *reactive.Signal[bool]) EventHandler {
	return func(e Event) {
		sig.Set(CheckboxChecked(e))
	}
}

// BindSelect returns an EventHandler that sets signal to the select's current
// value on every "change" event.
func BindSelect(sig *reactive.Signal[string]) EventHandler {
	return func(e Event) {
		sig.Set(SelectValue(e))
	}
}
