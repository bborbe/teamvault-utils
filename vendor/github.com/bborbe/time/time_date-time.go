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

	"github.com/bborbe/validation"
)

type DateTimes []DateTime

func (t DateTimes) Interfaces() []interface{} {
	result := make([]interface{}, len(t))
	for i, ss := range t {
		result[i] = ss
	}
	return result
}

func (t DateTimes) Strings() []string {
	result := make([]string, len(t))
	for i, ss := range t {
		result[i] = ss.String()
	}
	return result
}

func DateTimeFromBinary(ctx context.Context, value []byte) (*DateTime, error) {
	var t stdtime.Time
	if err := t.UnmarshalBinary(value); err != nil {
		return nil, errors.Wrapf(ctx, err, "unmarshalBinary failed")
	}
	return DateTime(t).Ptr(), nil
}

func ParseDateTimeDefault(ctx context.Context, value interface{}, defaultValue DateTime) DateTime {
	result, err := ParseDateTime(ctx, value)
	if err != nil {
		return defaultValue
	}
	return *result
}

func ParseDateTime(ctx context.Context, value interface{}) (*DateTime, error) {
	str, err := parse.ParseString(ctx, value)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "parse value failed")
	}
	time, err := ParseTime(ctx, str)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "parse time failed")
	}
	return DateTimePtr(time), nil
}

func DateTimePtr(time *stdtime.Time) *DateTime {
	if time == nil {
		return nil
	}
	return DateTime(*time).Ptr()
}

func DateTimeFromUnixMicro(ms int64) DateTime {
	return DateTime(stdtime.UnixMicro(ms))
}

type DateTime stdtime.Time

func (d DateTime) Year() int {
	return d.Time().Year()
}

func (d DateTime) Month() stdtime.Month {
	return d.Time().Month()
}

func (d DateTime) Day() int {
	return d.Time().Day()
}

func (d DateTime) Hour() int {
	return d.Time().Hour()
}

func (d DateTime) Minute() int {
	return d.Time().Minute()
}

func (d DateTime) Second() int {
	return d.Time().Second()
}

func (d DateTime) Nanosecond() int {
	return d.Time().Nanosecond()
}

func (d DateTime) Equal(stdTime DateTime) bool {
	return d.Time().Equal(stdTime.Time())
}

func (d *DateTime) EqualPtr(stdTime *DateTime) bool {
	if d == nil && stdTime == nil {
		return true
	}
	if d != nil && stdTime != nil {
		return d.Equal(*stdTime)
	}
	return false
}

func (d DateTime) String() string {
	return d.Format(stdtime.RFC3339Nano)
}

func (d DateTime) Validate(ctx context.Context) error {
	if d.Time().IsZero() {
		return errors.Wrapf(ctx, validation.Error, "time is zero")
	}
	return nil
}

func (d DateTime) Ptr() *DateTime {
	return &d
}

func (d *DateTime) UnmarshalJSON(b []byte) error {
	str := strings.Trim(string(b), `"`)
	switch str {
	case "", "null":
		*d = DateTime(stdtime.Time{})
		return nil
	case "NOW":
		*d = DateTime(Now())
		return nil
	default:
		t, err := stdtime.ParseInLocation(stdtime.RFC3339Nano, str, stdtime.UTC)
		if err != nil {
			return errors.Wrapf(context.Background(), err, "parse in location failed")
		}
		*d = DateTime(t)
		return nil
	}
}

func (d DateTime) MarshalJSON() ([]byte, error) {
	time := d.Time()
	if time.IsZero() {
		return json.Marshal(nil)
	}
	return json.Marshal(time.Format(stdtime.RFC3339Nano))
}

func (d *DateTime) Time() stdtime.Time {
	return stdtime.Time(*d)
}

func (d *DateTime) TimePtr() *stdtime.Time {
	t := stdtime.Time(*d)
	return &t
}

func (d DateTime) Format(layout string) string {
	return d.Time().Format(layout)
}

func (d DateTime) MarshalBinary() ([]byte, error) {
	return d.Time().MarshalBinary()
}

func (d DateTime) Clone() DateTime {
	return d
}

func (d *DateTime) ClonePtr() *DateTime {
	if d == nil {
		return nil
	}
	return d.Clone().Ptr()
}

func (d DateTime) UnixMicro() int64 {
	return d.Time().UnixMicro()
}

func (d DateTime) Unix() int64 {
	return d.Time().Unix()
}

func (d DateTime) Before(stdTime DateTime) bool {
	return d.Time().Before(stdTime.Time())
}

func (d DateTime) After(stdTime DateTime) bool {
	return d.Time().After(stdTime.Time())
}

func (d DateTime) Add(duration stdtime.Duration) DateTime {
	return DateTime(d.Time().Add(duration))
}

func (d DateTime) Sub(duration DateTime) Duration {
	return Duration(d.Time().Sub(duration.Time()))
}

func (d DateTime) Compare(stdTime DateTime) int {
	return Compare(d.Time(), stdTime.Time())
}

func (d *DateTime) ComparePtr(stdTime *DateTime) int {
	if d == nil && stdTime == nil {
		return 0
	}
	if d == nil {
		return -1
	}
	if stdTime == nil {
		return 1
	}
	return d.Compare(*stdTime)
}
