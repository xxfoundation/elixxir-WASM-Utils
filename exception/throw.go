////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package exception

import (
	"fmt"
	"unsafe"
)

var errStr string = "Error"

// Throw creates a Javascript Error object from a Go error and throws it as an
// exception.
func Throw(err error) {
	msg := err.Error()
	throw(unsafe.Pointer(&errStr), unsafe.Pointer(&msg))
}

// Throwf formats according to a format specifier, creates a Javascript Error
// object, and throws it as an exception.
func Throwf(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	throw(unsafe.Pointer(&errStr),
		unsafe.Pointer(&msg))
}

// ThrowTrace creates a Javascript Error object from a Go error and throws it as
// an exception. The error includes its stack trace.
func ThrowTrace(err error) {
	msg := fmt.Sprintf("%+v", err)
	throw(unsafe.Pointer(&errStr),
		unsafe.Pointer(&msg))
}
