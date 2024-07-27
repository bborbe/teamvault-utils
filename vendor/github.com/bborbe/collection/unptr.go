// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package collection

func UnPtr[T any](ptr *T) T {
	var t T
	if ptr != nil {
		t = *ptr
	}
	return t
}
