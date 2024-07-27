// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

import "math"

const float64EqualityThreshold = 1e-9

func Float64AlmostEqual(a, b float64) bool {
	return math.Abs(a-b) <= float64EqualityThreshold
}
