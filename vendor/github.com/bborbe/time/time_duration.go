// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import (
	"context"
	"regexp"
	"strconv"
	"time"

	"github.com/bborbe/errors"
)

const (
	Nanosecond  time.Duration = 1
	Microsecond               = 1000 * Nanosecond
	Millisecond               = 1000 * Microsecond
	Second                    = 1000 * Millisecond
	Minute                    = 60 * Second
	Hour                      = 60 * Minute
	Day                       = 24 * Hour
	Week                      = 7 * Day
)

// UnitMap contains units to duration mapping
var UnitMap = map[string]time.Duration{
	"ns": Nanosecond,
	"us": Microsecond,
	"ms": Millisecond,
	"s":  Second,
	"m":  Minute,
	"h":  Hour,
	"d":  Day,
	"w":  Week,
}

var durationRegexp = regexp.MustCompile(`^((\d*\.?\d+)(w))?((\d*\.?\d+)(d))?((\d*\.?\d+)(h))?((\d*\.?\d+)(m))?((\d*\.?\d+)(s))?((\d*\.?\d+)(ms))?((\d*\.?\d+)(us))?((\d*\.?\d+)(ns))?$`)

func ParseDuration(ctx context.Context, input string) (*time.Duration, error) {
	var isNegative bool
	if len(input) > 0 && input[0] == '-' {
		isNegative = true
		input = input[1:]
	}
	var result time.Duration
	matches := durationRegexp.FindStringSubmatch(input)
	if len(matches) == 0 {
		return nil, errors.Errorf(ctx, "parse failed")
	}
	for i := 1; i < len(matches); i += 3 {
		value := matches[i+1]
		unit := matches[i+2]
		if value == "" || unit == "" {
			continue
		}
		duration, err := parseAsDuration(ctx, value, unit)
		if err != nil {
			return nil, errors.Wrapf(ctx, err, "parse failed")
		}
		result += duration
	}
	if isNegative {
		result = result * -1
	}
	return &result, nil
}

func parseAsDuration(ctx context.Context, value string, unit string) (time.Duration, error) {
	factor, ok := UnitMap[unit]
	if !ok {
		return 0, errors.Errorf(ctx, "unkown unit '%s'", unit)
	}
	i, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, errors.Wrapf(ctx, err, "parse failed")
	}
	return time.Duration(i * float64(factor)), nil
}
