// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

func Abs[T ~float64 | ~int64 | ~int](value T) T {
	if value < 0 {
		return value * T(-1)
	}
	return value
}
