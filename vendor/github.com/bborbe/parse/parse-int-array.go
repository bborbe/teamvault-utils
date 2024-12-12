// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parse

import (
	"context"

	"github.com/bborbe/errors"
)

func ParseIntArray(ctx context.Context, value interface{}) ([]int, error) {
	switch v := value.(type) {
	case []int:
		return v, nil
	case []interface{}:
		return ParseIntArrayFromInterfaces(ctx, v)
	case []int32:
		return ParseIntArrayFromInterfaces(ctx, ToInterfaceList(v))
	case []int64:
		return ParseIntArrayFromInterfaces(ctx, ToInterfaceList(v))
	case []float32:
		return ParseIntArrayFromInterfaces(ctx, ToInterfaceList(v))
	case []float64:
		return ParseIntArrayFromInterfaces(ctx, ToInterfaceList(v))
	case []string:
		return ParseIntArrayFromInterfaces(ctx, ToInterfaceList(v))
	default:
		return nil, errors.Errorf(ctx, "invalid type %T", v)
	}
}

func ParseIntArrayDefault(ctx context.Context, value interface{}, defaultValue []int) []int {
	result, err := ParseIntArray(ctx, value)
	if err != nil {
		return defaultValue
	}
	return result
}

func ParseIntArrayFromInterfaces(ctx context.Context, values []interface{}) ([]int, error) {
	result := make([]int, len(values))
	for i, vv := range values {
		pi, err := ParseInt(ctx, vv)
		if err != nil {
			return nil, errors.Wrapf(ctx, err, "parse int failed")
		}
		result[i] = pi
	}
	return result, nil
}

func ToInterfaceList[T any](values []T) []interface{} {
	result := make([]interface{}, len(values))
	for i, value := range values {
		result[i] = value
	}
	return result
}
