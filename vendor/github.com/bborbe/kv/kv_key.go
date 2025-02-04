// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kv

type Key []byte

func (f Key) String() string {
	return string(f)
}

func (f Key) Bytes() []byte {
	return f
}
