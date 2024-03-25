// Copyright (c) 2019 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package run

import "context"

// Func interface for all run utils.
type Func func(context.Context) error

// Run the func
func (r Func) Run(ctx context.Context) error {
	return r(ctx)
}

//counterfeiter:generate -o mocks/runnable.go --fake-name Runnable . Runnable

// Runnable interface
type Runnable interface {
	Run(ctx context.Context) error
}
