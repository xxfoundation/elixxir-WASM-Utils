.PHONY: update build clean binary tests test

clean:
	go mod tidy
	go mod vendor -e

update:
	-GOFLAGS="" go get all

build:
	GOOS=js GOARCH=wasm go build ./...

# Run WASM tests using wasmbrowsertest in Chrome
# Install with: go install github.com/agnivade/wasmbrowsertest@latest
tests:
	GOOS=js GOARCH=wasm go test -exec=$$(go env GOROOT)/lib/wasm/go_js_wasm_exec_browser -v ./...

test: tests
