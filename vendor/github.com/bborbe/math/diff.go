// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

import "github.com/bborbe/collection"

func Diff[T ~float64 | ~int64 | ~uint64 | ~int](a T, b T) *float64 {
	if b == 0 {
		return nil
	}
	return collection.Ptr(float64(a) / float64(b))
}
