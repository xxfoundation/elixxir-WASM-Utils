////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package storage

import (
	"bytes"
	"github.com/pkg/errors"
	"os"
	"reflect"
	"strconv"
	"syscall/js"
	"testing"
)

// Unit test of GetLocalStorage.
func TestGetLocalStorage(t *testing.T) {
	expected := &localStorage{
		v:      &LocalStorageJS{js.Global().Get("localStorage")},
		prefix: localStorageWasmPrefix,
	}

	ls := GetLocalStorage()

	if !reflect.DeepEqual(expected, ls) {
		t.Errorf("Did not receive expected localStorage."+
			"\nexpected: %+v\nreceived: %+v", expected, ls)
	}
}

// Tests that a value set with localStorage.Set and retrieved with
// localStorage.Get matches the original.
func TestLocalStorage_Get_Set(t *testing.T) {
	values := map[string][]byte{
		"key1": []byte("key value"),
		"key2": {0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		"key3": {0, 49, 0, 0, 0, 38, 249, 93, 242, 189, 222, 32, 138, 248, 121,
			151, 42, 108, 82, 199, 163, 61, 4, 200, 140, 231, 225, 20, 35, 243,
			253, 161, 61, 2, 227, 208, 173, 183, 33, 66, 236, 107, 105, 119, 26,
			42, 44, 60, 109, 172, 38, 47, 220, 17, 129, 4, 234, 241, 141, 81,
			84, 185, 32, 120, 115, 151, 128, 196, 143, 117, 222, 78, 44, 115,
			109, 20, 249, 46, 158, 139, 231, 157, 54, 219, 141, 252},
	}

	for keyName, keyValue := range values {
		err := jsStorage.Set(keyName, keyValue)
		if err != nil {
			t.Errorf("Failed to set %q: %+v", keyName, err)
		}

		loadedValue, err := jsStorage.Get(keyName)
		if err != nil {
			t.Errorf("Failed to load %q: %+v", keyName, err)
		} else if !bytes.Equal(keyValue, loadedValue) {
			t.Errorf("Loaded value does not match original for %q"+
				"\nexpected: %q\nreceived: %q", keyName, keyValue, loadedValue)
		}
	}
}

// Tests that localStorage.Get returns the error os.ErrNotExist when the key
// does not exist in storage.
func TestLocalStorage_Get_NotExistError(t *testing.T) {
	_, err := jsStorage.Get("someKey")
	if err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Incorrect error for non existant key."+
			"\nexpected: %v\nreceived: %v", os.ErrNotExist, err)
	}
}

// Tests that localStorage.RemoveItem deletes a key from the store and that it
// cannot be retrieved.
func TestLocalStorage_RemoveItem(t *testing.T) {
	keyName := "key"
	if err := jsStorage.Set(keyName, []byte("value")); err != nil {
		t.Errorf("Failed to set %q: %+v", keyName, err)
	}
	jsStorage.RemoveItem(keyName)

	_, err := jsStorage.Get(keyName)
	if err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Failed to remove %q: %+v", keyName, err)
	}
}

// Tests that localStorage.Clear deletes all the WASM keys from storage and
// does not remove any others
func TestLocalStorage_Clear(t *testing.T) {
	jsStorage.LocalStorageUNSAFE().Clear()
	const numKeys = 10
	var yesPrefix, noPrefix []string

	for i := 0; i < numKeys; i++ {
		keyName := "keyNum" + strconv.Itoa(i)
		if i%2 == 0 {
			yesPrefix = append(yesPrefix, keyName)
			err := jsStorage.Set(keyName, []byte(strconv.Itoa(i)))
			if err != nil {
				t.Errorf("Failed to set with prefix %q: %+v", keyName, err)
			}
		} else {
			noPrefix = append(noPrefix, keyName)
			err := jsStorage.LocalStorageUNSAFE().SetItem(keyName, strconv.Itoa(i))
			if err != nil {
				t.Errorf("Failed to set with no prefix %q: %+v", keyName, err)
			}
		}
	}

	n := jsStorage.Clear()
	if n != numKeys/2 {
		t.Errorf("Incorrect number of keys.\nexpected: %d\nreceived: %d",
			numKeys/2, n)
	}

	for _, keyName := range noPrefix {
		if _, err := jsStorage.LocalStorageUNSAFE().GetItem(keyName); err != nil {
			t.Errorf("Could not get keyName %q: %+v", keyName, err)
		}
	}
	for _, keyName := range yesPrefix {
		keyValue, err := jsStorage.Get(keyName)
		if err == nil || !errors.Is(err, os.ErrNotExist) {
			t.Errorf("Found keyName %q: %q", keyName, keyValue)
		}
	}
}

