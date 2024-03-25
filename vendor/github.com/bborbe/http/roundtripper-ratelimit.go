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
	"github.com/golang/glog"
)

func NewRoundTripperRateLimit(
	ctx context.Context,
	tripper http.RoundTripper,
	requestPerSecond int64,
	logSamplerFactory log.SamplerFactory,
) http.RoundTripper {
	logSampler := logSamplerFactory.Sampler()

	var counter int64
	var mux sync.Mutex

	ticker := time.NewTicker(time.Second)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				mux.Lock()
				if counter > 0 {
					counter--
				}
				mux.Unlock()
			}
		}
	}()

	return RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		mux.Lock()
		counter++
		rateLimitedExceeded := counter > requestPerSecond
		sleepAmount := time.Second / time.Duration(requestPerSecond) * time.Duration(counter-requestPerSecond)
		glog.V(3).Infof("counter %d > requestPerSecond %d => rateLimitedExceeded %v", counter, requestPerSecond, rateLimitedExceeded)
		mux.Unlock()

		if sleepAmount > 0 {
			if logSampler.IsSample() {
				glog.V(2).Infof("rate limit exceeded => sleep for %v (sample)", sleepAmount)
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
