// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package validation

import (
	"context"

	"github.com/bborbe/errors"
)

func LengthEqual[T any](list []T, expectedLength int) HasValidation {
	return HasValidationFunc(func(ctx context.Context) error {
		if len(list) == expectedLength {
			return nil
		}
		return errors.Wrapf(ctx, Error, "length is not %d", expectedLength)
	})
}

func LengthGt[T any](list []T, expectedLength int) HasValidation {
	return HasValidationFunc(func(ctx context.Context) error {
		if len(list) > expectedLength {
			return nil
		}
		return errors.Wrapf(ctx, Error, "length is not > %d", expectedLength)
	})
}

func LengthGe[T any](list []T, expectedLength int) HasValidation {
	return HasValidationFunc(func(ctx context.Context) error {
		if len(list) >= expectedLength {
			return nil
		}
		return errors.Wrapf(ctx, Error, "length is not > %d", expectedLength)
	})
}

func LengthLt[T any](list []T, expectedLength int) HasValidation {
	return HasValidationFunc(func(ctx context.Context) error {
		if len(list) < expectedLength {
			return nil
		}
		return errors.Wrapf(ctx, Error, "length is not > %d", expectedLength)
	})
}

func LengthLe[T any](list []T, expectedLength int) HasValidation {
	return HasValidationFunc(func(ctx context.Context) error {
		if len(list) <= expectedLength {
			return nil
		}
		return errors.Wrapf(ctx, Error, "length is not > %d", expectedLength)
	})
}
