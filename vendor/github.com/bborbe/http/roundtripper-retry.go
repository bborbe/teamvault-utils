// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/bborbe/errors"
	"github.com/golang/glog"
)

const PreventRetryHeaderName = "X-Prevent-Retry"

// NewRoundTripperRetry wraps a given RoundTripper and retry the httpRequest with a delay between.
func NewRoundTripperRetry(
	roundTripper http.RoundTripper,
	retryLimit int,
	retryDelay time.Duration,
) http.RoundTripper {
	return &retryRoundTripper{
		roundTripper: roundTripper,
		retryLimit:   retryLimit,
		retryDelay:   retryDelay,
	}
}

type retryRoundTripper struct {
	roundTripper http.RoundTripper
	retryLimit   int
	retryDelay   time.Duration
}

func (r *retryRoundTripper) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	if req.Header.Get(PreventRetryHeaderName) != "" {
		glog.V(4).Infof("found prevent retry header")
		return r.roundTripper.RoundTrip(req)
	}

	ctx := req.Context()
	retryCounter := 0

	// TODO: implement me
	// limit body reader to x mb
	var body []byte
	if req.Body != nil {
		body, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			reqCloned := req.Clone(ctx)
			if req.Body != nil {
				reqCloned.Body = io.NopCloser(bytes.NewBuffer(body))
			}
			resp, err = r.roundTripper.RoundTrip(reqCloned.WithContext(ctx))
			if err != nil {
				if IsRetryError(err) && retryCounter < r.retryLimit {
					glog.V(1).Infof("%s request to %s failed with error: %v => retry", reqCloned.Method, removeSensibleArgs(reqCloned.URL.String()), err)
					if err := r.delay(ctx); err != nil {
						return nil, errors.Wrapf(ctx, err, "delay failed")
					}
					retryCounter++
					continue
				}
				return nil, errors.Wrapf(ctx, err, "roundtrip failed")
			}

			if !(resp.StatusCode < 400 ||
				resp.StatusCode == 400 ||
				resp.StatusCode == 401 ||
				resp.StatusCode == 404 ||
				r.retryLimit == retryCounter && resp.StatusCode != 502 && resp.StatusCode != 503 && resp.StatusCode != 504) {
				glog.V(1).Infof("%s request to %s failed with status code %d => retry", reqCloned.Method, removeSensibleArgs(reqCloned.URL.String()), resp.StatusCode)
				if err := r.delay(ctx); err != nil {
					return nil, errors.Wrapf(ctx, err, "delay failed")
				}
				retryCounter++
				continue
			}
			return resp, nil
		}
	}
}

func (r *retryRoundTripper) delay(ctx context.Context) error {
	if r.retryDelay > 0 {
		glog.V(3).Infof("sleep for %v", r.retryDelay)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.NewTicker(r.retryDelay).C:
		}
	}
	return nil
}

var removeSensibleArgsRegex = regexp.MustCompile(`hapikey=[^&]+`)

func removeSensibleArgs(value string) string {
	return removeSensibleArgsRegex.ReplaceAllString(value, "hapikey=***")
}
