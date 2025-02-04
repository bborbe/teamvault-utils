// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parse

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bborbe/math"
)

func ParseInt(ctx context.Context, value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case float32:
		return int(math.Round(float64(v))), nil
	case float64:
		return int(math.Round(v)), nil
	case string:
		return strconv.Atoi(v)
	case fmt.Stringer:
		return strconv.Atoi(v.String())
	default:
		return ParseInt(ctx, fmt.Sprintf("%v", value))
	}
}

func ParseIntDefault(ctx context.Context, value interface{}, defaultValue int) int {
	result, err := ParseInt(ctx, value)
	if err != nil {
		return defaultValue
	}
	return result
}
