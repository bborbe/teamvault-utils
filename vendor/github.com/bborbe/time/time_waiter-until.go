// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import (
	"context"
	"time"

	"github.com/golang/glog"
)

//counterfeiter:generate -o mocks/waiter-until.go --fake-name WaiterUntil . WaiterUntil
type WaiterUntil interface {
	WaitUntil(ctx context.Context, until DateTime) error
}

type WaiterUntilFunc func(ctx context.Context, until DateTime) error

func (w WaiterUntilFunc) WaitUntil(ctx context.Context, until DateTime) error {
	return w(ctx, until)
}

func NewWaiterUntil(currentDateTime CurrentDateTimeGetter) WaiterUntil {
	waiterDuration := NewWaiterDuration()
	return WaiterUntilFunc(func(ctx context.Context, until DateTime) error {
		now := currentDateTime.Now()
		if until.Before(now) {
			glog.V(4).Infof("until already past => skip wait")
			return nil
		}
		glog.V(4).Infof("now: %s wait until: %s", now.Format(time.RFC3339), until.Format(time.RFC3339))
		duration := until.Sub(now) + 10*Second
		glog.V(4).Infof("wait for: %v", duration)
		return waiterDuration.Wait(ctx, duration)
	})
}
