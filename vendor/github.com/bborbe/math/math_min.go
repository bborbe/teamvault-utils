// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

import "github.com/bborbe/collection"

func Min[T ~float64 | ~int64 | ~uint64 | ~int](values []T) *T {
	var result *T
	for _, v := range values {
		if result == nil || *result > v {
			result = collection.Ptr(v)
		}
	}
	return result
}
