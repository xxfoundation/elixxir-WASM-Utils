////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

//go:build js && wasm

package storage

import (
	"errors"
	"os"
	"strings"
	"syscall/js"

	"github.com/Max-Sum/base32768"

	"gitlab.com/elixxir/wasm-utils/utils"
)

// externalStorageWasmPrefix is prefixed to every keyName saved to external storage by
// externalStorage. It allows the identification and deletion of keys only created
// by this WASM binary while ignoring keys made by other scripts on the same
// page.
//
// The chosen prefix is two characters, that when converted to UTF16, take up 4
// bytes without any zeros to make them more unique.
const externalStorageWasmPrefix = "🞮🞮"

var UnimplementedErr = errors.New("not implemented")

// ExternalStorage defines an interface for setting persistent state in a KV format
// specifically for web-based implementations.
type ExternalStorage interface {
	// Get decodes and returns the value from the external storage given its key
	// name. Returns os.ErrNotExist if the key does not exist.
	Get(key string) ([]byte, error)

	// Set encodes the bytes to a string and adds them to external storage at the
	// given key name. Returns an error if external storage quota has been reached.
	Set(key string, value []byte) error

	// Delete removes a key's value from external storage given its name. If
	// there is no item with the given key, this function does nothing.
	Delete(keyName string) error

	// Clear clears all the keys in storage. Returns the number of keys cleared and any error.
	Clear() (int, error)

	// ClearPrefix clears all keys with the given prefix. Returns the number of
	// keys cleared and any error.
	ClearPrefix(prefix string) (int, error)

	// Key returns the name of the nth key in externalStorage. Returns
	// os.ErrNotExist if the key does not exist. The order of keys is not
	// defined.
	Key(n int) (string, error)

	// Keys returns a list of all key names in external storage.
	Keys() ([]string, error)

	// GetPrefix returns the full Prefix of the KV
	GetPrefix() string

	// HasPrefix returns whether this prefix exists in the KV
	HasPrefix(prefix string) bool

	// Prefix returns a new KV with the new prefix appending
	Prefix(prefix string) (ExternalStorage, error)

	// Root returns the KV with no prefixes
	Root() ExternalStorage

	// IsMemStore returns true if the underlying KV is memory based
	IsMemStore() (bool, error)

	// Length returns the number of keys in localStorage.
	Length() int

	// ExternalStorageUNSAFE returns the underlying external storage wrapper. This can
	// be UNSAFE and should only be used if you know what you are doing.
	//
	// The returned wrapper wraps all the functions and fields on the Javascript
	// havenStorage object to handle type conversions and errors. But it does
	// not decode/sanitize the inputs/outputs or track entries using the prefix
	// system. If using it, make sure all key names and values can be converted
	// to valid UCS-2 strings.
	ExternalStorageUNSAFE() *HavenStorageJS
}

// externalStorage contains the js.Value representation of havenStorage.
type externalStorage struct {
	// The Javascript value containing the havenStorage object
	v *HavenStorageJS

	// The prefix appended to each key name. This is so that all keys created by
	// this structure can be deleted without affecting other keys in external
	// storage.
	prefix string
}

// jsStorage is the global that stores Javascript as window.havenStorage.
var jsExternalStorage ExternalStorage = newExternalStorage(externalStorageWasmPrefix)

// checkUnimplementedErr checks if the error is UnimplementedErr, if yes return
// UnimplementedErr otherwise return the error
func checkUnimplementedErr(jsErr []js.Value) error {
	// todo it can be of non error type
	jsError := js.Error{Value: jsErr[0]}
	if jsError.Error() == "JavaScript error: not implemented" {
		return UnimplementedErr
	}
	return jsError
}

// newExternalStorage creates a new externalStorage object with the specified prefix.
func newExternalStorage(prefix string) *externalStorage {
	return &externalStorage{
		v:      &HavenStorageJS{js.Global().Get("havenStorage")},
		prefix: prefix,
	}
}

// GetExternalStorage returns Javascript's external storage.
func GetExternalStorage() ExternalStorage {
	return jsExternalStorage
}

// Get decodes and returns the value from the external storage given its key
// name. Returns os.ErrNotExist if the key does not exist.
func (ls *externalStorage) Get(keyName string) ([]byte, error) {
	value, err := ls.v.GetItem(ls.prefix + keyName)
	if err != nil {
		return nil, err
	}

	return base32768.SafeEncoding.DecodeString(value)
}

