// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package run

import (
	"context"
)

// Triggered calls the given fn if something arrives on chan
func Triggered(fn Func, trigger <-chan struct{}) Func {
	return func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-trigger:
			return fn(ctx)
		}
	}
}
