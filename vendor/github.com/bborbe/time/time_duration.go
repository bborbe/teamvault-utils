// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import (
	"context"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	stdtime "time"

	"github.com/bborbe/errors"
	"github.com/bborbe/parse"
)

const (
	Nanosecond  Duration = 1
	Microsecond          = 1000 * Nanosecond
	Millisecond          = 1000 * Microsecond
	Second               = 1000 * Millisecond
	Minute               = 60 * Second
	Hour                 = 60 * Minute
	Day                  = 24 * Hour
	Week                 = 7 * Day
)

// UnitMap contains units to duration mapping
var UnitMap = map[string]Duration{
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

type Durations []Duration

func (t Durations) Interfaces() []interface{} {
	result := make([]interface{}, len(t))
	for i, ss := range t {
		result[i] = ss
	}
	return result
}

func (t Durations) Strings() []string {
	result := make([]string, len(t))
	for i, ss := range t {
		result[i] = ss.String()
	}
	return result
}

func ParseDurationDefault(ctx context.Context, value interface{}, defaultValue Duration) Duration {
	result, err := ParseDuration(ctx, value)
	if err != nil {
		return defaultValue
	}
	return *result
}

func ParseDuration(ctx context.Context, value interface{}) (*Duration, error) {
	str, err := parse.ParseString(ctx, value)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "parse value failed")
	}
	if number, err := strconv.ParseInt(str, 10, 64); err == nil {
		return Duration(number).Ptr(), err
	}

	var isNegative bool
	if len(str) > 0 && str[0] == '-' {
		isNegative = true
		str = str[1:]
	}
	var result Duration
	matches := durationRegexp.FindStringSubmatch(str)
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

func parseAsDuration(ctx context.Context, value string, unit string) (Duration, error) {
	factor, ok := UnitMap[unit]
	if !ok {
		return 0, errors.Errorf(ctx, "unkown unit '%s'", unit)
	}
	i, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, errors.Wrapf(ctx, err, "parse failed")
	}
	return Duration(i * float64(factor)), nil
}

type Duration stdtime.Duration

func (d Duration) Duration() stdtime.Duration {
	return stdtime.Duration(d)
}

func (d Duration) Abs() Duration {
	return Duration(d.Duration().Abs())
}

func (d Duration) Ptr() *Duration {
	return &d
}

func (d Duration) String() string {
	var builder strings.Builder
	remaining := d

	if weeks := remaining / Week; weeks > 0 {
		remaining = remaining - weeks*Week
		builder.WriteString(strconv.Itoa(int(weeks)))
		builder.WriteString("w")
	}

	if days := remaining / Day; days > 0 {
		remaining = remaining - days*Day
		builder.WriteString(strconv.Itoa(int(days)))
		builder.WriteString("d")
	}

	builder.WriteString(remaining.Duration().String())
	return builder.String()
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	str := strings.Trim(string(b), `"`)
	if len(str) == 0 || str == "null" {
		*d = Duration(0)
		return nil
	}
	ctx := context.Background()
	duration, err := ParseDuration(ctx, str)
	if err != nil {
		return errors.Wrapf(ctx, err, "parse duration failed")
	}

	*d = *duration
	return nil
}

func (d Duration) MarshalJSON() ([]byte, error) {
	// use stdtime.Duration.String to produce output in standard golang format
	return json.Marshal(d.Duration().String())
}