// Set encodes the bytes to a string and adds them to external storage at the
// given key name. Returns an error if external storage quota has been reached.
func (ls *externalStorage) Set(keyName string, keyValue []byte) error {
	encoded := base32768.SafeEncoding.EncodeToString(keyValue)
	return ls.v.SetItem(ls.prefix+keyName, encoded)
}

// RemoveItem removes a key's value from external storage given its name. If there
// is no item with the given key, this function does nothing.
func (ls *externalStorage) Delete(keyName string) error {
	return ls.v.Delete(ls.prefix + keyName)
}

// Clear clears all the keys in storage. Returns the number of keys cleared and any error.
func (ls *externalStorage) Clear() (int, error) {
	// Get a copy of all key names at once
	keys, err := ls.v.KeysPrefix(ls.prefix)
	if err != nil {
		return 0, err
	}

	// Loop through each key
	for _, keyName := range keys {
		if err := ls.Delete(keyName); err != nil {
			return 0, err
		}
	}

	return len(keys), nil
}

// ClearPrefix clears all keys with the given prefix. Returns the number of
// keys cleared and any error.
func (ls *externalStorage) ClearPrefix(prefix string) (int, error) {
	// Get a copy of all key names at once
	keys, err := ls.v.KeysPrefix(ls.prefix + prefix)
	if err != nil {
		return 0, err
	}

	// Loop through each key
	for _, keyName := range keys {
		if err := ls.Delete(prefix + keyName); err != nil {
			return 0, err
		}
	}

	return len(keys), nil
}

// Key returns the name of the nth key in externalStorage. Return [os.ErrNotExist]
// if the key does not exist. The order of keys is not defined.
func (ls *externalStorage) Key(n int) (string, error) {
	keyName, err := ls.v.Key(n)
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(keyName, ls.prefix), nil
}

// Keys returns a list of all key names in external storage.
func (ls *externalStorage) Keys() ([]string, error) {
	keys, err := ls.v.KeysPrefix(ls.prefix)
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (ls *externalStorage) GetPrefix() string {
	return ls.prefix
}

func (ls *externalStorage) HasPrefix(prefix string) bool {
	return strings.HasPrefix(ls.prefix, prefix)
}

func (ls *externalStorage) Prefix(prefix string) (ExternalStorage, error) {
	return newExternalStorage(ls.prefix + prefix), nil
}

func (ls *externalStorage) Root() ExternalStorage {
	return newExternalStorage("")
}

func (ls *externalStorage) IsMemStore() (bool, error) {
	return ls.v.IsMemStore()
}

func (ls *externalStorage) Length() int {
	return ls.v.Length()
}

// ExternalStorageUNSAFE returns the underlying external storage wrapper. This can be
// UNSAFE and should only be used if you know what you are doing.
//
// The returned wrapper wraps all the functions and fields on the Javascript
// havenStorage object to handle type conversions and errors. But it does not
// decode/sanitize the inputs/outputs or track entries using the prefix system.
// If using it, make sure all key names and values can be converted to valid
// UCS-2 strings.
func (ls *externalStorage) ExternalStorageUNSAFE() *HavenStorageJS {
	return ls.v
}

////////////////////////////////////////////////////////////////////////////////
// Javascript Wrappers                                                        //
////////////////////////////////////////////////////////////////////////////////

// StorageOperation defines the supported operations for HavenStorageJS
type StorageOperation string

const (
	// GetItemOp represents the "getItem" operation
	GetItemOp StorageOperation = "getItem"
	// SetItemOp represents the "setItem" operation
	SetItemOp StorageOperation = "setItem"
	// DeleteOp represents the "delete" operation
	DeleteOp StorageOperation = "delete"
	// ClearOp represents the "clear" operation
	ClearOp StorageOperation = "clear"
	// KeysOp represents the "getKeys" operation
	KeysOp StorageOperation = "getKeys"

	// IsMemStoreOp represents the "isMemStore" operation
	IsMemStoreOp StorageOperation = "isMemStore"
)

// HavenStorageJS stores the Javascript window.havenStorage object and wraps all
// of its methods and fields to handle type conversations and errors.
type HavenStorageJS struct {
	js.Value
}

// callStorage is a helper function that calls the specified operation on the storage object
// with the provided arguments and returns the result.
func (ls *HavenStorageJS) callStorage(op StorageOperation, args ...interface{}) js.Value {
	return ls.Call(string(op), args...)
}

// GetItem returns the value from the external storage given its key name. Returns
// [os.ErrNotExist] if the key does not exist.
//
// Doc: https://developer.mozilla.org/en-US/docs/Web/API/Storage/getItem
func (ls *HavenStorageJS) GetItem(keyName string) (keyValue string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = utils.ErrFromPanic(r)
		}
	}()
	promise := ls.callStorage(GetItemOp, keyName)
	result, jsErr := utils.Await(promise)
	if jsErr != nil {
		return "", checkUnimplementedErr(jsErr)
	}
	if result[0].IsNull() {
		return "", os.ErrNotExist
	}
	return result[0].String(), nil
}

