// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package validation

import (
	"context"

	"github.com/bborbe/errors"
)

func Equal[T comparable](value T, expected T) HasValidation {
	return HasValidationFunc(func(ctx context.Context) error {
		if value != expected {
			return errors.Wrapf(ctx, Error, "expect to be %v", expected)
		}
		return nil
	})
}
