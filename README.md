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
to run, such as running them in a browser. To automate this process get
[wasmbrowsertest](https://github.com/agnivade/wasmbrowsertest) and follow their
[installation instructions](https://github.com/agnivade/wasmbrowsertest#quickstart).
Then, tests can be run using the following command.

```shell
$ GOOS=js GOARCH=wasm go test ./...
```

Note, this will fail because `exception/throw_js.s` contains custom commands
that require our modified `wasm_exec.js` file and wasmbrowsertest does not use
it. To get tests to run, temporarily delete the body of `exception/throw_js.s`
during testing.

## `wasm_exec.js`

`wasm_exec.js` is provided by Go and is used to import the WebAssembly module in
the browser. It can be retrieved from Go using the following command.

```shell
$ cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .
```

Note that this repository makes edits to `wasm_exec.js` and you must either use
the one in this repository or add the following lines in the `go` `importObject`
on `global.Go`.

```javascript
global.Go = class {
    constructor() {
        // ...
        this.importObject = {
            go: {
                // ...
                // func Throw(exception string, message string)
                'gitlab.com/elixxir/wasm-utils/exception.throw': (sp) => {
                    const exception = loadString(sp + 8)
                    const message = loadString(sp + 24)
                    throw globalThis[exception](message)
                },
            }
        }
    }
}
```
