// Copyright (c) 2021 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package run

import (
	"context"
	"fmt"
)

// CatchPanic catchs all panics for the given func
func CatchPanic(fn Func) Func {
	return func(ctx context.Context) (err error) {
		defer func() {
			if panic := recover(); panic != nil {
				err = fmt.Errorf("catch panic: %v", panic)
			}
		}()
		return fn(ctx)
	}
}
