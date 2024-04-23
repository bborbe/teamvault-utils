// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import (
	"context"
	"encoding/json"
	"strings"
	stdtime "time"

	"github.com/bborbe/errors"
)

const TimeOfDayLayout = "15:04:05.999999999Z07:00"

func ParseTimeOfDay(ctx context.Context, value string) (*TimeOfDay, error) {
	const nowConst = "NOW"
	if strings.HasPrefix(value, nowConst) {
		now := Now()
		return TimeOfDayFromTime(now).Ptr(), nil
	}
	if parts := strings.Split(value, " "); len(parts) == 2 {
		location, err := stdtime.LoadLocation(parts[1])
		if err != nil {
			return nil, errors.Wrapf(ctx, err, "load location '%s' failed", parts[1])
		}
		timeOfDay, err := ParseTimeOfDay(ctx, parts[0])
		if err != nil {
			return nil, errors.Wrapf(ctx, err, "parse time of day failed")
		}
		timeOfDay.Location = location
		return timeOfDay, nil
	}

	var err error
	var t stdtime.Time
	for _, layout := range []string{
		"15:04:05.999999999Z07:00",
		"15:04:05.999999999",
		"15:04Z07:00",
		"15:04",
		stdtime.RFC3339Nano,
		stdtime.RFC3339,
		stdtime.DateTime,
	} {
		t, err = stdtime.Parse(layout, value)
		if err == nil {
			return TimeOfDayFromTime(t).Ptr(), nil
		}
	}
	return nil, errors.Wrapf(ctx, err, "parse timeOfDay failed")
}

func TimeOfDayFromTime(date stdtime.Time) TimeOfDay {
	return TimeOfDay{
		Hour:       date.Hour(),
		Minute:     date.Minute(),
		Second:     date.Second(),
		Nanosecond: date.Nanosecond(),
		Location:   date.Location(),
	}
}

type TimeOfDay struct {
	Hour       int
	Minute     int
	Second     int
	Nanosecond int
	Location   *stdtime.Location
}

func (t TimeOfDay) String() string {
	return t.Format(TimeOfDayLayout)
}

func (t TimeOfDay) Format(layout string) string {
	return t.date(1970, stdtime.January, 1).Format(layout)
}

func (t TimeOfDay) Date(year int, month stdtime.Month, day int) (*stdtime.Time, error) {
	date := t.date(year, month, day)
	return &date, nil
}

func (t TimeOfDay) date(year int, month stdtime.Month, day int) stdtime.Time {
	return stdtime.Date(year, month, day, t.Hour, t.Minute, t.Second, t.Nanosecond, t.Location)
}

func (t *TimeOfDay) UnmarshalJSON(b []byte) error {
	str := strings.Trim(string(b), `"`)
	parse, err := ParseTimeOfDay(context.Background(), str)
	if err != nil {
		return errors.Wrapf(context.Background(), err, "parse day of time failed")
	}
	*t = *parse
	return nil
}

func (t TimeOfDay) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

func (t TimeOfDay) Ptr() *TimeOfDay {
	return &t
}
