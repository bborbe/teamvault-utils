// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package collection

func Copy[T any](values []T) []T {
	result := make([]T, 0, len(values))
	for _, v := range values {
		result = append(result, v)
	}
	return result
}
