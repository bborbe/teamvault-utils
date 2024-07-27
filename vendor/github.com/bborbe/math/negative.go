// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

func Negative[T ~float64 | ~int64 | ~uint64 | ~int](values []T) []T {
	result := make([]T, 0)
	for _, pp := range values {
		if pp < 0 {
			result = append(result, pp)
		}
	}
	return result
}
