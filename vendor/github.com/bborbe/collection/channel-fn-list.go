// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package collection

import (
	"context"

	"github.com/bborbe/errors"
)

func ChannelFnList[T interface{}](
	ctx context.Context,
	fn func(ctx context.Context, ch chan<- T) error,
) ([]T, error) {
	result := make([]T, 0)
	err := ChannelFnMap(
		ctx,
		fn,
		func(ctx context.Context, t T) error {
			result = append(result, t)
			return nil
		},
	)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "convert channel to list failed")
	}
	return result, nil
}