// Tests that localStorage.ClearPrefix deletes only the keys with the given
// prefix.
func TestLocalStorage_ClearPrefix(t *testing.T) {
	jsStorage.LocalStorageUNSAFE().Clear()
	const numKeys = 10
	var yesPrefix, noPrefix []string
	prefix := "keyNamePrefix/"

	for i := 0; i < numKeys; i++ {
		keyName := "keyNum " + strconv.Itoa(i)
		if i%2 == 0 {
			keyName = prefix + keyName
			yesPrefix = append(yesPrefix, keyName)
		} else {
			noPrefix = append(noPrefix, keyName)
		}

		if err := jsStorage.Set(keyName, []byte(strconv.Itoa(i))); err != nil {
			t.Errorf("Failed to set %q: %+v", keyName, err)
		}
	}

	n := jsStorage.ClearPrefix(prefix)
	if n != numKeys/2 {
		t.Errorf("Incorrect number of keys.\nexpected: %d\nreceived: %d",
			numKeys/2, n)
	}

	for _, keyName := range noPrefix {
		if _, err := jsStorage.Get(keyName); err != nil {
			t.Errorf("Could not get keyName %q: %+v", keyName, err)
		}
	}
	for _, keyName := range yesPrefix {
		keyValue, err := jsStorage.Get(keyName)
		if err == nil || !errors.Is(err, os.ErrNotExist) {
			t.Errorf("Found keyName %q: %q", keyName, keyValue)
		}
	}
}

// Tests that localStorage.Key return all added keys when looping through all
// indexes.
func TestLocalStorage_Key(t *testing.T) {
	jsStorage.LocalStorageUNSAFE().Clear()
	values := map[string][]byte{
		"key1": []byte("key value"),
		"key2": {0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		"key3": {0, 49, 0, 0, 0, 38, 249, 93},
	}

	for keyName, keyValue := range values {
		if err := jsStorage.Set(keyName, keyValue); err != nil {
			t.Errorf("Failed to set %q: %+v", keyName, err)
		}
	}

	numKeys := len(values)
	for i := 0; i < numKeys; i++ {
		keyName, err := jsStorage.Key(i)
		if err != nil {
			t.Errorf("No key found for index %d: %+v", i, err)
		}

		if _, exists := values[keyName]; !exists {
			t.Errorf("No key with name %q added to storage.", keyName)
		}
		delete(values, keyName)
	}

	if len(values) != 0 {
		t.Errorf("%d keys not read from storage: %q", len(values), values)
	}
}

// Tests that localStorage.Key returns the error os.ErrNotExist when the index
// is greater than or equal to the number of keys.
func TestLocalStorage_Key_NotExistError(t *testing.T) {
	jsStorage.LocalStorageUNSAFE().Clear()
	if err := jsStorage.Set("key", []byte("value")); err != nil {
		t.Errorf("Failed to set: %+v", err)
	}

	_, err := jsStorage.Key(1)
	if err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Incorrect error for non existant key index."+
			"\nexpected: %v\nreceived: %v", os.ErrNotExist, err)
	}

	_, err = jsStorage.Key(2)
	if err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Incorrect error for non existant key index."+
			"\nexpected: %v\nreceived: %v", os.ErrNotExist, err)
	}
}

// Tests that localStorage.Length returns the correct Length when adding and
// removing various keys.
func TestLocalStorage_Length(t *testing.T) {
	jsStorage.LocalStorageUNSAFE().Clear()
	values := map[string][]byte{
		"key1": []byte("key value"),
		"key2": {0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		"key3": {0, 49, 0, 0, 0, 38, 249, 93},
	}

	i := 0
	for keyName, keyValue := range values {
		if err := jsStorage.Set(keyName, keyValue); err != nil {
			t.Errorf("Failed to set %q: %+v", keyName, err)
		}
		i++

		if jsStorage.Length() != i {
			t.Errorf("Incorrect length.\nexpected: %d\nreceived: %d",
				i, jsStorage.Length())
		}
	}

	i = len(values)
	for keyName := range values {
		jsStorage.RemoveItem(keyName)
		i--

		if jsStorage.Length() != i {
			t.Errorf("Incorrect length.\nexpected: %d\nreceived: %d",
				i, jsStorage.Length())
		}
	}
}

// Tests that localStorage.Keys return a list that contains all the added keys.
func TestLocalStorage_Keys(t *testing.T) {
	jsStorage.LocalStorageUNSAFE().Clear()
	values := map[string][]byte{
		"key1": []byte("key value"),
		"key2": {0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		"key3": {0, 49, 0, 0, 0, 38, 249, 93},
	}

	for keyName, keyValue := range values {
		if err := jsStorage.Set(keyName, keyValue); err != nil {
			t.Errorf("Failed to set %q: %+v", keyName, err)
		}
	}

	keys := jsStorage.Keys()
	for i, keyName := range keys {
		if _, exists := values[keyName]; !exists {
			t.Errorf("Key %q does not exist (%d).", keyName, i)
		}
	}
}
