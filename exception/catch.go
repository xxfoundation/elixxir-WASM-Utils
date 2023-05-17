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

// Catch recovers from panics and attempts to convert the value into an error.
// This must be used directly in a defer statement and cannot be called
// elsewhere.
//
// Set err to the address of the return value. This is typically done with a
// named return error value.
//
// Example:
//
//	defer exception.Catch(&err)
func Catch(err *error) {
	if recoverErr := handleRecovery(recover()); recoverErr != nil {
		*err = recoverErr
	}
}

// CatchHandler is the same as [Catch], but enables custom error handling after
// recovering.
func CatchHandler(fn func(err error)) {
	if err := handleRecovery(recover()); err != nil {
		fn(err)
	}
}

func handleRecovery(r interface{}) error {
	if r == nil {
		return nil
	}
	switch val := r.(type) {
	case error:
		return val
	case js.Value:
		return js.Error{Value: val}
	default:
		return fmt.Errorf("%+v", val)
	}
}
