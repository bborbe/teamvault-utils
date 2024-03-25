// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package validation

import (
	"context"

	"github.com/bborbe/errors"
)

func Not(hasValidation HasValidation) HasValidation {
	return HasValidationFunc(func(ctx context.Context) error {
		if err := hasValidation.Validate(ctx); err != nil {
			if errors.Is(err, Error) {
				return nil
			}
			return err
		}
		return errors.Wrapf(ctx, Error, "expected not valid")
	})
}
