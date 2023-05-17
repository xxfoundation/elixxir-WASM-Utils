////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package exception

import "fmt"

// Throw creates a Javascript Error object from a Go error and throws it as an
// exception.
func Throw(err error) {
	throw("Error", err.Error())
}

// Throwf formats according to a format specifier, creates a Javascript Error
// object, and throws it as an exception.
func Throwf(format string, a ...any) {
	throw("Error", fmt.Sprintf(format, a...))
}

// ThrowTrace creates a Javascript Error object from a Go error and throws it as
// an exception. The error includes its stack trace.
func ThrowTrace(err error) {
	throw("Error", fmt.Sprintf("%+v", err))
}

// throw is a function stub that connects to the bindings in wasm_exec.js to
// allow throwing exceptions.
func throw(exception string, message string)
