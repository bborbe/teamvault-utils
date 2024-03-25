// Copyright (c) 2020 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package run

import (
	"context"
	"time"

	"github.com/bborbe/errors"
)

var DefaultWaiter = NewWaiter()

// Backoff settings for retry
type Backoff struct {
	// Initial delay to wait on retry
	Delay time.Duration `json:"delay"`
	// Factor initial delay is multipled on retries
	Factor float64 `json:"factor"`
	// Retries how often to retry
	Retries int `json:"retries"`
	// IsRetryAble allow the check if error is retryable
	IsRetryAble func(error) bool `json:"-"`
}

// Retry on error n times and wait between the given delay.
func Retry(backoff Backoff, fn Func) Func {
	return RetryWaiter(backoff, DefaultWaiter, fn)
}

// RetryWaiter allow use of custom Waiter
func RetryWaiter(backoff Backoff, waiter Waiter, fn Func) Func {
	return func(ctx context.Context) error {
		var counter int
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				if err := fn(ctx); err != nil {
					if counter == backoff.Retries {
						return errors.Wrapf(ctx, err, "reached try counter(%d)", backoff.Retries)
					}
					if backoff.IsRetryAble != nil && backoff.IsRetryAble(err) == false {
						return errors.Wrap(ctx, err, "error is not retryable")
					}
					counter++
					if backoff.Delay > 0 {
						delay := backoff.Delay + backoff.Delay*time.Duration(backoff.Factor*float64(counter-1))
						if err := waiter.Wait(ctx, delay); err != nil {
							return errors.Wrapf(ctx, err, "wait %v failed", backoff.Delay)
						}
					}
					continue
				}
				return nil
			}
		}
	}
}
