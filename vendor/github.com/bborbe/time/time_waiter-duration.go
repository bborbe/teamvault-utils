// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import (
	"context"
	"time"

	"github.com/golang/glog"
)

//counterfeiter:generate -o mocks/waiter-duration.go --fake-name WaiterDuration . WaiterDuration
type WaiterDuration interface {
	Wait(ctx context.Context, duration Duration) error
}

type WaiterDurationFunc func(ctx context.Context, duration Duration) error

func (w WaiterDurationFunc) Wait(ctx context.Context, duration Duration) error {
	return w(ctx, duration)
}

func NewWaiterDuration() WaiterDuration {
	return WaiterDurationFunc(func(ctx context.Context, duration Duration) error {
		if duration <= 0 {
			glog.V(4).Infof("duration <= 0 => skip wait")
			return nil
		}
		glog.V(4).Infof("wait for %v started", duration)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.NewTimer(duration.Duration()).C:
			glog.V(4).Infof("wait for %v completed", duration)
			return nil
		}
	})
}
