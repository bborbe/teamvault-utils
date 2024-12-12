// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parse

import (
	"context"

	"github.com/bborbe/errors"
)

func ParseBool(ctx context.Context, value interface{}) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		switch v {
		case "true":
			return true, nil
		case "false":
			return false, nil
		}
	}
	return false, errors.Errorf(ctx, "invalid type")
}

func ParseBoolDefault(ctx context.Context, value interface{}, defaultValue bool) bool {
	result, err := ParseBool(ctx, value)
	if err != nil {
		return defaultValue
	}
	return result
}
