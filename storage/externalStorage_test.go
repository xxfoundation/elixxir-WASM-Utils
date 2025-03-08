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
	"os"
	"reflect"
	"strconv"
	"syscall/js"
	"testing"

	"github.com/pkg/errors"
)

// setupMockHavenStorage sets up a mock implementation of havenStorage in the JavaScript environment
func setupMockHavenStorage(_ *testing.T) {
	const setupScript = `
	const havenStorage = {
		getItem: function(key) {
			return Promise.resolve(localStorage.getItem(key));
		},

		setItem: function(key, value) {
			localStorage.setItem(key, value);
			return Promise.resolve();
		},

		delete: function(key) {
			localStorage.removeItem(key);
			return Promise.resolve();
		},

		clear: function() {
			localStorage.clear();
			return Promise.resolve();
		},

		getKeys: function() {
			return Promise.resolve(Object.keys(localStorage));
		},

		key: function(index) {
			// To Test unimplemented error
			return Promise.reject(new Error('not implemented'));
		}
	};

	// Initialize mock storage
	window._mockHavenStorage = {};
	window.havenStorage = havenStorage;
	`
	js.Global().Call("eval", setupScript)
	jsExternalStorage = newExternalStorage(externalStorageWasmPrefix)
}

// Unit test of GetExternalStorage.
func TestGetExternalStorage(t *testing.T) {
	setupMockHavenStorage(t)

	expected := &externalStorage{
		v:      &HavenStorageJS{js.Global().Get("havenStorage")},
		prefix: externalStorageWasmPrefix,
	}

	es := GetExternalStorage()

	if !reflect.DeepEqual(expected, es) {
		t.Errorf("Did not receive expected externalStorage."+
			"\nexpected: %+v\nreceived: %+v", expected, es)
	}
}

// Tests that a value set with externalStorage.Set and retrieved with
// externalStorage.Get matches the original.
func TestExternalStorage_Get_Set(t *testing.T) {
	setupMockHavenStorage(t)

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
		err := jsExternalStorage.Set(keyName, keyValue)
		if err != nil {
			t.Errorf("Failed to set %q: %+v", keyName, err)
		}

		loadedValue, err := jsExternalStorage.Get(keyName)
		if err != nil {
			t.Errorf("Failed to load %q: %+v", keyName, err)
		} else if !bytes.Equal(keyValue, loadedValue) {
			t.Errorf("Loaded value does not match original for %q"+
				"\nexpected: %q\nreceived: %q", keyName, keyValue, loadedValue)
		}
	}
}

// Tests that externalStorage.Delete deletes a key from the store and that it
// cannot be retrieved.
func TestExternalStorage_Delete(t *testing.T) {
	setupMockHavenStorage(t)

	keyName := "key"
	if err := jsExternalStorage.Set(keyName, []byte("value")); err != nil {
		t.Errorf("Failed to set %q: %+v", keyName, err)
	}

	err := jsExternalStorage.Delete(keyName)
	if err != nil {
		t.Errorf("Failed to delete key %q: %+v", keyName, err)
	}

	_, err = jsExternalStorage.Get(keyName)
	if err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Failed to remove %q: %+v", keyName, err)
	}
}

// Tests that externalStorage.Clear deletes all the WASM keys from storage and
// does not remove any others
func TestExternalStorage_Clear(t *testing.T) {
	setupMockHavenStorage(t)

	// Clear any existing data
	jsExternalStorage.ExternalStorageUNSAFE().Clear()

	const numKeys = 10
	var yesPrefix, noPrefix []string

	for i := 0; i < numKeys; i++ {
		keyName := "keyNum" + strconv.Itoa(i)
		if i%2 == 0 {
			yesPrefix = append(yesPrefix, keyName)
			err := jsExternalStorage.Set(keyName, []byte(strconv.Itoa(i)))
			if err != nil {
				t.Errorf("Failed to set with prefix %q: %+v", keyName, err)
			}
		} else {
			noPrefix = append(noPrefix, keyName)
			err := jsExternalStorage.ExternalStorageUNSAFE().SetItem(keyName, strconv.Itoa(i))
			if err != nil {
				t.Errorf("Failed to set with no prefix %q: %+v", keyName, err)
			}
		}
	}

	n, err := jsExternalStorage.Clear()
	if err != nil {
		t.Errorf("Failed to clear storage: %+v", err)
	}
	if n != numKeys/2 {
		t.Errorf("Incorrect number of keys.\nexpected: %d\nreceived: %d",
			numKeys/2, n)
	}

	for _, keyName := range noPrefix {
		if _, err := jsExternalStorage.ExternalStorageUNSAFE().GetItem(keyName); err != nil {
			t.Errorf("Could not get keyName %q: %+v", keyName, err)
		}
	}
	for _, keyName := range yesPrefix {
		keyValue, err := jsExternalStorage.Get(keyName)
		if err == nil || !errors.Is(err, os.ErrNotExist) {
			t.Errorf("Found keyName %q: %q", keyName, keyValue)
		}
	}
}

