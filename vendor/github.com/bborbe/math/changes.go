// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

// Changes return changes between values (result has length-1 elements)
func Changes[T ~float64 | ~int64 | ~int](values []T) []T {
	if len(values) == 0 {
		return nil
	}
	result := make([]T, len(values)-1)
	for i := 1; i < len(values); i++ {
		result[i-1] = values[i] - values[i-1]
	}
	return result
}
