// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kv

import (
	"bytes"
	"strings"
)

type BucketNames []BucketName

func (t BucketNames) Contains(value BucketName) bool {
	for _, tt := range t {
		if tt.Equal(value) {
			return true
		}
	}
	return false
}

func BucketFromStrings(values ...string) BucketName {

	return NewBucketName(strings.Join(values, "_"))
}

func NewBucketName(name string) BucketName {
	return BucketName(name)
}

type BucketName []byte

func (b BucketName) String() string {
	return string(b)
}

func (b BucketName) Bytes() []byte {
	return b
}

func (b BucketName) Equal(value BucketName) bool {
	return bytes.Compare(b, value) == 0
}