// SetItem adds the value to external storage at the given key name. Returns an
// error if external storage quota has been reached.
//
// Doc: https://developer.mozilla.org/en-US/docs/Web/API/Storage/setItem
func (ls *HavenStorageJS) SetItem(keyName, keyValue string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = utils.ErrFromPanic(r)
		}
	}()
	promise := ls.callStorage(SetItemOp, keyName, keyValue)
	_, jsErr := utils.Await(promise)
	if jsErr != nil {
		return checkUnimplementedErr(jsErr)
	}
	return nil
}

// RemoveItem removes a key's value from external storage given its name. If there
// is no item with the given key, this function does nothing.
//
// Doc: https://developer.mozilla.org/en-US/docs/Web/API/Storage/removeItem
func (ls *HavenStorageJS) Delete(keyName string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = utils.ErrFromPanic(r)
		}
	}()
	promise := ls.callStorage(DeleteOp, keyName)
	_, jsErr := utils.Await(promise)
	if jsErr != nil {
		return checkUnimplementedErr(jsErr)
	}
	return nil
}

// Clear clears all the keys in storage.
//
// Doc: https://developer.mozilla.org/en-US/docs/Web/API/Storage/clear
func (ls *HavenStorageJS) Clear() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = utils.ErrFromPanic(r)
		}
	}()
	promise := ls.callStorage(ClearOp)
	_, jsErr := utils.Await(promise)
	if jsErr != nil {
		return checkUnimplementedErr(jsErr)
	}
	return nil
}

// Key returns the name of the nth key in externalStorage. Return [os.ErrNotExist]
// if the key does not exist. The order of keys is not defined.
//
// Doc: https://developer.mozilla.org/en-US/docs/Web/API/Storage/key
func (ls *HavenStorageJS) Key(n int) (keyName string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = utils.ErrFromPanic(r)
		}
	}()
	promise := ls.Call("key", n)
	result, jsErr := utils.Await(promise)
	if jsErr != nil {
		return "", checkUnimplementedErr(jsErr)
	}
	if result[0].IsNull() {
		return "", os.ErrNotExist
	}
	return result[0].String(), nil
}

// Keys returns a list of all key names in external storage.
func (ls *HavenStorageJS) Keys() (keys []string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = utils.ErrFromPanic(r)
		}
	}()
	promise := ls.callStorage(KeysOp)
	result, jsErr := utils.Await(promise)
	if jsErr != nil {
		return []string{}, checkUnimplementedErr(jsErr)
	}

	keysJS := result[0]
	keys = make([]string, keysJS.Length())
	for i := range keys {
		keys[i] = keysJS.Index(i).String()
	}
	return keys, nil
}

// KeysPrefix returns a list of all key names in external storage with the given
// prefix and trims the prefix from each key name.
func (ls *HavenStorageJS) KeysPrefix(prefix string) (keys []string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = utils.ErrFromPanic(r)
		}
	}()
	promise := ls.callStorage(KeysOp)
	result, jsErr := utils.Await(promise)
	if jsErr != nil {
		return []string{}, checkUnimplementedErr(jsErr)
	}

	keysJS := result[0]
	keys = make([]string, 0, keysJS.Length())
	for i := 0; i < keysJS.Length(); i++ {
		keyName := keysJS.Index(i).String()
		if strings.HasPrefix(keyName, prefix) {
			keys = append(keys, strings.TrimPrefix(keyName, prefix))
		}
	}
	return keys, nil
}

func (ls *HavenStorageJS) IsMemStore() (bool, error) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = utils.ErrFromPanic(r)
		}
	}()
	result := ls.callStorage(IsMemStoreOp)
	if err != nil {
		//TODO
		jsErr, ok := err.(js.Error)
		if ok {
			return false, checkUnimplementedErr([]js.Value{js.ValueOf(jsErr)})
		}
		return false, err
	}
	return result.Bool(), nil
}
