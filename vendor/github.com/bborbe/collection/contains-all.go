// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package collection

// ContainsAll check if all elements of A is in B and all elements of B is in A
func ContainsAll[T comparable](a []T, b []T) bool {
	mapA := make(map[T]bool)
	for _, aa := range a {
		mapA[aa] = true
	}
	mapB := make(map[T]bool)
	for _, bb := range b {
		mapB[bb] = true
	}
	for _, aa := range a {
		if mapB[aa] == false {
			return false
		}
	}
	for _, bb := range b {
		if mapA[bb] == false {
			return false
		}
	}
	return true
}
