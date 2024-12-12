// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parse

import (
	"context"
	"time"

	"github.com/bborbe/errors"
)

func ParseTime(ctx context.Context, value interface{}, format string) (time.Time, error) {
	str, err := ParseString(ctx, value)
	if err != nil {
		return time.Time{}, errors.Wrapf(ctx, err, "parse %v as string failed", value)
	}
	t, err := time.Parse(format, str)
	if err != nil {
		return time.Time{}, errors.Wrapf(ctx, err, "parse '%s' with format '%s' failed", value, format)
	}
	return t, nil
}

func ParseTimeDefault(ctx context.Context, value interface{}, format string, defaultValue time.Time) time.Time {
	result, err := ParseTime(ctx, value, format)
	if err != nil {
		return defaultValue
	}
	return result
}
