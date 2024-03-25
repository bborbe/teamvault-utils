// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package errors

import (
	"context"

	"github.com/pkg/errors"
)

func New(ctx context.Context, message string) error {
	return AddContextDataToError(ctx, errors.New(message))
}

func Errorf(ctx context.Context, format string, args ...interface{}) error {
	return AddContextDataToError(ctx, errors.Errorf(format, args...))
}

func Wrap(ctx context.Context, err error, message string) error {
	if err == nil {
		return nil
	}
	return AddContextDataToError(ctx, errors.Wrap(err, message))
}

func Wrapf(ctx context.Context, err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return AddContextDataToError(ctx, errors.Wrapf(err, format, args...))
}

func AddContextDataToError(ctx context.Context, err error) error {
	return AddDataToError(err, DataFromContext(ctx))
}

func Cause(err error) error {
	return errors.Cause(err)
}
