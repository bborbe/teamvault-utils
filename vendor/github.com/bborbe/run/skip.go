// Copyright (c) 2020 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package run

import (
	"context"

	"github.com/bborbe/errors"
	"github.com/getsentry/raven-go"
	"github.com/golang/glog"
)

// SkipErrors runs the given Func and returns always nil.
func SkipErrors(fn Func) Func {
	return func(ctx context.Context) error {
		if err := fn(ctx); err != nil {
			glog.Warningf("run failed: %v", err)
		}
		return nil
	}
}

//counterfeiter:generate -o mocks/has-capture-error-and-wait.go --fake-name HasCaptureErrorAndWait . HasCaptureErrorAndWait

// HasCaptureErrorAndWait is compatibel with sentry.Client
type HasCaptureErrorAndWait interface {
	CaptureErrorAndWait(err error, tags map[string]string, interfaces ...raven.Interface) string
}

// SkipErrorsAndReport runs the given Func, report errors to sentry and returns always nil.
func SkipErrorsAndReport(
	fn Func,
	hasCaptureErrorAndWait HasCaptureErrorAndWait,
	tags map[string]string,
) Func {
	return func(ctx context.Context) error {
		if err := fn(ctx); err != nil && errors.Is(err, context.Canceled) == false {
			glog.Warningf("run failed: %v", err)
			hasCaptureErrorAndWait.CaptureErrorAndWait(err, tags)
		}
		return nil
	}
}
