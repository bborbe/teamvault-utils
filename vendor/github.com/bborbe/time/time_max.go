// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import "time"

func Max(a, b time.Time) time.Time {
	if a.Before(b) {
		return b
	}
	return a
}
