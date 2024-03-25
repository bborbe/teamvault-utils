// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package run

import (
	"context"
	"time"
)

// Delayed wraps the given function that delays the execution.
func Delayed(fn Func, duration time.Duration) Func {
	return func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.NewTimer(duration).C:
			return fn(ctx)
		}
	}
}
