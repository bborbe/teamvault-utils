// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package collection

func Unique[T comparable](list []T) []T {
	result := make([]T, 0)
	m := map[T]bool{}
	for _, ee := range list {
		if m[ee] {
			continue
		}
		m[ee] = true
		result = append(result, ee)
	}
	return result
}
