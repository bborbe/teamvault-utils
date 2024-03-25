// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import (
	"context"
	"strings"
	stdtime "time"

	"github.com/bborbe/errors"
)

func ParseTime(ctx context.Context, value string) (*stdtime.Time, error) {
	const nowConst = "NOW"
	if strings.HasPrefix(value, nowConst) {
		now := Now()
		if len(value) > len(nowConst) {
			durationString := value[len(nowConst):]
			duration, err := stdtime.ParseDuration(durationString)
			if err != nil {
				return nil, errors.Wrapf(ctx, err, "parse duration '%s' failed", durationString)
			}
			now = now.Add(duration)
		}
		return &now, nil
	}
	var err error
	var t stdtime.Time
	for _, layout := range []string{
		stdtime.RFC3339Nano,
		stdtime.RFC3339,
		stdtime.DateTime,
		stdtime.DateOnly,
	} {
		t, err = stdtime.Parse(layout, value)
		if err == nil {
			return &t, nil
		}
	}
	return nil, errors.Wrapf(ctx, err, "parse time failed")
}
