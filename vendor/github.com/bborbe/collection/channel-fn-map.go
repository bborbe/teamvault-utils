// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package collection

import (
	"context"
	"runtime"

	"github.com/bborbe/errors"
	"github.com/bborbe/run"
)

func ChannelFnMap[T interface{}](
	ctx context.Context,
	getFn func(ctx context.Context, ch chan<- T) error,
	mapFn func(ctx context.Context, t T) error,
) error {
	var err error

	ch := make(chan T, runtime.NumCPU())
	err = run.CancelOnFirstErrorWait(
		ctx,
		func(ctx context.Context) error {
			defer close(ch)
			return getFn(ctx, ch)
		},
		func(ctx context.Context) error {
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case t, ok := <-ch:
					if !ok {
						return nil
					}
					if err := mapFn(ctx, t); err != nil {
						return errors.Wrapf(ctx, err, "map failed")
					}
				}
			}
		},
	)
	if err != nil {
		return errors.Wrapf(ctx, err, "map channel failed")
	}
	return nil
}
