// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kv

//counterfeiter:generate -o mocks/item.go --fake-name Item . Item
type Item interface {
	Exists() bool
	Key() []byte
	Value(fn func(val []byte) error) error
}

func NewByteItem(
	key []byte,
	value []byte,
) Item {
	return &byteItem{
		key:   key,
		value: value,
	}
}

type byteItem struct {
	value []byte
	key   []byte
}

func (b byteItem) Exists() bool {
	return len(b.value) > 0
}

func (b byteItem) Key() []byte {
	return b.key
}

func (b byteItem) Value(fn func(val []byte) error) error {
	return fn(b.value)
}
