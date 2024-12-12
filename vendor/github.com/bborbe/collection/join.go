// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package collection

// Join allow to join two arrays into one new array
func Join[T any](a []T, b []T) []T {
	result := make([]T, 0, len(a)+len(b))
	for _, aa := range a {
		result = append(result, aa)
	}
	for _, bb := range b {
		result = append(result, bb)
	}
	return result
}
