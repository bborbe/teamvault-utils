// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import (
	"context"
	"time"

	"github.com/golang/glog"
)

//counterfeiter:generate -o mocks/time-waiter.go --fake-name TimeWaiter . Waiter
type Waiter interface {
	WaitUntil(ctx context.Context, until time.Time) error
}

func NewWaiter() Waiter {
	return &waiter{}
}

type waiter struct {
}

func (w *waiter) WaitUntil(ctx context.Context, until time.Time) error {
	now := Now()
	if until.Before(now) {
		glog.V(2).Infof("until already past => skip wait")
		return nil
	}
	glog.V(3).Infof("now: %s wait until: %s", now.Format(time.RFC3339), until.Format(time.RFC3339))
	duration := until.Sub(now) + 10*time.Second
	glog.V(3).Infof("wait for: %v", duration)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.NewTimer(duration).C:
		return nil
	}
}
