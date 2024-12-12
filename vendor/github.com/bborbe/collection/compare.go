// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package collection

import "strings"

func Compare[T ~string](a, b T) int {
	return strings.Compare(string(a), string(b))
}
