// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package collection

func Contains[T comparable](list []T, value T) bool {
	for _, e := range list {
		if e == value {
			return true
		}
	}
	return false
}
