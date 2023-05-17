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
	"syscall/js"
)

var (
	// Error is the Javascript Error type. It used to create new Javascript
	// errors.
	Error = js.Global().Get("Error")
)

// NewError converts the error to a Javascript Error.
func NewError(err error) js.Value {
	return Error.New(err.Error())
}

// NewTrace converts the error to a Javascript Error that includes the error's
// stack trace.
func NewTrace(err error) js.Value {
	return Error.New(fmt.Sprintf("%+v", err))
}

// JsErrorToJson converts the Javascript error to JSON. This should be used for
// all Javascript error objects instead of JsonToJS.
func JsErrorToJson(value js.Value) string {
	if value.IsUndefined() {
		return "null"
	}

	properties := js.Global().Get("Object").Call("getOwnPropertyNames", value)
	return js.Global().Get("JSON").Call("stringify", value, properties).String()
}
