# WASM Utils

This repository contains utilities for interfacing with Javascript.

## Building

The repository can only be compiled to a WebAssembly binary using `GOOS=js` and
`GOARCH=wasm`.

```shell
$ GOOS=js GOARCH=wasm go build -o xxdk.wasm
```

### Running Unit Tests

This repository depends on `syscall/js`, which requires a Javascript environment
to run, such as running them in a browser. To automate this process, get
[wasmbrowsertest](https://github.com/agnivade/wasmbrowsertest) and follow their
[installation instructions](https://github.com/agnivade/wasmbrowsertest#quickstart).
Then, tests can be run using the following command.

```shell
$ GOOS=js GOARCH=wasm go test ./...
```

## Breaking Changes in v1.0.0

### Removed `exception` package

The `exception` package has been removed. It is no longer necessary with the new
`SafeFunc` wrapper, which provides automatic panic recovery and cleaner error handling.

**Migration Guide:**

Replace:
```go
func MyFunc(_ js.Value, args []js.Value) any {
    result, err := someOperation()
    if err != nil {
        exception.ThrowTrace(err)
        return nil
    }
    return result
}
```

With:
```go
func MyFunc(_ js.Value, args []js.Value) any {
    return utils.SafeFunc(func(this js.Value, args []js.Value) (any, error) {
        result, err := someOperation()
        if err != nil {
            return nil, err
        }
        return result, nil
    })(js.Value{}, args)
}
```

For internal functions that need panic recovery without exposing to JavaScript:
```go
func MyInternalFunc() (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = utils.ErrFromPanic(r)
        }
    }()
    // ... code that might panic
    return nil
}
```

### Removed custom `wasm_exec.js`

Use the standard `wasm_exec.js` from Go:

```shell
$ cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" .
```

The custom modifications are no longer needed since `SafeFunc` handles error
propagation through standard Promise rejection.

### Deprecated `CreatePromise`

The `CreatePromise` function has been removed in favor of `SafeFunc`, which provides:
- Automatic panic recovery
- Cleaner API with `(any, error)` returns instead of `resolve/reject` callbacks
- Better error handling through Promise rejection

See the `SafeFunc` documentation in `utils/utils.go` for usage examples.
