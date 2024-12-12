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
	"github.com/bborbe/parse"
)

const TimeOfDayLayout = "15:04:05.999999999Z07:00"

type TimeOfDays []TimeOfDay

func (t TimeOfDays) Interfaces() []interface{} {
	result := make([]interface{}, len(t))
	for i, ss := range t {
		result[i] = ss
	}
	return result
}

func (t TimeOfDays) Strings() []string {
	result := make([]string, len(t))
	for i, ss := range t {
		result[i] = ss.String()
	}
	return result
}

func ParseTimeOfDayDefault(ctx context.Context, value interface{}, defaultValue TimeOfDay) TimeOfDay {
	result, err := ParseTimeOfDay(ctx, value)
	if err != nil {
		return defaultValue
	}
	return *result
}

func ParseTimeOfDay(ctx context.Context, value interface{}) (*TimeOfDay, error) {
	str, err := parse.ParseString(ctx, value)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "parse value failed")
	}
	const nowConst = "NOW"
	if strings.HasPrefix(str, nowConst) {
		now := Now()
		return TimeOfDayFromTime(now).Ptr(), nil
	}
	if parts := strings.Split(str, " "); len(parts) == 2 {
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

	var t stdtime.Time
	for _, layout := range []string{
		"15:04:05.999999999Z07:00",
		"15:04:05.999999999",
		"15:04:05Z07:00",
		"15:04:05",
		"15:04Z07:00",
		"15:04",
		stdtime.RFC3339Nano,
		stdtime.RFC3339,
		stdtime.DateTime,
	} {
		t, err = stdtime.Parse(layout, str)
		if err == nil {
			return TimeOfDayFromTime(t.In(stdtime.UTC)).Ptr(), nil
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

func (t TimeOfDay) DateTime(year int, month stdtime.Month, day int) DateTime {
	return DateTime(t.date(year, month, day))
}

func (t TimeOfDay) Time(year int, month stdtime.Month, day int) stdtime.Time {
	return t.date(year, month, day)
}

func (t TimeOfDay) Date(year int, month stdtime.Month, day int) (*stdtime.Time, error) {
	time := t.date(year, month, day)
	return &time, nil
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

func (t TimeOfDay) Before(stdTime TimeOfDay) bool {
	return t.Time(2024, 07, 01).Before(stdTime.Time(2024, 07, 01))
}

func (t TimeOfDay) After(stdTime TimeOfDay) bool {
	return t.Time(2024, 07, 01).After(stdTime.Time(2024, 07, 01))
}

func (t TimeOfDay) Equal(stdTime TimeOfDay) bool {
	return t.Time(2024, 07, 01).Equal(stdTime.Time(2024, 07, 01))
}
