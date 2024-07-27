// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package collection

func Exclude[T comparable](list []T, excludes ...T) []T {
	e := make(map[T]bool)
	for _, exclude := range excludes {
		e[exclude] = true
	}
	result := make([]T, 0)
	for _, l := range list {
		if e[l] {
			continue
		}
		result = append(result, l)
	}
	return result
}
