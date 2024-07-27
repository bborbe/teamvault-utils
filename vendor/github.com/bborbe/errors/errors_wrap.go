// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package errors

import (
	"context"

	"github.com/pkg/errors"
)

func Wrap(ctx context.Context, err error, message string) error {
	if err == nil {
		return nil
	}
	return AddContextDataToError(ctx, errors.Wrap(err, message))
}
