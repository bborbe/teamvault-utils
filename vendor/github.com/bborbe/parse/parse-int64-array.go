// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parse

import (
	"context"

	"github.com/bborbe/errors"
)

func ParseInt64Array(ctx context.Context, value interface{}) ([]int64, error) {
	switch v := value.(type) {
	case []int64:
		return v, nil
	case []interface{}:
		return ParseInt64ArrayFromInterfaces(ctx, v)
	case []int:
		return ParseInt64ArrayFromInterfaces(ctx, ToInterfaceList(v))
	case []int32:
		return ParseInt64ArrayFromInterfaces(ctx, ToInterfaceList(v))
	case []float32:
		return ParseInt64ArrayFromInterfaces(ctx, ToInterfaceList(v))
	case []float64:
		return ParseInt64ArrayFromInterfaces(ctx, ToInterfaceList(v))
	case []string:
		return ParseInt64ArrayFromInterfaces(ctx, ToInterfaceList(v))
	default:
		return nil, errors.Errorf(ctx, "invalid type %T", v)
	}
}

func ParseInt64ArrayDefault(ctx context.Context, value interface{}, defaultValue []int64) []int64 {
	result, err := ParseInt64Array(ctx, value)
	if err != nil {
		return defaultValue
	}
	return result
}

func ParseInt64ArrayFromInterfaces(ctx context.Context, values []interface{}) ([]int64, error) {
	result := make([]int64, len(values))
	for i, vv := range values {
		pi, err := ParseInt64(ctx, vv)
		if err != nil {
			return nil, errors.Wrapf(ctx, err, "parse int64 failed")
		}
		result[i] = pi
	}
	return result, nil
}
