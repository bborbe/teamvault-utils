// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"net/http"
	"time"

	libtime "github.com/bborbe/time"
	"github.com/golang/glog"
)

func NewRoundTripperLog(tripper http.RoundTripper) http.RoundTripper {
	return RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		now := libtime.Now()
		resp, err := tripper.RoundTrip(req)
		if err != nil {
			glog.V(2).Infof("%s request to %s in %d ms failed: %v", req.Method, req.URL, time.Since(now).Milliseconds(), err)
			return nil, err
		}
		glog.V(2).Infof("%s request to %s completed with statusCode %d in %d ms", req.Method, req.URL, resp.StatusCode, time.Since(now).Milliseconds())
		return resp, nil
	})
}
