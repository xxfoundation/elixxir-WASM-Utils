////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package exception

import (
	"errors"
	"syscall/js"
	"testing"
)

func TestCatch(t *testing.T) {
	t.Parallel()

	t.Run("no error and no panic", func(t *testing.T) {
		t.Parallel()
		resultErr := func() (err error) {
			defer Catch(&err)
			// no-op
			return nil
		}()
		if resultErr != nil {
			t.Error(resultErr)
		}
	})

	t.Run("error and no panic", func(t *testing.T) {
		t.Parallel()
		someErr := errors.New("my error")
		resultErr := func() (err error) {
			defer Catch(&err)
			// no-op
			return someErr
		}()
		if resultErr == nil || !errors.Is(someErr, resultErr) {
			t.Errorf("Unexpected error.\nexpected: %v\nreceived: %v",
				someErr, resultErr)
		}
	})

	t.Run("panic with error", func(t *testing.T) {
		t.Parallel()
		someErr := errors.New("some error")
		resultErr := func() (err error) {
			defer Catch(&err)
			panic(someErr)
		}()
		if resultErr == nil || !errors.Is(someErr, resultErr) {
			t.Errorf("Unexpected error.\nexpected: %v\nreceived: %v",
				someErr, resultErr)
		}
	})

	t.Run("panic with js.Value", func(t *testing.T) {
		t.Parallel()
		someErr := testJSErrValue()
		resultErr := func() (err error) {
			defer Catch(&err)
			panic(someErr)
		}()
		expectedErr := js.Error{Value: someErr}
		if resultErr == nil || errors.Is(expectedErr, resultErr) {
			t.Errorf("Unexpected error.\nexpected: %v\nreceived: %v",
				expectedErr, resultErr)
		}
	})

	t.Run("panic with other type", func(t *testing.T) {
		t.Parallel()
		someErr := "some other type"
		resultErr := func() (err error) {
			defer Catch(&err)
			panic(someErr)
		}()
		if resultErr == nil || someErr != resultErr.Error() {
			t.Errorf("Unexpected error.\nexpected: %s\nreceived: %v",
				someErr, resultErr)
		}
	})
}

func testJSErrValue() (value js.Value) {
	defer func() {
		recoverVal := recover()
		value = recoverVal.(js.Error).Value
	}()
	js.Global().Get("Function").New(`throw Exception("some error")`).Invoke()
	panic("not a JS value. line above should do the panic")
}

func TestCatchHandler(t *testing.T) {
	t.Parallel()
	var calledHandler bool
	resultErr := func() (err error) {
		defer CatchHandler(func(handlerErr error) {
			calledHandler = true
			err = handlerErr
		})
		panic("some error")
	}()
	if resultErr == nil || "some error" != resultErr.Error() {
		t.Errorf("Unexpected error.\nexpected: %s\nreceived: %v",
			"some error", resultErr)
	}
	if calledHandler != true {
		t.Errorf("Unexpected calledHandler.\nexpected: %t\neceived: %t",
			true, calledHandler)
	}
}
