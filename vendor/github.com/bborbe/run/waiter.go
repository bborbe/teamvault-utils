// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package run

import (
	"context"
	"time"

	"github.com/golang/glog"
)

//counterfeiter:generate -o mocks/waiter.go --fake-name Waiter . Waiter
type Waiter interface {
	Wait(ctx context.Context, wait time.Duration) error
}

type WaiterFunc func(ctx context.Context, wait time.Duration) error

func (w WaiterFunc) Wait(ctx context.Context, wait time.Duration) error {
	return w(ctx, wait)
}

func NewWaiter() Waiter {
	return WaiterFunc(func(ctx context.Context, wait time.Duration) error {
		glog.V(3).Infof("sleep for %v", wait)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.NewTimer(wait).C:
			return nil
		}
	})
}
