// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package validation

// NilOrValid is valid if arg is nil or arg is valid
func NilOrValid(validation HasValidation) HasValidation {
	return Any{
		Nil(validation),
		validation,
	}
}
