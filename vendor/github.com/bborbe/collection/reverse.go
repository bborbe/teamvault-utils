// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package collection

func Reverse[T any](values []T) []T {
	length := len(values)
	result := make([]T, length)
	for i, value := range values {
		result[length-i-1] = value
	}
	return result
}
