// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	stdtime "time"

	"github.com/bborbe/errors"
	"github.com/bborbe/parse"

	"github.com/bborbe/validation"
)

type UnixTimes []UnixTime

func (t UnixTimes) Interfaces() []interface{} {
	result := make([]interface{}, len(t))
	for i, ss := range t {
		result[i] = ss
	}
	return result
}

func (t UnixTimes) Strings() []string {
	result := make([]string, len(t))
	for i, ss := range t {
		result[i] = ss.String()
	}
	return result
}

func UnixTimeFromBinary(ctx context.Context, value []byte) (*UnixTime, error) {
	var t stdtime.Time
	if err := t.UnmarshalBinary(value); err != nil {
		return nil, errors.Wrapf(ctx, err, "unmarshalBinary failed")
	}
	return UnixTime(t).Ptr(), nil
}

func ParseUnixTimeDefault(ctx context.Context, value interface{}, defaultValue UnixTime) UnixTime {
	result, err := ParseUnixTime(ctx, value)
	if err != nil {
		return defaultValue
	}
	return *result
}

func ParseUnixTime(ctx context.Context, value interface{}) (*UnixTime, error) {
	number, err := parse.ParseInt64(ctx, value)
	if err == nil {
		return UnixTimeFromSeconds(number).Ptr(), nil
	}
	str, err := parse.ParseString(ctx, value)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "parse value failed")
	}
	time, err := ParseTime(ctx, str)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "parse time failed")
	}
	return UnixTimePtr(time), nil
}

func UnixTimePtr(time *stdtime.Time) *UnixTime {
	if time == nil {
		return nil
	}
	return UnixTime(*time).Ptr()
}

func UnixTimeFromSeconds(seconds int64) UnixTime {
	return UnixTime(stdtime.Unix(seconds, 0))
}

func UnixTimeFromMilli(msec int64) UnixTime {
	return UnixTime(stdtime.UnixMilli(msec))
}

func UnixTimeFromMicro(usec int64) UnixTime {
	return UnixTime(stdtime.UnixMicro(usec))
}

type UnixTime stdtime.Time

func (u UnixTime) Year() int {
	return u.Time().Year()
}

func (u UnixTime) Month() stdtime.Month {
	return u.Time().Month()
}

func (u UnixTime) Day() int {
	return u.Time().Day()
}

func (u UnixTime) Hour() int {
	return u.Time().Hour()
}

func (u UnixTime) Minute() int {
	return u.Time().Minute()
}

func (u UnixTime) Second() int {
	return u.Time().Second()
}

func (u UnixTime) Nanosecond() int {
	return u.Time().Nanosecond()
}

func (u UnixTime) Equal(unixTime UnixTime) bool {
	return u.Time().Equal(unixTime.Time())
}

func (u *UnixTime) EqualPtr(unixTime *UnixTime) bool {
	if u == nil && unixTime == nil {
		return true
	}
	if u != nil && unixTime != nil {
		return u.Equal(*unixTime)
	}
	return false
}

func (u UnixTime) String() string {
	return u.Format(stdtime.RFC3339Nano)
}

func (u UnixTime) Validate(ctx context.Context) error {
	if u.Time().IsZero() {
		return errors.Wrapf(ctx, validation.Error, "time is zero")
	}
	return nil
}

func (u UnixTime) Ptr() *UnixTime {
	return &u
}

func (u *UnixTime) UnmarshalJSON(b []byte) error {
	str := strings.Trim(string(b), `"`)
	n, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return errors.Wrapf(context.Background(), err, "parse in location failed")
	}
	*u = UnixTimeFromSeconds(n)
	return nil
}

func (u UnixTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.Time().Unix())
}

func (u *UnixTime) Time() stdtime.Time {
	return stdtime.Time(*u)
}

func (u *UnixTime) TimePtr() *stdtime.Time {
	t := stdtime.Time(*u)
	return &t
}

func (u UnixTime) Format(layout string) string {
	return u.Time().Format(layout)
}

func (u UnixTime) MarshalBinary() ([]byte, error) {
	return u.Time().MarshalBinary()
}

func (u UnixTime) Add(duration stdtime.Duration) UnixTime {
	return UnixTime(u.Time().Add(duration))
}

func (u UnixTime) Sub(duration DateTime) Duration {
	return Duration(u.Time().Sub(duration.Time()))
}

func (u UnixTime) UnixMicro() int64 {
	return u.Time().UnixMicro()
}

func (u UnixTime) Unix() int64 {
	return u.Time().Unix()
}
