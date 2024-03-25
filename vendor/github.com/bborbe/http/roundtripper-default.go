// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/bborbe/log"
)

func CreateDefaultTroundTripper(
	ctx context.Context,
	logSamplerFactory log.SamplerFactory,
) RoundTripper {
	return NewRoundTripperRetry(
		NewRoundTripperRateLimit(
			ctx,
			NewRoundTripperLog(
				&http.Transport{
					Proxy: http.ProxyFromEnvironment,
					DialContext: defaultTransportDialContext(&net.Dialer{
						Timeout:   30 * time.Second,
						KeepAlive: 30 * time.Second,
					}),
					ForceAttemptHTTP2:     true,
					MaxIdleConns:          100,
					IdleConnTimeout:       90 * time.Second,
					TLSHandshakeTimeout:   10 * time.Second,
					ExpectContinueTimeout: 1 * time.Second,
					ResponseHeaderTimeout: 30 * time.Second,
				},
			),
			10,
			logSamplerFactory,
		),
		5,
		time.Second,
	)
}
func defaultTransportDialContext(dialer *net.Dialer) func(context.Context, string, string) (net.Conn, error) {
	return dialer.DialContext
}
