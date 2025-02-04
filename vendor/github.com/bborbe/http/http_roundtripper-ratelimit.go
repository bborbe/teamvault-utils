// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/bborbe/log"
	libmath "github.com/bborbe/math"
	"github.com/golang/glog"
)

func NewRoundTripperRateLimit(
	ctx context.Context,
	tripper http.RoundTripper,
	maxRequestPerInterval int64,
	intervalDurarion time.Duration,
	logSamplerFactory log.SamplerFactory,
) http.RoundTripper {
	logSampler := logSamplerFactory.Sampler()

	var counter int64
	var mux sync.Mutex

	ticker := time.NewTicker(intervalDurarion)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				mux.Lock()
				counter = *libmath.Max([]int64{counter - maxRequestPerInterval, 0})
				mux.Unlock()
			}
		}
	}()

	return RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		mux.Lock()
		counter++
		rateLimitedExceeded := counter > maxRequestPerInterval
		durationPerRequest := intervalDurarion / time.Duration(maxRequestPerInterval)
		requestOverflowCounter := counter - maxRequestPerInterval
		sleepAmount := durationPerRequest * time.Duration(requestOverflowCounter)
		glog.V(3).Infof("counter %d > requestPerSecond %d => rateLimitedExceeded %v", counter, maxRequestPerInterval, rateLimitedExceeded)
		mux.Unlock()

		if sleepAmount > 0 {
			if logSampler.IsSample() {
				glog.V(2).Infof("rate limit exceeded by (%d) => sleep for %v (sample)", requestOverflowCounter, sleepAmount)
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.NewTimer(sleepAmount).C:

			}
		}

		resp, err := tripper.RoundTrip(req)
		if err != nil {
			glog.V(3).Infof("%s request to %s failed: %w", req.Method, req.URL, err)
			return nil, err
		}
		glog.V(3).Infof("%s request to %s completed with statusCode %d", req.Method, req.URL, resp.StatusCode)
		return resp, nil
	})
}
