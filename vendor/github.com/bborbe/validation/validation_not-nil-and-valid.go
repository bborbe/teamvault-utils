// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package validation

import (
	"context"

	"github.com/bborbe/errors"
)

// NotNilAndValid checks if arg is not and if arg valid
func NotNilAndValid(value HasValidation) HasValidation {
	return HasValidationFunc(func(ctx context.Context) error {
		if Nil(value).Validate(ctx) == nil {
			return errors.Wrapf(ctx, Error, "should be not nil")
		}
		return value.Validate(ctx)
	})
}
