// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package validation

import (
	"context"

	"github.com/bborbe/errors"
)

func Name(fieldname string, validation HasValidation) HasValidation {
	return HasValidationFunc(func(ctx context.Context) error {
		if err := validation.Validate(ctx); err != nil {
			return errors.Wrapf(ctx, err, "validate %s failed", fieldname)
		}
		return nil
	})
}
