// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package collection

import (
	"context"

	"github.com/bborbe/errors"
)

func ChannelFnCount[T interface{}](
	ctx context.Context,
	fn func(ctx context.Context, ch chan<- T) error,
) (int, error) {
	result := 0
	err := ChannelFnMap(
		ctx,
		fn,
		func(ctx context.Context, t T) error {
			result++
			return nil
		},
	)
	if err != nil {
		return -1, errors.Wrapf(ctx, err, "count channel failed")
	}
	return result, nil
}
