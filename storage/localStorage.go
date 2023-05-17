////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package storage

import (
	"encoding/base64"
	"os"
	"strings"
	"syscall/js"

	// "github.com/Max-Sum/base32768"

	"gitlab.com/elixxir/wasm-utils/exception"
	"gitlab.com/elixxir/wasm-utils/utils"
)

// localStorageWasmPrefix is prefixed to every keyName saved to local storage by
// LocalStorage. It allows the identifications and deletion of keys only created
// by this WASM binary while ignoring keys made by other scripts on the same
// page.
//
// The chosen prefix is two characters, that when converted to UTF16, take up 4
// bytes without any zeros to make them more unique.
const localStorageWasmPrefix = "xxdkWasmStorage/"

// const localStorageWasmPrefix = "🞮🞮"

// LocalStorage contains the js.Value representation of localStorage.
type LocalStorage struct {
	// The Javascript value containing the localStorage object
	v *LocalStorageJS

	// The prefix appended to each key name. This is so that all keys created by
	// this structure can be deleted without affecting other keys in local
	// storage.
	prefix string
}

// jsStorage is the global that stores Javascript as window.localStorage.
//
// Doc: https://developer.mozilla.org/en-US/docs/Web/API/Window/localStorage
var jsStorage = newLocalStorage(localStorageWasmPrefix)

// newLocalStorage creates a new LocalStorage object with the specified prefix.
func newLocalStorage(prefix string) *LocalStorage {
	return &LocalStorage{
		v:      &LocalStorageJS{js.Global().Get("localStorage")},
		prefix: prefix,
	}
}

// GetLocalStorage returns Javascript's local storage.
func GetLocalStorage() *LocalStorage {
	return jsStorage
}

// Get decodes and returns the value from the local storage given its key
// name. Returns os.ErrNotExist if the key does not exist.
func (ls *LocalStorage) Get(keyName string) ([]byte, error) {
	value, err := ls.v.GetItem(ls.prefix + keyName)
	if err != nil {
		return nil, err
	}

	// return base32768.SafeEncoding.DecodeString(value)
	return base64.StdEncoding.DecodeString(value)
}

// Set encodes the bytes to a string and adds them to local storage at the
// given key name. Returns an error if local storage quota has been reached.
func (ls *LocalStorage) Set(keyName string, keyValue []byte) error {
	// encoded := base32768.SafeEncoding.EncodeToString(keyValue)
	encoded := base64.StdEncoding.EncodeToString(keyValue)
	return ls.v.SetItem(ls.prefix+keyName, encoded)
}

// RemoveItem removes a key's value from local storage given its name. If there
// is no item with the given key, this function does nothing.
func (ls *LocalStorage) RemoveItem(keyName string) {
	ls.v.RemoveItem(ls.prefix + keyName)
}

// Clear clears all the keys in storage. Returns the number of keys cleared.
func (ls *LocalStorage) Clear() int {
	// Get a copy of all key names at once
	keys := ls.v.KeysPrefix(ls.prefix)

	// Loop through each key
	for _, keyName := range keys {
		ls.RemoveItem(keyName)
	}

	return len(keys)
}

// ClearPrefix clears all keys with the given prefix.  Returns the number of
// keys cleared.
func (ls *LocalStorage) ClearPrefix(prefix string) int {
	// Get a copy of all key names at once
	keys := ls.v.KeysPrefix(ls.prefix + prefix)

	// Loop through each key
	for _, keyName := range keys {
		ls.RemoveItem(prefix + keyName)
	}

	return len(keys)
}

// Key returns the name of the nth key in localStorage. Return [os.ErrNotExist]
// if the key does not exist. The order of keys is not defined.
func (ls *LocalStorage) Key(n int) (string, error) {
	keyName, err := ls.v.Key(n)
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(keyName, ls.prefix), nil
}

// Keys returns a list of all key names in local storage.
func (ls *LocalStorage) Keys() []string {
	return ls.v.KeysPrefix(ls.prefix)
}

// Length returns the number of keys in localStorage.
func (ls *LocalStorage) Length() int {
	return ls.v.Length()
}

