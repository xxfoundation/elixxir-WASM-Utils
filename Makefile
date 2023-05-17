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
	GOOS=js GOARCH=wasm go test -v ./...
