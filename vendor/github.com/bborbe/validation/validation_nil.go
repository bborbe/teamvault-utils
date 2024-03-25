// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package validation

import (
	"context"
	"reflect"

	"github.com/bborbe/errors"
)

func Nil(value any) HasValidation {
	return HasValidationFunc(func(ctx context.Context) error {
		// reflect.ValueOf(value).IsZero()
		if value == nil || (reflect.ValueOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil()) {
			return nil
		}
		return errors.Wrapf(ctx, Error, "should be nil")
	})
}
