// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

import "github.com/bborbe/collection"

func Avg[T ~float64 | ~int64 | ~uint64 | ~int](values []T) *T {
	if len(values) == 0 {
		return nil
	}
	return collection.Ptr(Sum(values) / T(len(values)))
}
