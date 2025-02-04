// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

import "math"

func Round[T ~float64](value T) T {
	return T(math.Round(float64(value)))
}
