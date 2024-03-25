// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package validation

import "context"

//counterfeiter:generate -o mocks/validation-has-validation.go --fake-name ValidationHasValidation . HasValidation
type HasValidation interface {
	Validate(ctx context.Context) error
}

type HasValidationFunc func(ctx context.Context) error

func (v HasValidationFunc) Validate(ctx context.Context) error {
	return v(ctx)
}
