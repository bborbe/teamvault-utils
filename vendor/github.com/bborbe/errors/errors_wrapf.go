// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package errors

import (
	"context"

	"github.com/pkg/errors"
)

func Wrapf(ctx context.Context, err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return AddContextDataToError(ctx, errors.Wrapf(err, format, args...))
}
