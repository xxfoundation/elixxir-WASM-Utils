////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package exception

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"sort"
	"strings"
	"syscall/js"
	"testing"
)

// Tests that TestNewError returns a Javascript Error object with the expected
// message.
func TestNewError(t *testing.T) {
	err := errors.New("test error")
	expectedErr := err.Error()
	newError := NewError(err).Get("message").String()

	if newError != expectedErr {
		t.Errorf("Failed to get expected error message."+
			"\nexpected: %s\nreceived: %s", expectedErr, newError)
	}
}

// Tests that TestNewTrace returns a Javascript Error object with the expected
// message and stack trace.
func TestNewTrace(t *testing.T) {
	err := errors.New("test error")
	expectedErr := fmt.Sprintf("%+v", err)
	newError := NewTrace(err).Get("message").String()

	if newError != expectedErr {
		t.Errorf("Failed to get expected error message."+
			"\nexpected: %s\nreceived: %s", expectedErr, newError)
	}
}

// Tests that JsErrorToJson can convert a Javascript object to JSON that matches
// the output of json.Marshal on the Go version of the same object.
func TestJsErrorToJson(t *testing.T) {
	testObj := map[string]any{
		"nil":    nil,
		"bool":   true,
		"int":    1,
		"float":  1.5,
		"string": "I am string",
		"array":  []any{1, 2, 3},
		"object": map[string]any{"int": 5},
	}

	expected, err := json.Marshal(testObj)
	if err != nil {
		t.Errorf("Failed to JSON marshal test object: %+v", err)
	}

	jsJson := JsErrorToJson(js.ValueOf(testObj))

	// Javascript does not return the JSON object fields sorted so the letters
	// of each Javascript string are sorted and compared
	er := []rune(string(expected))
	sort.SliceStable(er, func(i, j int) bool { return er[i] < er[j] })
	jj := []rune(jsJson)
	sort.SliceStable(jj, func(i, j int) bool { return jj[i] < jj[j] })

	if string(er) != string(jj) {
		t.Errorf("Recieved incorrect JSON from Javascript object."+
			"\nexpected: %s\nreceived: %s", expected, jsJson)
	}
}

// Tests that JsErrorToJson return a null object when the Javascript object is
// undefined.
func TestJsErrorToJson_Undefined(t *testing.T) {
	expected, err := json.Marshal(nil)
	if err != nil {
		t.Errorf("Failed to JSON marshal test object: %+v", err)
	}

	jsJson := JsErrorToJson(js.Undefined())

	if string(expected) != jsJson {
		t.Errorf("Recieved incorrect JSON from Javascript object."+
			"\nexpected: %s\nreceived: %s", expected, jsJson)
	}
}

// Tests that JsErrorToJson returns a JSON object containing the original error
// string.
func TestJsErrorToJson_ErrorObject(t *testing.T) {
	expected := "An error"
	jsErr := Error.New(expected)
	jsJson := JsErrorToJson(jsErr)

	if !strings.Contains(jsJson, expected) {
		t.Errorf("Recieved incorrect JSON from Javascript error."+
			"\nexpected: %s\nreceived: %s", expected, jsJson)
	}
}
