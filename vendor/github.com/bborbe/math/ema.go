// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

import "github.com/bborbe/collection"

// Ema calculates the Exponentially Weighted Moving Average
func Ema[T ~float64 | ~int64 | ~uint64 | ~int](values []T) *T {
	if len(values) == 0 {
		return nil
	}
	alpha := 1.0 / T(len(values))
	ema := values[0]
	for i := 1; i < len(values); i++ {
		ema = (values[i] * alpha) + (ema * (1 - alpha))
	}
	return collection.Ptr(ema)
}
