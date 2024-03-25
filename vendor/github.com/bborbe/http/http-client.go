// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"net/http"
	"time"

	"github.com/bborbe/log"
)

func CreateDefaultHttpClient(ctx context.Context, logSamplerFactory log.SamplerFactory) *http.Client {
	return CreateHttpClient(ctx, logSamplerFactory, 30*time.Second)
}

func CreateHttpClient(
	ctx context.Context,
	logSamplerFactory log.SamplerFactory,
	timeout time.Duration,
) *http.Client {
	return &http.Client{
		Transport: CreateDefaultTroundTripper(ctx, logSamplerFactory),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: timeout,
	}
}
