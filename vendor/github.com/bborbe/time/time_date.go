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

type Dates []Date

func (d Dates) Interfaces() []interface{} {
	result := make([]interface{}, len(d))
	for i, ss := range d {
		result[i] = ss
	}
	return result
}

func (d Dates) Strings() []string {
	result := make([]string, len(d))
	for i, ss := range d {
		result[i] = ss.String()
	}
	return result
}

func DateFromBinary(ctx context.Context, value []byte) (*Date, error) {
	var t stdtime.Time
	if err := t.UnmarshalBinary(value); err != nil {
		return nil, errors.Wrapf(ctx, err, "unmarshalBinary failed")
	}
	return Date(t).Ptr(), nil
}

func ParseDate(ctx context.Context, value interface{}) (*Date, error) {
	str, err := parse.ParseString(ctx, value)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "parse value failed")
	}
	time, err := ParseTime(ctx, str)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "parse time failed")
	}
	return DatePtr(time), nil
}

func DatePtr(value *stdtime.Time) *Date {
	if value == nil {
		return nil
	}
	return ToDate(*value).Ptr()
}

func ToDate(value stdtime.Time) Date {
	year, month, day := value.Date()
	return Date(stdtime.Date(year, month, day, 0, 0, 0, 0, stdtime.UTC))
}

type Date stdtime.Time

func (d Date) Year() int {
	return d.Time().Year()
}

func (d Date) Month() stdtime.Month {
	return d.Time().Month()
}

func (d Date) Day() int {
	return d.Time().Day()
}

func (d Date) String() string {
	return d.Format(stdtime.DateOnly)
}

func (d Date) Validate(ctx context.Context) error {
	if d.Time().IsZero() {
		return errors.Wrapf(ctx, validation.Error, "time is zero")
	}
	return nil
}

func (d Date) Ptr() *Date {
	return &d
}

func (d *Date) UnmarshalJSON(b []byte) error {
	str := strings.Trim(string(b), `"`)
	if len(str) == 0 || str == "null" {
		*d = Date(stdtime.Time{})
		return nil
	}
	t, err := stdtime.ParseInLocation(stdtime.DateOnly, str, stdtime.UTC)
	if err != nil {
		return errors.Wrapf(context.Background(), err, "parse in location failed")
	}
	*d = Date(t)
	return nil
}

func (d Date) MarshalJSON() ([]byte, error) {
	time := d.Time()
	if time.IsZero() {
		return json.Marshal(nil)
	}
	return json.Marshal(time.Format(stdtime.DateOnly))
}

func (d *Date) Time() stdtime.Time {
	return stdtime.Time(*d)
}

func (d *Date) TimePtr() *stdtime.Time {
	t := stdtime.Time(*d)
	return &t
}

func (d Date) Format(layout string) string {
	return d.Time().Format(layout)
}

func (d Date) MarshalBinary() ([]byte, error) {
	return d.Time().MarshalBinary()
}

func (d Date) Compare(stdTime Date) int {
	return Compare(d.Time(), stdTime.Time())
}

func (d *Date) ComparePtr(stdTime *Date) int {
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

func (d Date) Add(duration stdtime.Duration) Date {
	return Date(d.Time().Add(duration))
}

func (d Date) Sub(duration DateTime) Duration {
	return Duration(d.Time().Sub(duration.Time()))
}

func (d Date) UnixMicro() int64 {
	return d.Time().UnixMicro()
}

func (d Date) Unix() int64 {
	return d.Time().Unix()
}
