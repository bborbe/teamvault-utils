// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

import "math"

func Sqrt[T ~float64](x T) T {
	return T(math.Sqrt(float64(x)))
}
