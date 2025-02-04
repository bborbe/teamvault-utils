// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parse

import (
	"context"
	stderrors "errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/bborbe/errors"
)

var InvalidTypeError = stderrors.New("invalid type")

func ParseString(ctx context.Context, value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case bool:
		return strconv.FormatBool(v), nil
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32), nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case int:
		return strconv.FormatInt(int64(v), 10), nil
	case int32:
		return strconv.FormatInt(int64(v), 10), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case uint:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint64:
		return strconv.FormatUint(v, 10), nil
	case fmt.Stringer:
		return v.String(), nil
	default:
		if isSubtypeOfString(value) {
			return ParseString(ctx, fmt.Sprintf("%s", value))
		}
		return "", errors.Wrapf(ctx, InvalidTypeError, "parse failed")
	}
}

func isSubtypeOfString(value interface{}) bool {
	t := reflect.TypeOf(value)
	return t != nil && t.Kind() == reflect.String
}

func ParseStringDefault(ctx context.Context, value interface{}, defaultValue string) string {
	result, err := ParseString(ctx, value)
	if err != nil {
		return defaultValue
	}
	return result
}
