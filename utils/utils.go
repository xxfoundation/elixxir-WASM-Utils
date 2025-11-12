////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package utils

import (
	"fmt"

	"github.com/pkg/errors"
	jww "github.com/spf13/jwalterweatherman"
	"syscall/js"
)

var (
	// JSON is the Javascript JSON type. It is used to perform JSON operations
	// on the Javascript layer.
	JSON = js.Global().Get("JSON")

	// Object is the Javascript Object type. It is used to perform Object
	// operations on the Javascript layer.
	Object = js.Global().Get("Object")

	// Promise is the Javascript Promise type. It is used to generate new
	// promises.
	Promise = js.Global().Get("Promise")

	// Uint8Array is the Javascript Uint8Array type. It is used to create new
	// Uint8Array.
	Uint8Array = js.Global().Get("Uint8Array")
)

// WrapCB wraps a Javascript function in an object so that it can be called
// later with only the arguments and without specifying the function name.
//
// Panics if m is not a function.
func WrapCB(parent js.Value, m string) func(args ...any) js.Value {
	if parent.Get(m).Type() != js.TypeFunction {
		// Create the error separate from the print so stack trace is printed
		err := errors.Errorf("Function %q is not of type %s", m, js.TypeFunction)
		jww.FATAL.Panicf("%+v", err)
	}

	return func(args ...any) js.Value { return parent.Call(m, args...) }
}

// ErrFromPanic converts a recovered panic value into an error.
// Use this in defer/recover blocks to convert panics to errors.
//
// Example:
//
//	func MyFunc() (err error) {
//	    defer func() {
//	        if r := recover(); r != nil {
//	            err = utils.ErrFromPanic(r)
//	        }
//	    }()
//	    // ... code that might panic
//	    return nil
//	}
func ErrFromPanic(r any) error {
	jww.ERROR.Printf("Panic recovered: %+v", r)
	return fmt.Errorf("panic: %v", r)
}

// SafeFunc wraps a Go function to return a JavaScript Promise that properly
// handles both expected errors and unexpected panics.
//
// The wrapped function:
// - Returns a Promise that resolves on success or rejects on error
// - Catches panics via defer/recover and rejects the Promise
// - Runs in a goroutine to avoid blocking the JavaScript event loop
//
// Usage:
//
//	func MyWasmFunc(_ js.Value, args []js.Value) any {
//	    return SafeFunc(func(this js.Value, args []js.Value) (any, error) {
//	        // Your function logic here
//	        result, err := someOperation()
//	        if err != nil {
//	            return nil, err  // Becomes Promise.reject()
//	        }
//	        return result, nil  // Becomes Promise.resolve()
//	    })(js.Value{}, args)
//	}
//
// This replaces the deprecated CreatePromise function, which used resolve/reject
// callbacks and did not include panic recovery.
func SafeFunc(fn func(this js.Value, args []js.Value) (any, error)) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		// Create Promise handler
		handler := js.FuncOf(func(_ js.Value, promiseArgs []js.Value) any {
			resolve := promiseArgs[0]
			reject := promiseArgs[1]

			// Run in goroutine to avoid blocking JavaScript event loop
			go func() {
				// Defer panic recovery
				defer func() {
					if r := recover(); r != nil {
						// Panic occurred - create JavaScript Error and reject Promise
						errorMsg := errors.Errorf("Go panic: %v", r)
						errorConstructor := js.Global().Get("Error")
						errorObject := errorConstructor.New(errorMsg.Error())
						reject.Invoke(errorObject)
					}
				}()

				// Call the actual function
				result, err := fn(this, args)

				if err != nil {
					// Expected error - reject Promise with Error object
					errorConstructor := js.Global().Get("Error")
					errorObject := errorConstructor.New(err.Error())
					reject.Invoke(errorObject)
					return
				}

				// Success - resolve Promise with result
				// Handle nil result (Go 1.25+ doesn't allow js.ValueOf(nil))
				if result == nil {
					resolve.Invoke(js.Undefined())
				} else {
					// Convert result to js.Value if it isn't already
					var jsResult js.Value
					if v, ok := result.(js.Value); ok {
						jsResult = v
					} else {
						jsResult = js.ValueOf(result)
					}
					resolve.Invoke(jsResult)
				}
			}()

			return js.Undefined()
		})

		// Create and return new Promise
		return Promise.New(handler)
	})
}

// Await waits on a Javascript value. It blocks until the awaitable successfully
// resolves to the result or rejects to err.
//
// If there is a result, err will be nil and vice versa.
func Await(awaitable js.Value) (result []js.Value, err []js.Value) {
	then := make(chan []js.Value)
	defer close(then)
	thenFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		then <- args
		return js.Undefined()
	})
	defer thenFunc.Release()

	catch := make(chan []js.Value)
	defer close(catch)
	catchFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		catch <- args
		return js.Undefined()
	})
	defer catchFunc.Release()

	awaitable.Call("then", thenFunc).Call("catch", catchFunc)

	select {
	case result = <-then:
		return result, nil
	case err = <-catch:
		return nil, err
	}
}