// LocalStorageUNSAFE returns the underlying local storage wrapper. This can be
// UNSAFE and should only be used if you know what you are doing.
//
// The returned wrapper wraps all the functions and fields on the Javascript
// localStorage object to handle type conversions and errors. But it does not
// decode/sanitize the inputs/outputs or track entries using the prefix system.
// If using it, make sure all key names and values can be converted to valid
// UCS-2 strings.
func (ls *LocalStorage) LocalStorageUNSAFE() *LocalStorageJS {
	return ls.v
}

////////////////////////////////////////////////////////////////////////////////
// Javascript Wrappers                                                        //
////////////////////////////////////////////////////////////////////////////////

// LocalStorageJS stores the Javascript window.localStorage object and wraps all
// of its methods and fields to handle type conversations and errors.
//
// Doc: https://developer.mozilla.org/en-US/docs/Web/API/Window/localStorage
type LocalStorageJS struct {
	js.Value
}

// GetItem returns the value from the local storage given its key name. Returns
// [os.ErrNotExist] if the key does not exist.
//
// Doc: https://developer.mozilla.org/en-US/docs/Web/API/Storage/getItem
func (ls *LocalStorageJS) GetItem(keyName string) (keyValue string, err error) {
	defer exception.Catch(&err)
	keyValueJS := ls.Call("getItem", keyName)
	if keyValueJS.IsNull() {
		return "", os.ErrNotExist
	}
	return keyValueJS.String(), nil
}

// SetItem adds the value to local storage at the given key name. Returns an
// error if local storage quota has been reached.
//
// Doc: https://developer.mozilla.org/en-US/docs/Web/API/Storage/setItem
func (ls *LocalStorageJS) SetItem(keyName, keyValue string) (err error) {
	defer exception.Catch(&err)
	ls.Call("setItem", keyName, keyValue)
	return nil
}

// RemoveItem removes a key's value from local storage given its name. If there
// is no item with the given key, this function does nothing.
//
// Doc: https://developer.mozilla.org/en-US/docs/Web/API/Storage/removeItem
func (ls *LocalStorageJS) RemoveItem(keyName string) {
	ls.Call("removeItem", keyName)
}

// Clear clears all the keys in storage.
//
// Doc: https://developer.mozilla.org/en-US/docs/Web/API/Storage/clear
func (ls *LocalStorageJS) Clear() {
	ls.Call("clear")
}

// Key returns the name of the nth key in localStorage. Return [os.ErrNotExist]
// if the key does not exist. The order of keys is not defined.
//
// Doc: https://developer.mozilla.org/en-US/docs/Web/API/Storage/key
func (ls *LocalStorageJS) Key(n int) (keyName string, err error) {
	defer exception.Catch(&err)
	keyNameJS := ls.Call("key", n)
	if keyNameJS.IsNull() {
		return "", os.ErrNotExist
	}
	return keyNameJS.String(), nil
}

// Keys returns a list of all key names in local storage.
func (ls *LocalStorageJS) Keys() []string {
	keysJS := utils.Object.Call("keys", ls.Value)
	keys := make([]string, keysJS.Length())
	for i := range keys {
		keys[i] = keysJS.Index(i).String()
	}
	return keys
}

// KeysPrefix returns a list of all key names in local storage with the given
// prefix and trims the prefix from each key name.
func (ls *LocalStorageJS) KeysPrefix(prefix string) []string {
	keysJS := utils.Object.Call("keys", ls.Value)
	keys := make([]string, 0, keysJS.Length())
	for i := 0; i < keysJS.Length(); i++ {
		keyName := keysJS.Index(i).String()
		if strings.HasPrefix(keyName, prefix) {
			keys = append(keys, strings.TrimPrefix(keyName, prefix))
		}
	}
	return keys
}

// Length returns the number of keys in localStorage.
//
// Doc: https://developer.mozilla.org/en-US/docs/Web/API/Storage/length
func (ls *LocalStorageJS) Length() int {
	return ls.Get("length").Int()
}
