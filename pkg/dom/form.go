//go:build js && wasm

package dom

// InputValue returns the current string value of an <input> or <textarea>
// element from an "input" or "change" event.
func InputValue(e Event) string {
	return e.Value("target").Get("value").String()
}

// CheckboxChecked returns the checked state of a checkbox input
// from a "change" event.
func CheckboxChecked(e Event) bool {
	return e.Value("target").Get("checked").Bool()
}

// SelectValue returns the selected option value of a <select> element
// from a "change" event.
func SelectValue(e Event) string {
	return e.Value("target").Get("value").String()
}

// InputValueFromElement reads the current value directly from an Element,
// without requiring an event. Useful for reading state on submit.
func InputValueFromElement(el Element) string {
	return el.val.Get("value").String()
}

// CheckboxCheckedFromElement reads the checked property directly from an Element.
func CheckboxCheckedFromElement(el Element) bool {
	return el.val.Get("checked").Bool()
}
