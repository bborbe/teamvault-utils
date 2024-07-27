// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package errors

import stderrors "errors"

func As(err error, target any) bool {
	if stderrors.As(err, target) {
		return true
	}
	if stderrors.As(Unwrap(err), target) {
		return true
	}
	return false
}
