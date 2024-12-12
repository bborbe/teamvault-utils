// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import "time"

func Compare(x, y time.Time) int {
	if x.Before(y) {
		return -1
	}
	if x.After(y) {
		return 1
	}
	return 0
}
