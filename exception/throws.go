////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

package exception

// This file contains the stub for the throw function, which is linked via
// assembly in throw_js.s to a custom function added to wasm_exec.js that throws
// the passed elements. This adds the ability to throw a Javascript exception
// from the go webassembly.
//
// Testing uses the [wasmbrowsertest], which uses the default wasm_exec.js file,
// which causes compile-time errors. To avoid this, throw_js.s must be cleared
// out and the stub below must be replaced with an actual function, as shown
// below.
//
//  func throw(exception string, message string) {}
//
// To make it easier, you can also replace throws.go temporarily with throws.dev
// while running tests, which replaces the stub with a function that panics.
//
// You can also add the following lines to your Makefile.
//
//  wasmException = "vendor/gitlab.com/elixxir/wasm-utils/exception"
//
//  wasm_tests:
//      cp $(wasmException)/throw_js.s $(wasmException)/throw_js.s.bak
//      cp $(wasmException)/throws.go $(wasmException)/throws.go.bak
//      > $(wasmException)/throw_js.s
//      cp $(wasmException)/throws.dev $(wasmException)/throws.go
//      -GOOS=js GOARCH=wasm go test -v ./...
//      mv $(wasmException)/throw_js.s.bak $(wasmException)/throw_js.s
//      mv $(wasmException)/throws.go.bak $(wasmException)/throws.go
//
// [wasmbrowsertest]: https://github.com/agnivade/wasmbrowsertest

// throw is a function stub that connects to the bindings in wasm_exec.js to
// allow throwing exceptions.
func throw(exception string, message string)