// Tests that externalStorage.ClearPrefix deletes only the keys with the given
// prefix.
func TestExternalStorage_ClearPrefix(t *testing.T) {
	setupMockHavenStorage(t)

	// Clear any existing data
	jsExternalStorage.ExternalStorageUNSAFE().Clear()

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

		if err := jsExternalStorage.Set(keyName, []byte(strconv.Itoa(i))); err != nil {
			t.Errorf("Failed to set %q: %+v", keyName, err)
		}
	}

	n, err := jsExternalStorage.ClearPrefix(prefix)
	if err != nil {
		t.Errorf("Failed to clear prefix: %+v", err)
	}
	if n != numKeys/2 {
		t.Errorf("Incorrect number of keys.\nexpected: %d\nreceived: %d",
			numKeys/2, n)
	}

	for _, keyName := range noPrefix {
		if _, err := jsExternalStorage.Get(keyName); err != nil {
			t.Errorf("Could not get keyName %q: %+v", keyName, err)
		}
	}
	for _, keyName := range yesPrefix {
		keyValue, err := jsExternalStorage.Get(keyName)
		if err == nil || !errors.Is(err, os.ErrNotExist) {
			t.Errorf("Found keyName %q: %q", keyName, keyValue)
		}
	}
}

// Tests that externalStorage.Key return all added keys when looping through all
// indexes.
func TestExternalStorage_Key(t *testing.T) {
	setupMockHavenStorage(t)

	// This test will verify that the UnimplementedErr is properly returned
	// since the mock implementation rejects the key method
	_, err := jsExternalStorage.Key(0)
	if err == nil || !errors.Is(err, UnimplementedErr) {
		t.Errorf("Expected UnimplementedErr for Key method.\nexpected: %v\nreceived: %v",
			UnimplementedErr, err)
	}
}

// Tests that externalStorage.Get returns the error os.ErrNotExist when the key
// does not exist in storage.
func TestExternalStorage_Get_NotExistError(t *testing.T) {
	setupMockHavenStorage(t)

	_, err := jsExternalStorage.Get("someKey")
	if err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Incorrect error for non existant key."+
			"\nexpected: %v\nreceived: %v", os.ErrNotExist, err)
	}
}

// Tests that externalStorage.Keys return a list that contains all the added keys.
func TestExternalStorage_Keys(t *testing.T) {
	setupMockHavenStorage(t)

	// Clear any existing data
	jsExternalStorage.ExternalStorageUNSAFE().Clear()

	values := map[string][]byte{
		"key1": []byte("key value"),
		"key2": {0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		"key3": {0, 49, 0, 0, 0, 38, 249, 93},
	}

	for keyName, keyValue := range values {
		if err := jsExternalStorage.Set(keyName, keyValue); err != nil {
			t.Errorf("Failed to set %q: %+v", keyName, err)
		}
	}

	keys, err := jsExternalStorage.Keys()
	if err != nil {
		t.Errorf("Failed to get keys: %+v", err)
	}

	if len(keys) != len(values) {
		t.Errorf("Incorrect number of keys.\nexpected: %d\nreceived: %d",
			len(values), len(keys))
	}

	for i, keyName := range keys {
		if _, exists := values[keyName]; !exists {
			t.Errorf("Key %q does not exist (%d).", keyName, i)
		}
	}
}
