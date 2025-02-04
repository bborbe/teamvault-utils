// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

func Sum[T ~float64 | ~int64 | ~uint64 | ~int](values []T) T {
	var result T = 0
	for _, v := range values {
		result += v
	}
	return result
}
