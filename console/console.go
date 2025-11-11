////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package console

import (
	"syscall/js"
)

var console js.Value

func init() {
	console = js.Global().Get("console")
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
