// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package validation

import (
	"context"

	"github.com/bborbe/errors"
)

func NotNil(value any) HasValidation {
	return HasValidationFunc(func(ctx context.Context) error {
		if Nil(value).Validate(ctx) == nil {
			return errors.Wrapf(ctx, Error, "should be not nil")
		}
		return nil
	})
}
