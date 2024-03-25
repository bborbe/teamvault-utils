// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package validation

import (
	"context"

	"github.com/bborbe/errors"
)

type All []HasValidation

func (l All) Validate(ctx context.Context) error {
	for _, ll := range l {
		if err := ll.Validate(ctx); err != nil {
			return errors.Wrapf(ctx, err, "validate failed")
		}
	}
	return nil
}
