// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package validation

import (
	"context"

	"github.com/bborbe/errors"
)

// NotEmptyString return as valdation
// that check if string is not empty
func NotEmptyString[T ~string](value T) HasValidation {
	return HasValidationFunc(func(ctx context.Context) error {
		if len(value) == 0 {
			return errors.Wrapf(ctx, Error, "empty string")
		}
		return nil
	})
}
