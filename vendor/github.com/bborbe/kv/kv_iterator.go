// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kv

//counterfeiter:generate -o mocks/iterator.go --fake-name Iterator . Iterator
type Iterator interface {
	Close()
	Item() Item
	Next()
	Valid() bool
	Rewind()
	Seek(key []byte)
}
