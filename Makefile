.PHONY: update build clean binary tests

clean:
	go mod tidy
	go mod vendor -e

update:
	-GOFLAGS="" go get all

build:
	GOOS=js GOARCH=wasm go build ./...

binary:
	GOOS=js GOARCH=wasm go build -ldflags '-w -s' -trimpath -o wasm-utils.wasm main.go

tests:
	cp exception/throw_js.s exception/throw_js.s.bak
	> exception/throw_js.s
	-GOOS=js GOARCH=wasm go test -cover -v ./storage/...
	mv exception/throw_js.s.bak exception/throw_js.s
