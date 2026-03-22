.PHONY: test test-reactive test-dom test-component test-all

# Run reactive tests (no browser needed)
test-reactive:
	go test -v ./pkg/reactive/...

# Run dom tests in headless Chrome via wasmbrowsertest
test-dom:
	GOOS=js GOARCH=wasm go test -v \
		-exec="$(shell go env GOPATH)/bin/wasmbrowsertest" \
		./pkg/dom/...

# Run component tests in headless Chrome
test-component:
	GOOS=js GOARCH=wasm go test -v \
		-exec="$(shell go env GOPATH)/bin/wasmbrowsertest" \
		./pkg/component/...

# Run all tests
test-all: test-reactive test-dom test-component

# Default target
test: test-all
