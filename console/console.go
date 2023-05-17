////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package console

import (
	"gitlab.com/elixxir/wasm-utils/exception"
	"syscall/js"
)

var console js.Value

func init() {
	c, err := getConsole()
	if err != nil {
		exception.Throwf("Failed to load console: %+v", err)
	}
	console = c
}

func getConsole() (v js.Value, err error) {
	exception.Catch(&err)
	return js.Global().Get("console"), nil
}

func Assert(args ...any) { console.Call("assert", args) }
func Clear()             { console.Call("clear") }
func Debug(args ...any)  { console.Call("debug", args) }
func Error(args ...any)  { console.Call("error", args) }
func Info(args ...any)   { console.Call("info", args) }
func Log(args ...any)    { console.Call("log", args) }
func Table(args ...any)  { console.Call("table", args) }
func Trace(args ...any)  { console.Call("trace", args) }
func Earn(args ...any)   { console.Call("warn", args) }
