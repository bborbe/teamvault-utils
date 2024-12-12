// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import (
	"context"
	"strings"
	stdtime "time"

	"github.com/bborbe/errors"
	"github.com/bborbe/parse"
)

func ParseTimeDefault(ctx context.Context, value interface{}, defaultValue stdtime.Time) stdtime.Time {
	result, err := ParseTime(ctx, value)
	if err != nil {
		return defaultValue
	}
	return *result
}

func ParseTime(ctx context.Context, value interface{}) (*stdtime.Time, error) {
	str, err := parse.ParseString(ctx, value)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "parse value failed")
	}
	const nowConst = "NOW"
	if strings.HasPrefix(str, nowConst) {
		now := Now()
		if len(str) > len(nowConst) {
			durationString := str[len(nowConst):]
			duration, err := stdtime.ParseDuration(durationString)
			if err != nil {
				return nil, errors.Wrapf(ctx, err, "parse duration '%s' failed", durationString)
			}
			now = now.Add(duration)
		}
		return &now, nil
	}
	var t stdtime.Time
	for _, layout := range []string{
		stdtime.RFC3339Nano,
		stdtime.RFC3339,
		"2006-01-02T15:04Z07:00",
		stdtime.DateTime,
		stdtime.DateOnly,
	} {
		t, err = stdtime.Parse(layout, str)
		if err == nil {
			return &t, nil
		}
	}
	return nil, errors.Wrapf(ctx, err, "parse time failed")
}
