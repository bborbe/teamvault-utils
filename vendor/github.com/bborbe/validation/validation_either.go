// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package validation

import (
	"context"

	"github.com/bborbe/errors"
)

type Either []HasValidation

func (l Either) Validate(ctx context.Context) error {
	if len(l) == 0 {
		return errors.Wrapf(ctx, Error, "either can't be empty")
	}
	var errs []error
	for _, ll := range l {
		if err := ll.Validate(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	switch len(l) - len(errs) {
	case 1:
		return nil
	default:
		return errors.Wrapf(ctx, Error, "only one should be valid")
	}
}
